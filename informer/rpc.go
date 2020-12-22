package informer

import (
	"github.com/spiral/roadrunner-plugins/logger"
	"github.com/spiral/roadrunner/v2/interfaces/worker"
)

type rpc struct {
	srv *Plugin
	log logger.Logger
}

// WorkerList contains list of workers.
type WorkerList struct {
	// Workers is list of workers.
	Workers []ProcessState `json:"workers"`
}

// List all resettable services.
func (rpc *rpc) List(_ bool, list *[]string) error {
	rpc.log.Debug("Started List method")
	*list = make([]string, 0, len(rpc.srv.registry))

	for name := range rpc.srv.registry {
		*list = append(*list, name)
	}
	rpc.log.Debug("list of services", "list", *list)

	rpc.log.Debug("successfully finished List method")
	return nil
}

// Workers state of a given service.
func (rpc *rpc) Workers(service string, list *WorkerList) error {
	rpc.log.Debug("started Workers method", "service", service)
	workers, err := rpc.srv.Workers(service)
	if err != nil {
		return err
	}

	list.Workers = make([]ProcessState, 0)
	for _, w := range workers {
		ps, err := WorkerProcessState(w.(worker.BaseProcess))
		if err != nil {
			continue
		}

		list.Workers = append(list.Workers, ps)
	}
	rpc.log.Debug("list of workers", "workers", list.Workers)
	rpc.log.Debug("successfully finished Workers method")
	return nil
}
