package rpc

import (
	"net"
	"net/rpc"
	"sync/atomic"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/config"
	api "github.com/spiral/roadrunner-plugins/v2/api/v2/rpc"
	"go.uber.org/zap"
)

// PluginName contains default plugin name.
const PluginName = "rpc"

// Plugin is RPC service.
type Plugin struct {
	cfg Config
	log *zap.Logger
	rpc *rpc.Server
	// set of the plugins, which are implement RPCer interface and can be plugged into the RR via RPC
	plugins  map[string]api.RPCer
	listener net.Listener
	closed   uint32
}

// Init rpc service. Must return true if service is enabled.
func (s *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
	const op = errors.Op("rpc_plugin_init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	// Init defaults
	s.cfg.InitDefaults()
	// Init pluggable plugins map
	s.plugins = make(map[string]api.RPCer, 1)
	// init logs
	s.log = log

	// set up state
	atomic.StoreUint32(&s.closed, 0)

	// validate config
	err = s.cfg.Valid()
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Serve serves the service.
func (s *Plugin) Serve() chan error {
	const op = errors.Op("rpc_plugin_serve")
	errCh := make(chan error, 1)

	s.rpc = rpc.NewServer()

	plugins := make([]string, 0, len(s.plugins))

	// Attach all services
	for name := range s.plugins {
		err := s.Register(name, s.plugins[name].RPC())
		if err != nil {
			errCh <- errors.E(op, err)
			return errCh
		}

		plugins = append(plugins, name)
	}

	var err error
	s.listener, err = s.cfg.Listener()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	s.log.Debug("plugin was started", zap.String("address", s.cfg.Listen), zap.Strings("list of the plugins with RPC methods:", plugins))

	go func() {
		for {
			conn, errA := s.listener.Accept()
			if errA != nil {
				if atomic.LoadUint32(&s.closed) == 1 {
					// just continue, this is not a critical issue, we just called Stop
					return
				}

				s.log.Error("failed to accept the connection", zap.Error(errA))
				continue
			}

			go s.rpc.ServeCodec(goridgeRpc.NewCodec(conn))
		}
	}()

	return errCh
}

// Stop stops the service.
func (s *Plugin) Stop() error {
	const op = errors.Op("rpc_plugin_stop")
	// store closed state
	atomic.StoreUint32(&s.closed, 1)
	err := s.listener.Close()
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Name contains service name.
func (s *Plugin) Name() string {
	return PluginName
}

// Collects all plugins which implement Name + RPCer interfaces
func (s *Plugin) Collects() []interface{} {
	return []interface{}{
		s.RegisterPlugin,
	}
}

// RegisterPlugin registers RPC service plugin.
func (s *Plugin) RegisterPlugin(name endure.Named, p api.RPCer) {
	s.plugins[name.Name()] = p
}

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method of exported type
//	- two arguments, both of exported type
//	- the second argument is a pointer
//	- one return value, of type error
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
func (s *Plugin) Register(name string, svc interface{}) error {
	if s.rpc == nil {
		return errors.E("RPC service is not configured")
	}

	return s.rpc.RegisterName(name, svc)
}
