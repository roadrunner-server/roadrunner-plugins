package http

import (
	"net/http"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/utils"
)

func (p *Plugin) serveHTTPS(errCh chan error) {
	const op = errors.Op("serveHTTPS")
	if len(p.mdwr) > 0 {
		applyMiddlewares(p.https, p.mdwr, p.cfg.Middleware, p.log)
	}

	l, err := utils.CreateListener(p.cfg.SSLConfig.Address)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	/*
		ACME powered server
	*/
	if p.cfg.EnableACME() {
		p.log.Debug("https(acme) server is running", "address", p.cfg.SSLConfig.Address)
		err = p.https.ServeTLS(
			l,
			"",
			"",
		)
		if err != nil && err != http.ErrServerClosed {
			errCh <- errors.E(op, err)
			return
		}
		return
	}

	p.log.Debug("https server is running", "address", p.cfg.SSLConfig.Address)
	err = p.https.ServeTLS(
		l,
		p.cfg.SSLConfig.Cert,
		p.cfg.SSLConfig.Key,
	)

	if err != nil && err != http.ErrServerClosed {
		errCh <- errors.E(op, err)
		return
	}
}
