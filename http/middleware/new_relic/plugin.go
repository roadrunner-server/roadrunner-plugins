package newrelic

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/utils"
)

const (
	pluginName             string = "new_relic"
	path                   string = "http.new_relic"
	rrNewRelicKey          string = "Rr_newrelic"
	rrNewRelicErr          string = "Rr_newrelic_error"
	newRelicTransactionKey string = "transaction_name"
)

type Plugin struct {
	cfg *Config
	app *newrelic.Application

	writersPool sync.Pool
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

	p.writersPool = sync.Pool{
		New: func() interface{} {
			wr := new(writer)
			wr.code = -1
			wr.hdrToSend = make(map[string][]string, 10)
			return wr
		},
	}
	p.app = app

	return nil
}

func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txn := p.app.StartTransaction(r.RequestURI)
		defer txn.End()

		w = txn.SetWebResponse(w)
		txn.SetWebRequestHTTP(r)

		// overwrite original rw, because we need to delete sensitive rr_newrelic headers
		rrWriter := p.getWriter(w)
		defer p.putWriter(rrWriter)

		r = newrelic.RequestWithTransactionContext(r, txn)

		next.ServeHTTP(rrWriter, r)

		// handle error
		if len(rrWriter.hdrToSend[rrNewRelicErr]) > 0 {
			err := handleErr(rrWriter.hdrToSend[rrNewRelicErr])
			txn.NoticeError(err)

			// to be sure
			delete(rrWriter.hdrToSend, rrNewRelicKey)
			delete(rrWriter.hdrToSend, rrNewRelicErr)

			for k := range rrWriter.hdrToSend {
				for kk := range rrWriter.hdrToSend[k] {
					w.Header().Add(k, rrWriter.hdrToSend[k][kk])
				}
			}

			return
		}

		// no error, general case
		hdr := rrWriter.hdrToSend[rrNewRelicKey]
		if len(hdr) == 0 {
			// to be sure
			delete(rrWriter.hdrToSend, rrNewRelicKey)

			for k := range rrWriter.hdrToSend {
				for kk := range rrWriter.hdrToSend[k] {
					w.Header().Add(k, rrWriter.hdrToSend[k][kk])
				}
			}

			return
		}

		for i := 0; i < len(hdr); i++ {
			key, value := split(utils.AsBytes(hdr[i]))

			if key == nil || value == nil {
				continue
			}

			if bytes.Equal(key, utils.AsBytes(newRelicTransactionKey)) {
				txn.SetName(utils.AsString(value))
				continue
			}

			txn.AddAttribute(utils.AsString(key), utils.AsString(value))
		}

		// delete sensitive information
		delete(rrWriter.hdrToSend, rrNewRelicKey)

		// send original data
		for k := range rrWriter.hdrToSend {
			for kk := range rrWriter.hdrToSend[k] {
				w.Header().Add(k, rrWriter.hdrToSend[k][kk])
				delete(rrWriter.hdrToSend, k)
			}
		}
	})
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) getWriter(w http.ResponseWriter) *writer {
	wr := p.writersPool.Get().(*writer)
	wr.w = w
	return wr
}

func (p *Plugin) putWriter(w *writer) {
	w.code = -1
	w.w = nil

	for k := range w.hdrToSend {
		delete(w.hdrToSend, k)
	}

	p.writersPool.Put(w)
}
