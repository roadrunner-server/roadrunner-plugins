package tcp

import (
	"bytes"
	"context"
	"net"
	"sync"

	"github.com/google/uuid"
	rrErrors "github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/tcp/handler"
	"github.com/spiral/roadrunner-plugins/v2/utils"
	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/pool"
)

const (
	pluginName string = "tcp"
	RrMode     string = "RR_MODE"
)

type Plugin struct {
	sync.RWMutex
	cfg         *Config
	log         logger.Logger
	server      server.Server
	connections sync.Map // uuid -> conn

	wPool     pool.Pool
	listeners sync.Map // server -> listener

	resBufPool   sync.Pool
	readBufPool  sync.Pool
	servInfoPool sync.Pool
	pldPool      sync.Pool
}

func (p *Plugin) Init(log logger.Logger, cfg config.Configurer, server server.Server) error {
	const op = rrErrors.Op("tcp_plugin_init")

	if !cfg.Has(pluginName) {
		return rrErrors.E(op, rrErrors.Disabled)
	}

	err := cfg.UnmarshalKey(pluginName, &p.cfg)
	if err != nil {
		return rrErrors.E(op, err)
	}

	err = p.cfg.InitDefault()
	if err != nil {
		return err
	}

	// buffer sent to the user
	p.resBufPool = sync.Pool{
		New: func() interface{} {
			buf := new(bytes.Buffer)
			buf.Grow(p.cfg.ReadBufferSize)
			return buf
		},
	}

	// cyclic buffer to read the data from the connection
	p.readBufPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, p.cfg.ReadBufferSize)
			return &buf
		},
	}

	p.servInfoPool = sync.Pool{
		New: func() interface{} {
			return new(handler.ServerInfo)
		},
	}

	p.pldPool = sync.Pool{
		New: func() interface{} {
			return new(payload.Payload)
		},
	}

	p.log = log
	p.server = server
	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.wPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: pluginName})
	if err != nil {
		errCh <- err
		return errCh
	}

	for k := range p.cfg.Servers {
		go func(addr string, delim []byte, name string) {
			// create a TCP listener
			l, err := utils.CreateListener(addr)
			if err != nil {
				errCh <- err
				return
			}

			p.listeners.Store(uuid.NewString(), l)

			for {
				conn, err := l.Accept()
				if err != nil {
					p.log.Warn("connection accept failed", "error", err)
					// just stop
					return
				}

				go func() {
					h := handler.NewHandler(conn, delim, name, p.Exec, &p.pldPool, &p.servInfoPool, &p.readBufPool, &p.resBufPool, &p.connections, p.log)
					h.Start()
					// release resources
					h.Release()
				}()
			}
		}(p.cfg.Servers[k].Addr, p.cfg.Servers[k].delimBytes, k)
	}

	return errCh
}

func (p *Plugin) Stop() error {
	// close all connections
	p.connections.Range(func(_, value interface{}) bool {
		conn := value.(net.Conn)
		if conn != nil {
			_ = conn.Close()
		}
		return true
	})

	// then close all listeners
	p.listeners.Range(func(_, value interface{}) bool {
		_ = value.(net.Listener).Close()
		return true
	})

	return nil
}

func (p *Plugin) Reset() error {
	p.Lock()
	defer p.Unlock()

	var err error
	p.wPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: pluginName})
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) Close(uuid string) error {
	if c, ok := p.connections.LoadAndDelete(uuid); ok {
		conn := c.(net.Conn)
		if conn != nil {
			return conn.Close()
		}
	}

	return nil
}

func (p *Plugin) RPC() interface{} {
	return &rpc{
		p: p,
	}
}

func (p *Plugin) Exec(pld *payload.Payload) (*payload.Payload, error) {
	p.RLock()
	rsp, err := p.wPool.Exec(pld)
	if err != nil {
		p.RUnlock()
		return nil, err
	}

	p.RUnlock()
	return rsp, nil
}
