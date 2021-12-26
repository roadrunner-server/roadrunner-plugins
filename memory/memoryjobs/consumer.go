package memoryjobs

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/jobs"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/jobs/pipeline"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
	"github.com/spiral/roadrunner/v2/utils"
	"go.uber.org/zap"
)

const (
	prefetch      string = "prefetch"
	goroutinesMax uint64 = 1000
)

type Config struct {
	Priority int64  `mapstructure:"priority"`
	Prefetch uint64 `mapstructure:"prefetch"`
}

type consumer struct {
	cfg           *Config
	log           *zap.Logger
	pipeline      atomic.Value
	pq            priorityqueue.Queue
	localPrefetch chan *Item

	// time.sleep goroutines max number
	goroutines uint64

	delayed *int64
	active  *int64

	priority  int64
	listeners uint32
	stopCh    chan struct{}
}

func FromConfig(configKey string, log *zap.Logger, cfg config.Configurer, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_ephemeral_pipeline")

	jb := &consumer{
		log:        log,
		pq:         pq,
		goroutines: 0,
		active:     utils.Int64(0),
		delayed:    utils.Int64(0),
		stopCh:     make(chan struct{}),
	}

	err := cfg.UnmarshalKey(configKey, &jb.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if jb.cfg == nil {
		return nil, errors.E(op, errors.Errorf("config not found by provided key: %s", configKey))
	}

	if jb.cfg.Prefetch == 0 {
		jb.cfg.Prefetch = 100_000
	}

	if jb.cfg.Priority == 0 {
		jb.cfg.Priority = 10
	}

	jb.priority = jb.cfg.Priority

	// initialize a local queue
	jb.localPrefetch = make(chan *Item, jb.cfg.Prefetch)

	return jb, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log *zap.Logger, pq priorityqueue.Queue) (*consumer, error) {
	return &consumer{
		log:           log,
		pq:            pq,
		localPrefetch: make(chan *Item, pipeline.Int(prefetch, 100_000)),
		goroutines:    0,
		active:        utils.Int64(0),
		delayed:       utils.Int64(0),
		priority:      pipeline.Priority(),
		stopCh:        make(chan struct{}),
	}, nil
}

func (c *consumer) Push(ctx context.Context, jb *jobs.Job) error {
	const op = errors.Op("ephemeral_push")
	// check if the pipeline registered
	_, ok := c.pipeline.Load().(*pipeline.Pipeline)
	if !ok {
		return errors.E(op, errors.Errorf("no such pipeline: %s", jb.Options.Pipeline))
	}

	err := c.handleItem(ctx, fromJob(jb))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (c *consumer) State(_ context.Context) (*jobs.State, error) {
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	return &jobs.State{
		Pipeline: pipe.Name(),
		Driver:   pipe.Driver(),
		Queue:    pipe.Name(),
		Active:   atomic.LoadInt64(c.active),
		Delayed:  atomic.LoadInt64(c.delayed),
		Ready:    ready(atomic.LoadUint32(&c.listeners)),
	}, nil
}

func (c *consumer) Register(_ context.Context, pipeline *pipeline.Pipeline) error {
	c.pipeline.Store(pipeline)
	return nil
}

func (c *consumer) Run(_ context.Context, pipe *pipeline.Pipeline) error {
	const op = errors.Op("memory_jobs_run")
	t := time.Now()

	l := atomic.LoadUint32(&c.listeners)
	// listener already active
	if l == 1 {
		c.log.Warn("listener already in the active state")
		return errors.E(op, errors.Str("listener already in the active state"))
	}

	c.consume()
	atomic.StoreUint32(&c.listeners, 1)

	c.log.Debug("pipeline was started", zap.String("driver", pipe.Driver()), zap.String("pipeline", pipe.Name()), zap.String("start", time.Now().String()), zap.String("elapsed", time.Since(t).String()))
	return nil
}

func (c *consumer) Pause(_ context.Context, p string) {
	start := time.Now()
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", zap.String("pause was requested: ", p))
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 0 {
		c.log.Warn("no active listeners, nothing to pause")
		return
	}

	atomic.AddUint32(&c.listeners, ^uint32(0))

	// stop the consumer
	c.stopCh <- struct{}{}

	c.log.Debug("pipeline was paused", zap.String("driver", pipe.Driver()), zap.String("pipeline", pipe.Name()), zap.String("start", time.Now().String()), zap.String("elapsed", time.Since(start).String()))
}

func (c *consumer) Resume(_ context.Context, p string) {
	start := time.Now()
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", zap.String("resume was requested: ", p))
	}

	l := atomic.LoadUint32(&c.listeners)
	// listener already active
	if l == 1 {
		c.log.Warn("listener already in the active state")
		return
	}

	// resume the consumer on the same channel
	c.consume()

	atomic.StoreUint32(&c.listeners, 1)
	c.log.Debug("pipeline was resumed", zap.String("driver", pipe.Driver()), zap.String("pipeline", pipe.Name()), zap.String("start", time.Now().String()), zap.String("elapsed", time.Since(start).String()))
}

func (c *consumer) Stop(_ context.Context) error {
	start := time.Now()
	pipe := c.pipeline.Load().(*pipeline.Pipeline)

	select {
	case c.stopCh <- struct{}{}:
	default:
		break
	}

	for i := 0; i < len(c.localPrefetch); i++ {
		// drain all jobs from the channel
		<-c.localPrefetch
	}

	c.localPrefetch = nil

	c.log.Debug("pipeline was stopped", zap.String("driver", pipe.Driver()), zap.String("pipeline", pipe.Name()), zap.String("start", time.Now().String()), zap.String("elapsed", time.Since(start).String()))
	return nil
}

func (c *consumer) handleItem(ctx context.Context, msg *Item) error {
	const op = errors.Op("ephemeral_handle_request")
	// handle timeouts
	// theoretically, some bad user may send millions requests with a delay and produce a billion (for example)
	// goroutines here. We should limit goroutines here.
	if msg.Options.Delay > 0 {
		// if we have 1000 goroutines waiting on the delay - reject 1001
		if atomic.LoadUint64(&c.goroutines) >= goroutinesMax {
			return errors.E(op, errors.Str("max concurrency number reached"))
		}

		go func(jj *Item) {
			atomic.AddUint64(&c.goroutines, 1)
			atomic.AddInt64(c.delayed, 1)

			time.Sleep(jj.Options.DelayDuration())

			select {
			case c.localPrefetch <- jj:
				atomic.AddUint64(&c.goroutines, ^uint64(0))
			default:
				c.log.Warn("can't push job", zap.String("error", "local queue closed or full"))
			}
		}(msg)

		return nil
	}

	// increase number of the active jobs
	atomic.AddInt64(c.active, 1)

	// insert to the local, limited pipeline
	select {
	case c.localPrefetch <- msg:
		return nil
	case <-ctx.Done():
		return errors.E(op, errors.Errorf("local pipeline is full, consider to increase prefetch number, current limit: %d, context error: %v", c.cfg.Prefetch, ctx.Err()))
	}
}

func (c *consumer) consume() {
	go func() {
		// redirect
		for {
			select {
			case item, ok := <-c.localPrefetch:
				if !ok {
					c.log.Debug("ephemeral local prefetch queue closed")
					return
				}

				if item.Priority() == 0 {
					item.Options.Priority = c.priority
				}
				// set requeue channel
				item.Options.requeueFn = c.handleItem
				item.Options.active = c.active
				item.Options.delayed = c.delayed

				c.pq.Insert(item)
			case <-c.stopCh:
				return
			}
		}
	}()
}

func ready(r uint32) bool {
	return r > 0
}
