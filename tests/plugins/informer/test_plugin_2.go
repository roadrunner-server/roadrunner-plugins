package informer

import (
	"context"
	"sync"
	"time"

	"github.com/roadrunner-server/api/plugins/v2/config"
	"github.com/roadrunner-server/api/plugins/v2/server"
	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/state/process"
)

var testPoolConfig2 = &pool.Config{
	NumWorkers:      10,
	MaxJobs:         100,
	AllocateTimeout: time.Second * 10,
	DestroyTimeout:  time.Second * 10,
	Supervisor: &pool.SupervisorConfig{
		WatchTick:       60 * time.Second,
		TTL:             1000 * time.Second,
		IdleTTL:         10 * time.Second,
		ExecTTL:         10 * time.Second,
		MaxWorkerMemory: 1000,
	},
}

// Gauge //////////////
type Plugin2 struct {
	sync.Mutex

	config config.Configurer
	server server.Server

	pool pool.Pool
}

func (p2 *Plugin2) Init(cfg config.Configurer, server server.Server) error {
	p2.config = cfg
	p2.server = server
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Second * 5)
		p2.Lock()
		defer p2.Unlock()
		var err error
		p2.pool, err = p2.server.NewWorkerPool(context.Background(), testPoolConfig, nil)
		if err != nil {
			panic(err)
		}
	}()
	return errCh
}

func (p2 *Plugin2) Stop() error {
	return nil
}

func (p2 *Plugin2) Name() string {
	return "informer.plugin2"
}

func (p2 *Plugin2) Workers() []*process.State {
	if p2.pool == nil {
		return nil
	}
	ps := make([]*process.State, 0, len(p2.pool.Workers()))
	workers := p2.pool.Workers()
	for i := 0; i < len(workers); i++ {
		state, err := process.WorkerProcessState(workers[i])
		if err != nil {
			return nil
		}
		ps = append(ps, state)
	}

	return ps
}
