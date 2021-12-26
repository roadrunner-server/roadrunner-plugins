package informer

import (
	"context"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/informer"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/jobs"
	"github.com/spiral/roadrunner/v2/state/process"
)

const PluginName = "informer"

type Plugin struct {
	withJobs    map[string]informer.JobsStat
	withWorkers map[string]informer.Informer
}

func (p *Plugin) Init() error {
	p.withWorkers = make(map[string]informer.Informer)
	p.withJobs = make(map[string]informer.JobsStat)
	return nil
}

// Workers provides BaseProcess slice with workers for the requested plugin
func (p *Plugin) Workers(name string) []*process.State {
	svc, ok := p.withWorkers[name]
	if !ok {
		return nil
	}

	return svc.Workers()
}

// Jobs provides information about jobs for the registered plugin using jobs
func (p *Plugin) Jobs(name string) []*jobs.State {
	svc, ok := p.withJobs[name]
	if !ok {
		return nil
	}

	st, err := svc.JobsState(context.Background())
	if err != nil {
		// skip errors here
		return nil
	}

	return st
}

// Collects declares services to be collected.
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectWorkers,
		p.CollectJobs,
	}
}

// CollectWorkers obtains plugins with workers inside.
func (p *Plugin) CollectWorkers(name endure.Named, r informer.Informer) {
	p.withWorkers[name.Name()] = r
}

func (p *Plugin) CollectJobs(name endure.Named, j informer.JobsStat) {
	p.withJobs[name.Name()] = j
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p}
}
