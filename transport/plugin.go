package transport

import (
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/config"
	commonHttp "github.com/spiral/roadrunner-plugins/v2/internal/common/http"
	"github.com/spiral/roadrunner-plugins/v2/internal/common/transport"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/server"
)

const (
	name string = "transport"
)

type Plugin struct {
	log    logger.ZapLogger
	cfg    config.Configurer
	server server.Server

	httpMiddleware map[string]commonHttp.Middleware
	drivers        map[string]transport.Driver
	constructors   map[string]transport.Constructor
}

func (p *Plugin) Init(cfg config.Configurer, log logger.ZapLogger, server server.Server) error {
	p.log = log
	p.cfg = cfg
	p.server = server
	p.httpMiddleware = make(map[string]commonHttp.Middleware)
	p.drivers = make(map[string]transport.Driver)
	p.constructors = make(map[string]transport.Constructor)
	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	//

	return errCh
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectHTTPMiddleware,
		// collect drivers (http, grpc, fcgi, ws, etc..)
		p.CollectProtocolDrivers,
	}
}

// CollectHTTPMiddleware collects middleware plugins
func (p *Plugin) CollectHTTPMiddleware(n endure.Named, mdw commonHttp.Middleware) {
	p.httpMiddleware[n.Name()] = mdw
}

// CollectProtocolDrivers collects protocol (http1,2,3(s), grpc, etc) drivers
func (p *Plugin) CollectProtocolDrivers(n endure.Named, dr transport.Constructor) {
	p.constructors[n.Name()] = dr
}

func (p *Plugin) Name() string {
	return name
}
