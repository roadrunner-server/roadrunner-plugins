package jobs

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/api/jobs"
	"github.com/spiral/roadrunner-plugins/v2/api/jobs/pipeline"
	"github.com/spiral/roadrunner-plugins/v2/config"
	rh "github.com/spiral/roadrunner-plugins/v2/jobs/protocol"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/utils"
	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/pool"
	pq "github.com/spiral/roadrunner/v2/priority_queue"
	"github.com/spiral/roadrunner/v2/state/process"
)

const (
	// RrMode env variable
	RrMode     string = "RR_MODE"
	RrModeJobs string = "jobs"

	PluginName string = "jobs"
	pipelines  string = "pipelines"

	// v2.7 and newer config key
	cfgKey string = "config"
)

type metrics struct {
	jobsOk, pushOk, jobsErr, pushErr *uint64
}

type Plugin struct {
	sync.RWMutex

	// Jobs plugin configuration
	cfg         *Config `structure:"jobs"`
	log         logger.Logger
	workersPool pool.Pool
	server      server.Server

	jobConstructors map[string]jobs.Constructor
	consumers       sync.Map // map[string]jobs.Consumer

	metrics *metrics

	// priority queue implementation
	queue pq.Queue

	// parent config for broken options. keys are pipelines names, values - pointers to the associated pipeline
	pipelines sync.Map

	// initial set of the pipelines to consume
	consume map[string]struct{}

	// signal channel to stop the pollers
	stopCh chan struct{}

	// internal payloads pool
	pldPool       sync.Pool
	statsExporter *statsExporter
	respHandler   *rh.RespHandler
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger, server server.Server) error {
	const op = errors.Op("jobs_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	p.cfg.InitDefaults()

	p.server = server

	p.jobConstructors = make(map[string]jobs.Constructor)
	p.consume = make(map[string]struct{})
	p.stopCh = make(chan struct{}, 1)

	p.pldPool = sync.Pool{New: func() interface{} {
		// with nil fields
		return &payload.Payload{}
	}}

	// initial set of pipelines
	for i := range p.cfg.Pipelines {
		p.pipelines.Store(i, p.cfg.Pipelines[i])
	}

	if len(p.cfg.Consume) > 0 {
		for i := 0; i < len(p.cfg.Consume); i++ {
			p.consume[p.cfg.Consume[i]] = struct{}{}
		}
	}

	// initialize priority queue
	p.queue = pq.NewBinHeap(p.cfg.PipelineSize)
	p.log = log
	p.metrics = &metrics{
		jobsOk:  utils.Uint64(0),
		pushOk:  utils.Uint64(0),
		jobsErr: utils.Uint64(0),
		pushErr: utils.Uint64(0),
	}

	// metrics
	p.statsExporter = newStatsExporter(p, p.metrics.jobsOk, p.metrics.pushOk, p.metrics.jobsErr, p.metrics.pushErr)
	p.respHandler = rh.NewResponseHandler(log)

	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	const op = errors.Op("jobs_plugin_serve")

	// register initial pipelines
	p.pipelines.Range(func(key, value interface{}) bool {
		t := time.Now()
		// pipeline name (ie test-local, sqs-aws, etc)
		name := key.(string)

		// pipeline associated with the name
		pipe := value.(*pipeline.Pipeline)
		// driver for the pipeline (ie amqp, ephemeral, etc)
		dr := pipe.Driver()

		if dr == "" {
			p.log.Warn("can't find driver name for the pipeline, please, check that the 'driver' keyword for the pipelines specified correctly, JOBS plugin will try to run the next pipeline")
			return true
		}

		// jobConstructors contains constructors for the drivers
		// we need here to initialize these drivers for the pipelines
		if _, ok := p.jobConstructors[dr]; ok {
			// v2.7 and newer
			// config key for the particular sub-driver jobs.pipelines.test-local
			configKey := fmt.Sprintf("%s.%s.%s.%s", PluginName, pipelines, name, cfgKey)

			// init the driver
			initializedDriver, err := p.jobConstructors[dr].ConsumerFromConfig(configKey, p.queue)
			if err != nil {
				errCh <- errors.E(op, err)
				return false
			}

			// add driver to the set of the consumers (name - pipeline name, value - associated driver)
			p.consumers.Store(name, initializedDriver)

			// register pipeline for the initialized driver
			err = initializedDriver.Register(context.Background(), pipe)
			if err != nil {
				errCh <- errors.E(op, errors.Errorf("pipe register failed for the driver: %s with pipe name: %s", pipe.Driver(), pipe.Name()))
				return false
			}

			// if pipeline initialized to be consumed, call Run on it
			if _, ok := p.consume[name]; ok {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
				defer cancel()
				err = initializedDriver.Run(ctx, pipe)
				if err != nil {
					errCh <- errors.E(op, err)
					return false
				}
				return true
			}

			return true
		}

		p.log.Debug("driver ready", "pipeline", pipe.Name(), "driver", pipe.Driver(), "start", t, "elapsed", time.Since(t))
		return true
	})

	// do not continue processing, immediately stop if channel contains an error
	if len(errCh) > 0 {
		return errCh
	}

	go func() {
		p.Lock()
		var err error
		p.workersPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: RrModeJobs})
		if err != nil {
			p.Unlock()
			errCh <- err
			return
		}
		p.Unlock()

		// start listening
		p.listener()
	}()

	return errCh
}

