package resetter

import (
	"context"
	"time"

	"github.com/roadrunner-server/api/plugins/v2/config"
	"github.com/roadrunner-server/api/plugins/v2/server"
	poolImpl "github.com/spiral/roadrunner/v2/pool"
)

var testPoolConfig = &poolImpl.Config{
	NumWorkers:      10,
	MaxJobs:         100,
	AllocateTimeout: time.Second * 10,
	DestroyTimeout:  time.Second * 10,
	Supervisor: &poolImpl.SupervisorConfig{
		WatchTick:       60 * time.Second,
		TTL:             1000 * time.Second,
		IdleTTL:         10 * time.Second,
		ExecTTL:         10 * time.Second,
		MaxWorkerMemory: 1000,
	},
}

// Gauge //////////////
type Plugin1 struct {
	config config.Configurer
	server server.Server

	p poolImpl.Pool
}

func (p1 *Plugin1) Init(cfg config.Configurer, server server.Server) error {
	p1.config = cfg
	p1.server = server
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	var err error
	p1.p, err = p1.server.NewWorkerPool(context.Background(), testPoolConfig, nil)
	if err != nil {
		errCh <- err
		return errCh
	}
	return errCh
}

func (p1 *Plugin1) Stop() error {
	return nil
}

func (p1 *Plugin1) Name() string {
	return "resetter.plugin1"
}

func (p1 *Plugin1) Reset() error {
	for i := 0; i < 10; i++ {
		err := p1.p.Reset(context.Background())
		if err != nil {
			return err
		}
	}

	return nil
}
