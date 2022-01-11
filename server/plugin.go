package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/ipc"
	"go.uber.org/zap"

	"github.com/roadrunner-server/api/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/ipc/pipe"
	"github.com/spiral/roadrunner/v2/ipc/socket"
	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/utils"
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
)

// Plugin manages worker
type Plugin struct {
	sync.Mutex

	cfg          *Config
	rpcCfg       *RPCConfig
	preparedCmd  []string
	preparedEnvs []string

	log     *zap.Logger
	factory ipc.Factory

	pools []pool.Pool
}

// Init application provider.
func (p *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
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

	p.log = new(zap.Logger)
	*p.log = *log
	p.preparedCmd = append(p.preparedCmd, strings.Split(p.cfg.Command, " ")...)

	p.preparedEnvs = append(os.Environ(), fmt.Sprintf(RrRelay+"=%s", p.cfg.Relay))
	if p.rpcCfg != nil && p.rpcCfg.Listen != "" {
		p.preparedEnvs = append(p.preparedEnvs, fmt.Sprintf("%s=%s", RrRPC, p.rpcCfg.Listen))
	}

	// set env variables from the config
	if len(p.cfg.Env) > 0 {
		for k, v := range p.cfg.Env {
			p.preparedEnvs = append(p.preparedEnvs, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
		}
	}

	p.pools = make([]pool.Pool, 0, 4)

	return nil
}

// Name contains service name.
func (p *Plugin) Name() string {
	return PluginName
}

// Serve (Start) server plugin (just a mock here to satisfy interface)
func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	if p.cfg.OnInit != nil {
		err := p.runOnInitCommand()
		if err != nil {
			p.log.Error("on_init was finished with errors", zap.Error(err))
		}
	}

	return errCh
}

// Stop used to close chosen in config factory
func (p *Plugin) Stop() error {
	p.Lock()
	defer p.Unlock()

	// destroy all pools
	for i := 0; i < len(p.pools); i++ {
		if p.pools[i] != nil {
			p.pools[i].Destroy(context.Background())
		}
	}

	// just to be sure, that all logs are synced
	time.Sleep(time.Second)
	return p.factory.Close()
}

// CmdFactory provides worker command factory associated with given context.
func (p *Plugin) CmdFactory(env map[string]string) func() *exec.Cmd {
	return func() *exec.Cmd {
		var cmd *exec.Cmd

		if len(p.preparedCmd) == 1 {
			cmd = exec.Command(p.preparedCmd[0]) //nolint:gosec
		} else {
			cmd = exec.Command(p.preparedCmd[0], p.preparedCmd[1:]...) //nolint:gosec
		}

		// copy prepared envs
		cmd.Env = make([]string, len(p.preparedEnvs))
		copy(cmd.Env, p.preparedEnvs)

		// append external envs
		if len(env) > 0 {
			for k, v := range env {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
			}
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

		return cmd
	}
}

// NewWorker issues new standalone worker.
func (p *Plugin) NewWorker(ctx context.Context, env map[string]string) (*worker.Process, error) {
	const op = errors.Op("server_plugin_new_worker")

	spawnCmd := p.CmdFactory(env)

	w, err := p.factory.SpawnWorkerWithTimeout(ctx, spawnCmd())
	if err != nil {
		return nil, errors.E(op, err)
	}

	return w, nil
}

// NewWorkerPool issues new worker pool.
func (p *Plugin) NewWorkerPool(ctx context.Context, opt *pool.Config, env map[string]string, _ ...pool.Options) (pool.Pool, error) {
	p.Lock()
	defer p.Unlock()

	pl, err := pool.NewStaticPool(ctx, p.CmdFactory(env), p.factory, opt, pool.WithLogger(p.log))
	if err != nil {
		return nil, err
	}
	p.pools = append(p.pools, pl)

	return pl, nil
}

// creates relay and worker factory.
func (p *Plugin) initFactory() (ipc.Factory, error) {
	const op = errors.Op("server_plugin_init_factory")
	if p.cfg.Relay == "" || p.cfg.Relay == "pipes" {
		return pipe.NewPipeFactory(p.log), nil
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
		return socket.NewSocketServer(lsn, p.cfg.RelayTimeout, p.log), nil
	case "tcp":
		return socket.NewSocketServer(lsn, p.cfg.RelayTimeout, p.log), nil
	default:
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}
}
