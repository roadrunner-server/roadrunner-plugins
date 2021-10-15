package beanstalk

import (
	"bytes"
	"context"
	"encoding/gob"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/spiral/errors"
	cfgPlugin "github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/jobs/job"
	"github.com/spiral/roadrunner-plugins/v2/jobs/pipeline"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/events"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
	jobState "github.com/spiral/roadrunner/v2/state/job"
	"github.com/spiral/roadrunner/v2/utils"
)

type consumer struct {
	log logger.Logger
	eh  events.Handler
	pq  priorityqueue.Queue

	pipeline  atomic.Value
	listeners uint32

	// beanstalk
	pool           *ConnPool
	addr           string
	network        string
	reserveTimeout time.Duration
	reconnectCh    chan struct{}
	tout           time.Duration
	// tube name
	tName        string
	tubePriority *uint32
	priority     int64

	stopCh    chan struct{}
	requeueCh chan *Item
}

func NewBeanstalkConsumer(configKey string, log logger.Logger, cfg cfgPlugin.Configurer, e events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_beanstalk_consumer")

	// PARSE CONFIGURATION -------
	var conf config
	if !cfg.Has(configKey) {
		return nil, errors.E(op, errors.Errorf("no configuration by provided key: %s", configKey))
	}

	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global beanstalk configuration, global configuration should contain beanstalk addrs and timeout"))
	}

	err := cfg.UnmarshalKey(configKey, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	err = cfg.UnmarshalKey(pluginName, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	conf.InitDefault()

	// PARSE CONFIGURATION -------

	dsn := strings.Split(conf.Addr, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Errorf("invalid socket DSN (tcp://127.0.0.1:11300, unix://beanstalk.sock), provided: %s", conf.Addr))
	}

	cPool, err := NewConnPool(dsn[0], dsn[1], conf.Tube, conf.Timeout, log)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// initialize job consumer
	jc := &consumer{
		pq:             pq,
		log:            log,
		eh:             e,
		pool:           cPool,
		network:        dsn[0],
		addr:           dsn[1],
		tout:           conf.Timeout,
		tName:          conf.Tube,
		reserveTimeout: conf.ReserveTimeout,
		tubePriority:   conf.TubePriority,
		priority:       conf.PipePriority,

		// buffered with two because jobs root plugin can call Stop at the same time as Pause
		stopCh:      make(chan struct{}, 2),
		requeueCh:   make(chan *Item, 1000),
		reconnectCh: make(chan struct{}, 2),
	}

	return jc, nil
}

func FromPipeline(pipe *pipeline.Pipeline, log logger.Logger, cfg cfgPlugin.Configurer, e events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_beanstalk_consumer")

	// PARSE CONFIGURATION -------
	var conf config
	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global beanstalk configuration, global configuration should contain beanstalk addrs and timeout"))
	}

	err := cfg.UnmarshalKey(pluginName, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	conf.InitDefault()
	// PARSE CONFIGURATION -------

	dsn := strings.Split(conf.Addr, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Errorf("invalid socket DSN (tcp://127.0.0.1:11300, unix://beanstalk.sock), provided: %s", conf.Addr))
	}

	cPool, err := NewConnPool(dsn[0], dsn[1], pipe.String(tube, "default"), conf.Timeout, log)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// initialize job consumer
	jc := &consumer{
		pq:             pq,
		log:            log,
		eh:             e,
		pool:           cPool,
		network:        dsn[0],
		addr:           dsn[1],
		tout:           conf.Timeout,
		tName:          pipe.String(tube, "default"),
		reserveTimeout: time.Second * time.Duration(pipe.Int(reserveTimeout, 5)),
		tubePriority:   utils.Uint32(uint32(pipe.Int(tubePriority, 1))),
		priority:       pipe.Priority(),

		// buffered with two because jobs root plugin can call Stop at the same time as Pause
		stopCh:      make(chan struct{}, 2),
		requeueCh:   make(chan *Item, 1000),
		reconnectCh: make(chan struct{}, 2),
	}

	return jc, nil
}
func (c *consumer) Push(ctx context.Context, jb *job.Job) error {
	const op = errors.Op("beanstalk_push")
	// check if the pipeline registered

	// load atomic value
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != jb.Options.Pipeline {
		return errors.E(op, errors.Errorf("no such pipeline: %s, actual: %s", jb.Options.Pipeline, pipe.Name()))
	}

	err := c.handleItem(ctx, fromJob(jb))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (c *consumer) Register(_ context.Context, p *pipeline.Pipeline) error {
	// register the pipeline
	c.pipeline.Store(p)
	return nil
}

// State https://github.com/beanstalkd/beanstalkd/blob/master/doc/protocol.txt#L514
func (c *consumer) State(ctx context.Context) (*jobState.State, error) {
	const op = errors.Op("beanstalk_state")
	stat, err := c.pool.Stats(ctx)
	if err != nil {
		return nil, errors.E(op, err)
	}

	pipe := c.pipeline.Load().(*pipeline.Pipeline)

	out := &jobState.State{
		Pipeline: pipe.Name(),
		Driver:   pipe.Driver(),
		Queue:    c.tName,
		Ready:    ready(atomic.LoadUint32(&c.listeners)),
	}

	// set stat, skip errors (replace with 0)
	// https://github.com/beanstalkd/beanstalkd/blob/master/doc/protocol.txt#L523
	if v, err := strconv.Atoi(stat["current-jobs-ready"]); err == nil {
		out.Active = int64(v)
	}

	// https://github.com/beanstalkd/beanstalkd/blob/master/doc/protocol.txt#L525
	if v, err := strconv.Atoi(stat["current-jobs-reserved"]); err == nil {
		// this is not an error, reserved in beanstalk behaves like an active jobs
		out.Reserved = int64(v)
	}

	// https://github.com/beanstalkd/beanstalkd/blob/master/doc/protocol.txt#L528
	if v, err := strconv.Atoi(stat["current-jobs-delayed"]); err == nil {
		out.Delayed = int64(v)
	}

	return out, nil
}

func (c *consumer) Run(_ context.Context, p *pipeline.Pipeline) error {
	const op = errors.Op("beanstalk_run")
	start := time.Now()

	// load atomic value
	// check if the pipeline registered
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p.Name() {
		return errors.E(op, errors.Errorf("no such pipeline: %s, actual: %s", p.Name(), pipe.Name()))
	}

	atomic.AddUint32(&c.listeners, 1)

	go c.listen()

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    start,
		Elapsed:  time.Since(start),
	})

	return nil
}

