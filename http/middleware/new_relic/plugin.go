package newrelic

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	pluginName             string = "new_relic"
	path                   string = "http.new_relic"
	rrNewRelicKey          string = "Rr_newrelic"
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

		w = txn.SetWebResponse(w)
		txn.SetWebRequestHTTP(r)

		// overwrite original rw, because we need to delete sensitive rr_newrelic headers
		rrWriter := p.getWriter(w)

		defer p.putWriter(rrWriter)

		r = newrelic.RequestWithTransactionContext(r, txn)

		next.ServeHTTP(rrWriter, r)

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
			// 58 according to the ASCII table is -> :
			pos := bytes.IndexByte(utils.AsBytes(hdr[i]), 58)

			// not found
			if pos == -1 {
				continue
			}

			// remove spaces
			trimmed := bytes.Trim(utils.AsBytes(hdr[i]), " ")

			// ":foo" or ":"
			if pos == 0 {
				continue
			}

			/*
				we should split headers into 2 parts. Parts are separated by the colon (:)
				"foo:bar"
				we should not panic on cases like:
				":foo"
				"foo: bar"
				:
			*/

			// handle case like this "bar:"
			if len(trimmed) < pos+1 {
				continue
			}

			if bytes.Equal(trimmed[:pos], utils.AsBytes(newRelicTransactionKey)) {
				txn.SetName(utils.AsString(trimmed[pos+1:]))
				continue
			}

			txn.AddAttribute(utils.AsString(trimmed[:pos]), utils.AsString(trimmed[pos+1:]))
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

		txn.End()
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
	p.writersPool.Put(w)
}
