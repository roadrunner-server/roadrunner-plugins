package natsjobs

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	json "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"github.com/spiral/errors"
	cfgPlugin "github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/jobs/job"
	"github.com/spiral/roadrunner-plugins/v2/jobs/pipeline"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/events"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
	jobState "github.com/spiral/roadrunner/v2/state/job"
)

const (
	pluginName      string = "nats"
	reconnectBuffer int    = 20 * 1024 * 1024
)

type consumer struct {
	// system
	sync.RWMutex
	log       logger.Logger
	eh        events.Handler
	conn      *nats.Conn
	queue     priorityqueue.Queue
	prefetch  int
	listeners uint32
	pipeline  atomic.Value
	sub       *nats.Subscription
	msgCh     chan *nats.Msg
	stopCh    chan struct{}

	// config
	url   string
	natsQ string
}

func FromConfig(configKey string, e events.Handler, log logger.Logger, cfg cfgPlugin.Configurer, queue priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_nats_consumer")

	if !cfg.Has(configKey) {
		return nil, errors.E(op, errors.Errorf("no configuration by provided key: %s", configKey))
	}

	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global nats configuration, global configuration should contain NATS URL"))
	}

	var conf *config
	err := cfg.UnmarshalKey(configKey, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	err = cfg.UnmarshalKey(pluginName, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	conf.InitDefaults()

	conn, err := nats.Connect(conf.Addr,
		nats.NoEcho(),
		nats.Timeout(time.Minute),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second*2),
		nats.ReconnectBufSize(reconnectBuffer),
		nats.ReconnectHandler(reconnectHandler(log)),
		nats.DisconnectErrHandler(disconnectHandler(log)),
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cs := &consumer{
		log:      log,
		eh:       e,
		conn:     conn,
		natsQ:    conf.Queue,
		queue:    queue,
		url:      conf.Addr,
		prefetch: conf.Prefetch,
		msgCh:    make(chan *nats.Msg, conf.Prefetch),
		stopCh:   make(chan struct{}),
	}

	return cs, nil
}

func FromPipeline(pipe *pipeline.Pipeline, e events.Handler, log logger.Logger, cfg cfgPlugin.Configurer, queue priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_nats_consumer")

	// if no global section -- error
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global nats configuration, global configuration should contain NATS URL"))
	}

	var conf *config
	err := cfg.UnmarshalKey(pluginName, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	conf.InitDefaults()

	conn, err := nats.Connect(conf.Addr,
		//nats.NoEcho(),
		nats.Timeout(time.Minute),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second*2),
		nats.ReconnectBufSize(reconnectBuffer),
		nats.ReconnectHandler(reconnectHandler(log)),
		nats.DisconnectErrHandler(disconnectHandler(log)),
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cs := &consumer{
		log:      log,
		eh:       e,
		conn:     conn,
		natsQ:    pipe.String("queue", "default"),
		queue:    queue,
		prefetch: pipe.Int("prefetch", 100),
		msgCh:    make(chan *nats.Msg, pipe.Int("prefetch", 100)),
		stopCh:   make(chan struct{}),
	}

	return cs, nil
}

func (c *consumer) Push(_ context.Context, job *job.Job) error {
	const op = errors.Op("nats_consumer_push")
	if job.Options.Delay > 0 {
		return errors.E(op, errors.Str("nats doesn't support delayed messages, see: https://github.com/nats-io/nats-streaming-server/issues/324"))
	}

	data, err := json.Marshal(job)
	if err != nil {
		return errors.E(op, err)
	}

	err = c.conn.Publish(c.natsQ, data)
	if err != nil {
		return errors.E(op, err)
	}

	job = nil
	return nil
}

func (c *consumer) Register(_ context.Context, pipeline *pipeline.Pipeline) error {
	c.pipeline.Store(pipeline)
	return nil
}

func (c *consumer) Run(ctx context.Context, p *pipeline.Pipeline) error {
	start := time.Now()
	const op = errors.Op("nats_run")

	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p.Name() {
		return errors.E(op, errors.Errorf("no such pipeline registered: %s", pipe.Name()))
	}

	err := c.listenerInit()
	if err != nil {
		return errors.E(op, err)
	}

	atomic.AddUint32(&c.listeners, 1)

	go c.listenerStart()

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    start,
		Elapsed:  time.Since(start),
	})

	return nil
}

func (c *consumer) Stop(_ context.Context) error {
	start := time.Now()

	if atomic.LoadUint32(&c.listeners) > 0 {
		if c.sub != nil {
			_ = c.sub.Drain()
			_ = c.sub.Unsubscribe()
		}
		c.stopCh <- struct{}{}
	}

	pipe := c.pipeline.Load().(*pipeline.Pipeline)

	c.conn.Close()

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeStopped,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    start,
		Elapsed:  time.Since(start),
	})

	return nil
}

func (c *consumer) Pause(_ context.Context, p string) {
	start := time.Now()

	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", "requested pause on: ", p)
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 0 {
		c.log.Warn("no active listeners, nothing to pause")
		return
	}

	atomic.AddUint32(&c.listeners, ^uint32(0))

	if c.sub != nil {
		_ = c.sub.Drain()
		_ = c.sub.Unsubscribe()
	}

	c.stopCh <- struct{}{}
	c.sub = nil

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipePaused,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    start,
		Elapsed:  time.Since(start),
	})
}

func (c *consumer) Resume(_ context.Context, p string) {
	start := time.Now()
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", "requested resume on: ", p)
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 1 {
		c.log.Warn("nats listener already in the active state")
		return
	}

	err := c.listenerInit()
	if err != nil {
		c.log.Error("failed to resume NATS pipeline", "error", err, "pipeline", pipe.Name())
		return
	}

	go c.listenerStart()

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    start,
		Elapsed:  time.Since(start),
	})
}

func (c *consumer) State(_ context.Context) (*jobState.State, error) {
	if c.sub == nil {
		return nil, errors.Str("can't get state, no available subscribers")
	}

	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	res, _, err := c.sub.Pending()
	if err != nil {
		return nil, err
	}

	st := &jobState.State{
		Pipeline: pipe.Name(),
		Driver:   pipe.Driver(),
		Queue:    c.natsQ,
		Active:   0,
		Delayed:  0,
		Reserved: int64(res),
		Ready:    ready(atomic.LoadUint32(&c.listeners)),
	}

	return st, nil
}

// private

func (c *consumer) requeue(item *Item) error {
	const op = errors.Op("nats_requeue")
	if item.Options.Delay > 0 {
		return errors.E(op, errors.Str("nats doesn't support delayed messages, see: https://github.com/nats-io/nats-streaming-server/issues/324"))
	}

	data, err := json.Marshal(item)
	if err != nil {
		return errors.E(op, err)
	}

	err = c.conn.Publish(c.natsQ, data)
	if err != nil {
		return errors.E(op, err)
	}

	item = nil
	return nil
}

func reconnectHandler(log logger.Logger) func(*nats.Conn) {
	return func(_ *nats.Conn) {
		log.Warn("connection lost, reconnecting...")
	}
}

func disconnectHandler(log logger.Logger) func(*nats.Conn, error) {
	return func(_ *nats.Conn, err error) {
		if err != nil {
			log.Error("nast disconnected", "error", err)
			return
		}

		log.Warn("nast disconnected, trying to reconnect...")
	}
}

func ready(r uint32) bool {
	return r > 0
}
