package cache

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"sync"

	"github.com/roadrunner-server/api/v2/plugins/cache"
	"github.com/roadrunner-server/api/v2/plugins/config"
	cacheV1beta "github.com/roadrunner-server/api/v2/proto/cache/v1beta"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/cache/directives"
	"go.uber.org/zap"
)

const (
	root string = "http"
	name string = "cache"
)

type Plugin struct {
	hashPool    sync.Pool
	writersPool sync.Pool
	rspPool     sync.Pool
	rqPool      sync.Pool

	log             *zap.Logger
	cfg             *Config
	cache           cache.Cache
	collectedCaches map[string]cache.HTTPCacheFromConfig
}

func (p *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
	const op = errors.Op("cache_middleware_init")

	if !cfg.Has(fmt.Sprintf("%s.%s", root, name)) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(fmt.Sprintf("%s.%s", root, name), &p.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	// init default config values
	p.cfg.InitDefaults()

	p.hashPool = sync.Pool{
		New: func() interface{} {
			return fnv.New64a()
		},
	}

	p.writersPool = sync.Pool{
		New: func() interface{} {
			wr := new(writer)
			wr.Code = -1
			wr.Data = nil
			wr.HdrToSend = make(map[string][]string, 10)
			return wr
		},
	}

	p.rspPool = sync.Pool{New: func() interface{} {
		return &cacheV1beta.Response{}
	}}

	p.rqPool = sync.Pool{New: func() interface{} {
		return &directives.Req{}
	}}

	l := new(zap.Logger)
	*l = *log
	p.log = l
	p.collectedCaches = make(map[string]cache.HTTPCacheFromConfig, 1)

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	if _, ok := p.collectedCaches[p.cfg.Driver]; ok {
		p.cache, _ = p.collectedCaches[p.cfg.Driver].FromConfig(p.log)
		return errCh
	}

	errCh <- errors.E("no cache drivers registered")
	return errCh
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectCaches,
	}
}

func (p *Plugin) CollectCaches(name endure.Named, cache cache.HTTPCacheFromConfig) {
	p.collectedCaches[name.Name()] = cache
}

func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// https://datatracker.ietf.org/doc/html/rfc7234#section-3.2
		/*
			we MUST NOT use a cached response to a request with an Authorization header field
		*/
		if w.Header().Get(auth) != "" {
			next.ServeHTTP(w, r)
			return
		}

		rq := p.getRq()
		defer p.putRq(rq)

		// cwe-117
		cc := r.Header.Get(cacheControl)
		cc = strings.ReplaceAll(cc, "\n", "")
		cc = strings.ReplaceAll(cc, "\r", "")

		directives.ParseRequestCacheControl(cc, p.log, rq)
		// https://datatracker.ietf.org/doc/html/rfc7234#section-5.2.1.5
		/*
			The "no-store" request directive indicates that a cache MUST NOT
			store any part of either this request or any response to it.
		*/
		if rq.NoCache {
			next.ServeHTTP(w, r)
			return
		}

		switch r.Method {
		/*
			cacheable statuses by default: https://www.rfc-editor.org/rfc/rfc7231#section-6.1
			cacheable methods: https://www.rfc-editor.org/rfc/rfc7231#section-4.2.3 (GET, HEAD, POST (Responses to POST requests are only cacheable when they include explicit freshness information))
		*/
		case http.MethodGet:
			p.handleGET(w, r, next, rq)
			return
		case http.MethodHead:
			// TODO(rustatian): HEAD method is not supported
			fallthrough
		case http.MethodPost:
			// TODO(rustatian): POST method is not supported
			fallthrough
		default:
			// passthrough request to the worker for other methods
			next.ServeHTTP(w, r)
		}
	})
}

func (p *Plugin) Name() string {
	return name
}