func (c *consumer) Stop(context.Context) error {
	start := time.Now()
	pipe := c.pipeline.Load().(*pipeline.Pipeline)

	if atomic.LoadUint32(&c.listeners) == 1 {
		c.stopCh <- struct{}{}
	}

	// release associated resources
	c.pool.Stop()

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
	// load atomic value
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", "requested", p, "actual", pipe.Name())
		return
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 0 {
		c.log.Warn("no active listeners, nothing to pause")
		return
	}

	atomic.AddUint32(&c.listeners, ^uint32(0))

	c.stopCh <- struct{}{}

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
	// load atomic value
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", "requested", p, "actual", pipe.Name())
		return
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 1 {
		c.log.Warn("sqs listener already in the active state")
		return
	}

	// start listener
	go c.listen()

	// increase num of listeners
	atomic.AddUint32(&c.listeners, 1)

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    start,
		Elapsed:  time.Since(start),
	})
}

func (c *consumer) handleItem(ctx context.Context, item *Item) error {
	const op = errors.Op("beanstalk_handle_item")

	bb := new(bytes.Buffer)
	bb.Grow(64)
	err := gob.NewEncoder(bb).Encode(item)
	if err != nil {
		return errors.E(op, err)
	}

	body := make([]byte, bb.Len())
	copy(body, bb.Bytes())
	bb.Reset()
	bb = nil

	// https://github.com/beanstalkd/beanstalkd/blob/master/doc/protocol.txt#L458
	// <pri> is an integer < 2**32. Jobs with smaller priority values will be
	// scheduled before jobs with larger priorities. The most urgent priority is 0;
	// the least urgent priority is 4,294,967,295.
	//
	// <delay> is an integer number of seconds to wait before putting the job in
	// the ready queue. The job will be in the "delayed" state during this time.
	// Maximum delay is 2**32-1.
	//
	// <ttr> -- time to run -- is an integer number of seconds to allow a worker
	// to run this job. This time is counted from the moment a worker reserves
	// this job. If the worker does not delete, release, or bury the job within
	// <ttr> seconds, the job will time out and the server will release the job.
	//	The minimum ttr is 1. If the client sends 0, the server will silently
	// increase the ttr to 1. Maximum ttr is 2**32-1.
	id, err := c.pool.Put(ctx, body, *c.tubePriority, item.Options.DelayDuration(), c.tout)
	if err != nil {
		errD := c.pool.Delete(ctx, id)
		if errD != nil {
			return errors.E(op, errors.Errorf("%s:%s", err.Error(), errD.Error()))
		}
		return errors.E(op, err)
	}

	return nil
}

func (c *consumer) handleTPush(data []byte, tube string) error {
	// todo(rustatian): hash c.addr+c.network, store, and then get from the map and put data
	// it's tooo big overhead to connect every time
	connT, err := beanstalk.DialTimeout(c.network, c.addr, c.tout)
	if err != nil {
		return err
	}

	t := beanstalk.NewTube(connT, tube)
	_, err = t.Put(data, 0, 0, c.tout)
	if err != nil {
		return err
	}

	return connT.Close()
}

func ready(r uint32) bool {
	return r > 0
}