func (p *Plugin) Stop() error {
	// this function can block forever, but we don't care, because we might have a chance to exit from the pollers,
	// but if not, this is not a problem at all.
	// The main target is to stop the drivers
	go func() {
		for i := uint8(0); i < p.cfg.NumPollers; i++ {
			// stop jobs plugin pollers
			p.stopCh <- struct{}{}
		}
	}()

	// range over all consumers and call stop
	p.consumers.Range(func(key, value interface{}) bool {
		consumer := value.(jobs.Consumer)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
		err := consumer.Stop(ctx)
		if err != nil {
			cancel()
			p.log.Error("stop job driver", "error", err, "driver", key)
			return true
		}
		cancel()
		return true
	})

	p.Lock()
	if p.workersPool != nil {
		p.workersPool.Destroy(context.Background())
	}
	p.Unlock()

	p.pipelines.Range(func(key, _ interface{}) bool {
		p.pipelines.Delete(key)
		return true
	})

	return nil
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectMQBrokers,
	}
}

func (p *Plugin) CollectMQBrokers(name endure.Named, c jobs.Constructor) {
	p.jobConstructors[name.Name()] = c
}

func (p *Plugin) Workers() []*process.State {
	p.RLock()
	wrk := p.workersPool.Workers()
	p.RUnlock()

	ps := make([]*process.State, len(wrk))

	for i := 0; i < len(wrk); i++ {
		st, err := process.WorkerProcessState(wrk[i])
		if err != nil {
			p.log.Error("jobs workers state", "error", err)
			return nil
		}

		ps[i] = st
	}

	return ps
}

func (p *Plugin) JobsState(ctx context.Context) ([]*jobs.State, error) {
	const op = errors.Op("jobs_plugin_drivers_state")
	jst := make([]*jobs.State, 0, 2)
	var err error
	p.consumers.Range(func(key, value interface{}) bool {
		consumer := value.(jobs.Consumer)
		newCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.cfg.Timeout))

		var state *jobs.State
		state, err = consumer.State(newCtx)
		if err != nil {
			cancel()
			return false
		}

		jst = append(jst, state)
		cancel()
		return true
	})

	if err != nil {
		return nil, errors.E(op, err)
	}
	return jst, nil
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Reset() error {
	p.Lock()
	defer p.Unlock()

	const op = errors.Op("jobs_plugin_reset")
	p.log.Info("JOBS plugin received restart request. Restarting...")
	p.workersPool.Destroy(context.Background())
	p.workersPool = nil

	var err error
	p.workersPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: RrModeJobs})
	if err != nil {
		return errors.E(op, err)
	}

	p.log.Info("JOBS workers pool successfully restarted")

	return nil
}

