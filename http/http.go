package http

import (
	"net/http"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/utils"
	"go.uber.org/zap"
)

func (p *Plugin) serveHTTP(errCh chan error) {
	const op = errors.Op("serveHTTP")

	if len(p.mdwr) > 0 {
		applyMiddlewares(p.http, p.mdwr, p.cfg.Middleware, p.log)
	}

	l, err := utils.CreateListener(p.cfg.Address)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	p.log.Debug("http server was started", zap.String("address", p.cfg.Address))
	err = p.http.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		errCh <- errors.E(op, err)
		return
	}
}
