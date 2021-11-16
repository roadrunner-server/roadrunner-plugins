package grpc

import (
	"context"
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/grpc/codec"
	"github.com/spiral/roadrunner-plugins/v2/grpc/proxy"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/state/process"
	"github.com/spiral/roadrunner/v2/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

const (
	name   string = "grpc"
	RrGrpc string = "RR_GRPC"
)

type Plugin struct {
	mu        *sync.RWMutex
	config    *Config
	gPool     pool.Pool
	opts      []grpc.ServerOption
	services  []func(server *grpc.Server)
	server    *grpc.Server
	rrServer  server.Server
	proxyList []*proxy.Proxy

	log logger.Logger
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger, server server.Server) error {
	const op = errors.Op("grpc_plugin_init")

	if !cfg.Has(name) {
		return errors.E(errors.Disabled)
	}
	// register the codec
	encoding.RegisterCodec(&codec.Codec{
		Base: encoding.GetCodec(codec.Name),
	})

	err := cfg.UnmarshalKey(name, &p.config)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.config.InitDefaults()
	if err != nil {
		return errors.E(op, err)
	}

	p.opts = make([]grpc.ServerOption, 0)
	p.services = make([]func(server *grpc.Server), 0)
	p.rrServer = server
	p.proxyList = make([]*proxy.Proxy, 0, 1)

	// worker's GRPC mode
	if p.config.Env == nil {
		p.config.Env = make(map[string]string)
	}
	p.config.Env[RrGrpc] = "true"

	p.log = log
	p.mu = &sync.RWMutex{}

	return nil
}

func (p *Plugin) Serve() chan error {
	const op = errors.Op("grpc_plugin_serve")
	errCh := make(chan error, 1)

	var err error
	p.gPool, err = p.rrServer.NewWorkerPool(context.Background(), &pool.Config{
		Debug:           p.config.GrpcPool.Debug,
		NumWorkers:      p.config.GrpcPool.NumWorkers,
		MaxJobs:         p.config.GrpcPool.MaxJobs,
		AllocateTimeout: p.config.GrpcPool.AllocateTimeout,
		DestroyTimeout:  p.config.GrpcPool.DestroyTimeout,
		Supervisor:      p.config.GrpcPool.Supervisor,
	}, p.config.Env)
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	p.server, err = p.createGRPCserver()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	l, err := utils.CreateListener(p.config.Listen)
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	go func() {
		p.log.Info("GRPC server started", "address", p.config.Listen)
		err = p.server.Serve(l)
		if err != nil {
			// skip errors when stopping the server
			if err == grpc.ErrServerStopped {
				return
			}

			p.log.Error("grpc server stopped", "error", err)
			errCh <- errors.E(op, err)
			return
		}
	}()

	return errCh
}

func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.server != nil {
		p.server.Stop()
	}
	return nil
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return name
}

func (p *Plugin) Reset() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	const op = errors.Op("grpc_plugin_reset")

	// destroy old pool
	p.gPool.Destroy(context.Background())

	var err error
	p.gPool, err = p.rrServer.NewWorkerPool(context.Background(), &pool.Config{
		Debug:           p.config.GrpcPool.Debug,
		NumWorkers:      p.config.GrpcPool.NumWorkers,
		MaxJobs:         p.config.GrpcPool.MaxJobs,
		AllocateTimeout: p.config.GrpcPool.AllocateTimeout,
		DestroyTimeout:  p.config.GrpcPool.DestroyTimeout,
		Supervisor:      p.config.GrpcPool.Supervisor,
	}, p.config.Env)
	if err != nil {
		return errors.E(op, err)
	}

	// update pointers to the pool
	for i := 0; i < len(p.proxyList); i++ {
		p.proxyList[i].UpdatePool(p.gPool)
	}

	return nil
}

func (p *Plugin) Workers() []*process.State {
	p.mu.RLock()
	defer p.mu.RUnlock()

	workers := p.gPool.Workers()

	ps := make([]*process.State, 0, len(workers))
	for i := 0; i < len(workers); i++ {
		state, err := process.WorkerProcessState(workers[i])
		if err != nil {
			return nil
		}
		ps = append(ps, state)
	}

	return ps
}