func (p *Plugin) Push(j *jobs.Job) error {
	const op = errors.Op("jobs_plugin_push")

	start := time.Now()
	// get the pipeline for the job
	pipe, ok := p.pipelines.Load(j.Options.Pipeline)
	if !ok {
		return errors.E(op, errors.Errorf("no such pipeline, requested: %s", j.Options.Pipeline))
	}

	// type conversion
	ppl := pipe.(*pipeline.Pipeline)

	d, ok := p.consumers.Load(ppl.Name())
	if !ok {
		return errors.E(op, errors.Errorf("consumer not registered for the requested driver: %s", ppl.Driver()))
	}

	// if job has no priority, inherit it from the pipeline
	if j.Options.Priority == 0 {
		j.Options.Priority = ppl.Priority()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
	defer cancel()

	err := d.(jobs.Consumer).Push(ctx, j)
	if err != nil {
		atomic.AddUint64(p.metrics.pushErr, 1)
		p.log.Error("job push error", "error", err, "ID", j.Ident, "pipeline", ppl.Name(), "driver", ppl.Driver(), "start", start, "elapsed", time.Since(start))
		return errors.E(op, err)
	}

	atomic.AddUint64(p.metrics.pushOk, 1)
	p.log.Debug("job pushed successfully", "ID", j.Ident, "pipeline", ppl.Name(), "driver", ppl.Driver(), "start", start, "elapsed", time.Since(start))

	return nil
}

func (p *Plugin) PushBatch(j []*jobs.Job) error {
	const op = errors.Op("jobs_plugin_push")
	start := time.Now()

	for i := 0; i < len(j); i++ {
		// get the pipeline for the job
		pipe, ok := p.pipelines.Load(j[i].Options.Pipeline)
		if !ok {
			return errors.E(op, errors.Errorf("no such pipeline, requested: %s", j[i].Options.Pipeline))
		}

		ppl := pipe.(*pipeline.Pipeline)

		d, ok := p.consumers.Load(ppl.Name())
		if !ok {
			return errors.E(op, errors.Errorf("consumer not registered for the requested driver: %s", ppl.Driver()))
		}

		// if job has no priority, inherit it from the pipeline
		if j[i].Options.Priority == 0 {
			j[i].Options.Priority = ppl.Priority()
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
		err := d.(jobs.Consumer).Push(ctx, j[i])
		if err != nil {
			cancel()
			atomic.AddUint64(p.metrics.pushErr, 1)
			p.log.Error("job push batch error", "error", err, "ID", j[i].Ident, "pipeline", ppl.Name(), "driver", ppl.Driver(), "start", start, "elapsed", time.Since(start))
			return errors.E(op, err)
		}

		cancel()
	}

	return nil
}

func (p *Plugin) Pause(pp string) {
	pipe, ok := p.pipelines.Load(pp)

	if !ok {
		p.log.Error("no such pipeline", "requested", pp)
	}

	if pipe == nil {
		p.log.Error("no pipe registered, value is nil")
		return
	}

	ppl := pipe.(*pipeline.Pipeline)

	d, ok := p.consumers.Load(ppl.Name())
	if !ok {
		p.log.Warn("driver for the pipeline not found", "pipeline", pp)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
	defer cancel()
	// redirect call to the underlying driver
	d.(jobs.Consumer).Pause(ctx, ppl.Name())
}

func (p *Plugin) Resume(pp string) {
	pipe, ok := p.pipelines.Load(pp)
	if !ok {
		p.log.Error("no such pipeline", "requested", pp)
	}

	if pipe == nil {
		p.log.Error("no pipe registered, value is nil")
		return
	}

	ppl := pipe.(*pipeline.Pipeline)

	d, ok := p.consumers.Load(ppl.Name())
	if !ok {
		p.log.Warn("driver for the pipeline not found", "pipeline", pp)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
	defer cancel()
	// redirect call to the underlying driver
	d.(jobs.Consumer).Resume(ctx, ppl.Name())
}

// Declare a pipeline.
func (p *Plugin) Declare(pipeline *pipeline.Pipeline) error {
	const op = errors.Op("jobs_plugin_declare")
	// driver for the pipeline (ie amqp, ephemeral, etc)
	dr := pipeline.Driver()
	if dr == "" {
		return errors.E(op, errors.Errorf("no associated driver with the pipeline, pipeline name: %s", pipeline.Name()))
	}

	// jobConstructors contains constructors for the drivers
	// we need here to initialize these drivers for the pipelines
	if _, ok := p.jobConstructors[dr]; ok {
		// init the driver from pipeline
		initializedDriver, err := p.jobConstructors[dr].ConsumerFromPipeline(pipeline, p.queue)
		if err != nil {
			return errors.E(op, err)
		}

		// register pipeline for the initialized driver
		err = initializedDriver.Register(context.Background(), pipeline)
		if err != nil {
			return errors.E(op, errors.Errorf("pipe register failed for the driver: %s with pipe name: %s", pipeline.Driver(), pipeline.Name()))
		}

		// if pipeline initialized to be consumed, call Run on it
		// but likely for the dynamic pipelines it should be started manually
		if _, ok := p.consume[pipeline.Name()]; ok {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
			defer cancel()
			err = initializedDriver.Run(ctx, pipeline)
			if err != nil {
				return errors.E(op, err)
			}
		}

		// add driver to the set of the consumers (name - pipeline name, value - associated driver)
		p.consumers.Store(pipeline.Name(), initializedDriver)
		// save the pipeline
		p.pipelines.Store(pipeline.Name(), pipeline)
	}

	return nil
}

// Destroy pipeline and release all associated resources.
func (p *Plugin) Destroy(pp string) error {
	const op = errors.Op("jobs_plugin_destroy")
	pipe, ok := p.pipelines.Load(pp)
	if !ok {
		return errors.E(op, errors.Errorf("no such pipeline, requested: %s", pp))
	}

	// type conversion
	ppl := pipe.(*pipeline.Pipeline)
	if pipe == nil {
		p.log.Error("no pipe registered, value is nil")
		return errors.Str("no pipe registered, value is nil")
	}

	// delete consumer
	d, ok := p.consumers.LoadAndDelete(ppl.Name())
	if !ok {
		return errors.E(op, errors.Errorf("consumer not registered for the requested driver: %s", ppl.Driver()))
	}

	// delete old pipeline
	p.pipelines.LoadAndDelete(pp)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
	err := d.(jobs.Consumer).Stop(ctx)
	if err != nil {
		cancel()
		return errors.E(op, err)
	}

	cancel()
	return nil
}

func (p *Plugin) List() []string {
	out := make([]string, 0, 10)

	p.pipelines.Range(func(key, _ interface{}) bool {
		// we can safely convert value here as we know that we store keys as strings
		out = append(out, key.(string))
		return true
	})

	return out
}

func (p *Plugin) RPC() interface{} {
	return &rpc{
		log: p.log,
		p:   p,
	}
}

func (p *Plugin) getPayload(body, context []byte) *payload.Payload {
	pld := p.pldPool.Get().(*payload.Payload)
	pld.Body = body
	pld.Context = context
	return pld
}

func (p *Plugin) putPayload(pld *payload.Payload) {
	pld.Body = nil
	pld.Context = nil
	p.pldPool.Put(pld)
}
