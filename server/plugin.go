package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"

	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/transport"
	"github.com/spiral/roadrunner/v2/transport/pipe"
	"github.com/spiral/roadrunner/v2/transport/socket"
	"github.com/spiral/roadrunner/v2/utils"
	"github.com/spiral/roadrunner/v2/worker"
)

const (
	// PluginName for the server
	PluginName = "server"
	// RrRelay env variable key (internal)
	RrRelay = "RR_RELAY"
	// RrRPC env variable key (internal) if the RPC presents
	RrRPC = "RR_RPC"
)

// Plugin manages worker
type Plugin struct {
	cfg     Config
	log     logger.Logger
	factory transport.Factory
}

// Init application provider.
func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("server_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}
	err := cfg.Unmarshal(&p.cfg)
	if err != nil {
		return errors.E(op, errors.Init, err)
	}
	p.cfg.InitDefaults()
	p.log = log

	p.factory, err = p.initFactory()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

// Name contains service name.
func (p *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (p *Plugin) Available() {}

// Serve (Start) server plugin (just a mock here to satisfy interface)
func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

// Stop used to close chosen in config factory
func (p *Plugin) Stop() error {
	if p.factory == nil {
		return nil
	}

	return p.factory.Close()
}

// CmdFactory provides worker command factory associated with given context.
func (p *Plugin) CmdFactory(env Env) func() *exec.Cmd {
	return func() *exec.Cmd {
		var cmdArgs []string
		// create command according to the config
		cmdArgs = append(cmdArgs, strings.Split(p.cfg.Server.Command, " ")...)
		var cmd *exec.Cmd
		if len(cmdArgs) == 1 {
			cmd = exec.Command(cmdArgs[0]) //nolint:gosec
		} else {
			cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec
		}

		utils.IsolateProcess(cmd)
		// if user is not empty, and OS is linux or macos
		// execute php worker from that particular user
		if p.cfg.Server.User != "" {
			err := utils.ExecuteFromUser(cmd, p.cfg.Server.User)
			if err != nil {
				return nil
			}
		}

		cmd.Env = p.setEnv(env)
		return cmd
	}
}

// NewWorker issues new standalone worker.
func (p *Plugin) NewWorker(ctx context.Context, env Env) (*worker.Process, error) {
	const op = errors.Op("server_plugin_new_worker")

	spawnCmd := p.CmdFactory(env)

	w, err := p.factory.SpawnWorkerWithTimeout(ctx, spawnCmd())
	if err != nil {
		return nil, errors.E(op, err)
	}

	return w, nil
}

// NewWorkerPool issues new worker pool.
func (p *Plugin) NewWorkerPool(ctx context.Context, opt *pool.Config, env Env) (pool.Pool, error) {
	spawnCmd := p.CmdFactory(env)

	pl, err := pool.Initialize(ctx, spawnCmd, p.factory, opt)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

// creates relay and worker factory.
func (p *Plugin) initFactory() (transport.Factory, error) {
	const op = errors.Op("server_plugin_init_factory")
	if p.cfg.Server.Relay == "" || p.cfg.Server.Relay == "pipes" {
		return pipe.NewPipeFactory(), nil
	}

	dsn := strings.Split(p.cfg.Server.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}

	lsn, err := utils.CreateListener(p.cfg.Server.Relay)
	if err != nil {
		return nil, errors.E(op, errors.Network, err)
	}

	switch dsn[0] {
	// sockets group
	case "unix":
		return socket.NewSocketServer(lsn, p.cfg.Server.RelayTimeout), nil
	case "tcp":
		return socket.NewSocketServer(lsn, p.cfg.Server.RelayTimeout), nil
	default:
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}
}

func (p *Plugin) setEnv(e Env) []string {
	env := append(os.Environ(), fmt.Sprintf(RrRelay+"=%s", p.cfg.Server.Relay))
	for k, v := range e {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	if p.cfg.RPC != nil && p.cfg.RPC.Listen != "" {
		env = append(env, fmt.Sprintf("%s=%s", RrRPC, p.cfg.RPC.Listen))
	}

	// set env variables from the config
	if len(p.cfg.Server.Env) > 0 {
		for k, v := range p.cfg.Server.Env {
			env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
		}
	}

	return env
}

//func (p *Plugin) collectPoolEvents(event interface{}) {
//	if we, ok := event.(events.PoolEvent); ok {
//		switch we.Event {
//		case events.EventMaxMemory:
//			p.log.Warn("worker max memory reached", "pid", we.Payload.(worker.BaseProcess).Pid())
//			// debug case
//			p.log.Debug("debug", "worker", we.Payload.(worker.BaseProcess))
//		case events.EventNoFreeWorkers:
//			p.log.Warn("no free workers in the pool, consider increasing `pool.num_workers` property, or `pool.allocate_timeout`")
//			// show error only in the debug mode
//			if we.Payload != nil {
//				p.log.Debug("error", we.Payload.(error).Error())
//			}
//		case events.EventWorkerProcessExit:
//			p.log.Info("worker process exited")
//			p.log.Debug("debug", "error", we.Error)
//		case events.EventSupervisorError:
//			p.log.Error("pool supervisor error, turn on debug logger to see the error")
//			// debug
//			p.log.Debug("debug", "error", we.Payload.(error).Error())
//		case events.EventTTL:
//			p.log.Warn("worker TTL reached", "pid", we.Payload.(worker.BaseProcess).Pid())
//		case events.EventWorkerConstruct:
//			if _, ok := we.Payload.(error); ok {
//				p.log.Error("worker construction error", "error", we.Payload.(error).Error())
//				return
//			}
//			p.log.Debug("worker constructed", "pid", we.Payload.(worker.BaseProcess).Pid())
//		case events.EventWorkerDestruct:
//			p.log.Debug("worker destructed", "pid", we.Payload.(worker.BaseProcess).Pid())
//		case events.EventExecTTL:
//			p.log.Warn("worker execute timeout reached, consider increasing pool supervisor options")
//			// debug
//			p.log.Debug("debug", "error", we.Payload.(error).Error())
//		case events.EventIdleTTL:
//			p.log.Warn("worker idle timeout reached", "pid", we.Payload.(worker.BaseProcess).Pid())
//		case events.EventPoolRestart:
//			p.log.Warn("requested pool restart")
//		}
//	}
//}
//
//func (p *Plugin) collectWorkerEvents(event interface{}) {
//	if we, ok := event.(events.WorkerEvent); ok {
//		switch we.Event {
//		case events.EventWorkerError:
//			switch e := we.Payload.(type) { //nolint:gocritic
//			case error:
//				if errors.Is(errors.SoftJob, e) {
//					// get source error for the softjob error
//					p.log.Error(e.(*errors.Error).Err.Error())
//					return
//				}
//
//				// print full error for the other types of errors
//				p.log.Error(e.Error())
//				return
//			}
//			p.log.Error(we.Payload.(error).Error())
//		case events.EventWorkerLog:
//			p.log.Debug(utils.AsString(we.Payload.([]byte)))
//			// stderr event is INFO level
//		case events.EventWorkerStderr:
//			p.log.Info(utils.AsString(we.Payload.([]byte)))
//		case events.EventWorkerWaitExit:
//			p.log.Debug("worker process exited", "reason", we.Payload)
//		}
//	}
//}
