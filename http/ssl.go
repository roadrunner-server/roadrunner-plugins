package http

import (
	"net/http"
	"path"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/utils"
)

func (p *Plugin) serveHTTPS(errCh chan error) {
	if p.https == nil {
		return
	}
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
		err = p.https.ServeTLS(
			l,
			path.Join(p.cfg.Acme.CacheDir, p.cfg.Acme.CertificateName),
			path.Join(p.cfg.Acme.CacheDir, p.cfg.Acme.PrivateKeyName),
		)
		if err != nil && err != http.ErrServerClosed {
			errCh <- errors.E(op, err)
			return
		}
		return
	}

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
