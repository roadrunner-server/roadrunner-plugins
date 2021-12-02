package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/events"

	"github.com/spiral/roadrunner-plugins/v2/utils"
	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/transport"
	"github.com/spiral/roadrunner/v2/transport/pipe"
	"github.com/spiral/roadrunner/v2/transport/socket"
	"github.com/spiral/roadrunner/v2/worker"
)

const (
	// PluginName for the server
	PluginName string = "server"

	// RPCPluginName is the name of the RPC plugin, should be in sync with rpc/config.go
	RPCPluginName string = "rpc"

	// RrRelay env variable key (internal)
	RrRelay string = "RR_RELAY"

	// RrRPC env variable key (internal) if the RPC presents
	RrRPC string = "RR_RPC"

	// Event types

	PoolEvents    string = "pool.*"
	WorkerEvents  string = "worker.*"
	WatcherEvents string = "worker_watcher.*"
)

// Plugin manages worker
type Plugin struct {
	sync.Mutex

	cfg    *Config
	rpcCfg *RPCConfig

	log     logger.Logger
	factory transport.Factory
	stopCh  chan struct{}
}

// Init application provider.
func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("server_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, errors.Init, err)
	}

	err = cfg.UnmarshalKey(RPCPluginName, &p.rpcCfg)
	if err != nil {
		return errors.E(op, errors.Init, err)
	}

	err = p.cfg.InitDefaults()
	if err != nil {
		return errors.E(op, errors.Init, err)
	}

	p.factory, err = p.initFactory()
	if err != nil {
		return errors.E(op, err)
	}

	p.log = log
	p.stopCh = make(chan struct{})

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
	errCh := make(chan error, 1)

	go p.startEventsBus(errCh)

	if p.cfg.OnInit != nil {
		err := p.runOnInitCommand()
		if err != nil {
			p.log.Error("on_init finished with errors, RR continues to run", "error", err)
		}
	}

	return errCh
}

// Stop used to close chosen in config factory
func (p *Plugin) Stop() error {
	if p.factory == nil {
		return nil
	}

	p.stopCh <- struct{}{}

	return p.factory.Close()
}

// CmdFactory provides worker command factory associated with given context.
func (p *Plugin) CmdFactory(env Env) func() *exec.Cmd {
	return func() *exec.Cmd {
		var cmdArgs []string
		// create command according to the config
		cmdArgs = append(cmdArgs, strings.Split(p.cfg.Command, " ")...)
		var cmd *exec.Cmd
		if len(cmdArgs) == 1 {
			cmd = exec.Command(cmdArgs[0]) //nolint:gosec
		} else {
			cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec
		}

		utils.IsolateProcess(cmd)
		// if user is not empty, and OS is linux or macos
		// execute php worker from that particular user
		if p.cfg.User != "" {
			err := utils.ExecuteFromUser(cmd, p.cfg.User)
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
	p.Lock()
	defer p.Unlock()
	return pool.Initialize(ctx, p.CmdFactory(env), p.factory, opt)
}

// creates relay and worker factory.
func (p *Plugin) initFactory() (transport.Factory, error) {
	const op = errors.Op("server_plugin_init_factory")
	if p.cfg.Relay == "" || p.cfg.Relay == "pipes" {
		return pipe.NewPipeFactory(), nil
	}

	dsn := strings.Split(p.cfg.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}

	lsn, err := utils.CreateListener(p.cfg.Relay)
	if err != nil {
		return nil, errors.E(op, errors.Network, err)
	}

	switch dsn[0] {
	// sockets group
	case "unix":
		return socket.NewSocketServer(lsn, p.cfg.RelayTimeout), nil
	case "tcp":
		return socket.NewSocketServer(lsn, p.cfg.RelayTimeout), nil
	default:
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}
}

func (p *Plugin) setEnv(e Env) []string {
	env := append(os.Environ(), fmt.Sprintf(RrRelay+"=%s", p.cfg.Relay))
	for k, v := range e {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	if p.rpcCfg != nil && p.rpcCfg.Listen != "" {
		env = append(env, fmt.Sprintf("%s=%s", RrRPC, p.rpcCfg.Listen))
	}

	// set env variables from the config
	if len(p.cfg.Env) > 0 {
		for k, v := range p.cfg.Env {
			env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
		}
	}

	return env
}

func (p *Plugin) startEventsBus(errCh chan<- error) {
	eb, id := events.Bus()
	eventsCh := make(chan events.Event, 10)

	err := eb.SubscribeP(id, PoolEvents, eventsCh)
	if err != nil {
		errCh <- err
		return
	}

	err = eb.SubscribeP(id, WorkerEvents, eventsCh)
	if err != nil {
		errCh <- err
		return
	}
	err = eb.SubscribeP(id, WatcherEvents, eventsCh)
	if err != nil {
		errCh <- err
		return
	}

	for {
		select {
		case event := <-eventsCh:
			p.log.Info(event.Message())
		case <-p.stopCh:
			eb.Unsubscribe(id)
			return
		}
	}
}
