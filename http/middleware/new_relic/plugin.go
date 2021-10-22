package newrelic

import (
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
)

const (
	pluginName string = "new_relic"
	path       string = "http.new_relic"
)

type Plugin struct {
	cfg *Config
	app *newrelic.Application
}

func (p *Plugin) Init(cfg config.Configurer) error {
	const op = errors.Op("new_relic_mdw_init")
	if !cfg.Has(path) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(path, &p.cfg)
	if err != nil {
		return err
	}

	err = p.cfg.InitDefaults()
	if err != nil {
		return err
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(p.cfg.AppName),
		newrelic.ConfigLicense(p.cfg.LicenseKey),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if err != nil {
		return err
	}

	p.app = app

	return nil
}

func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txn := p.app.StartTransaction(r.RequestURI)

		w = txn.SetWebResponse(w)
		txn.SetWebRequestHTTP(r)

		r = newrelic.RequestWithTransactionContext(r, txn)

		next.ServeHTTP(w, r)

		txn.End()
	})
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return pluginName
}
