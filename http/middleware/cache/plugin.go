package cache

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	cacheV1beta "github.com/spiral/roadrunner-plugins/v2/api/proto/cache/v1beta"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/cache"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/config"
	"github.com/spiral/roadrunner/v2/utils"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	root string = "http"
	name string = "cache"
)

type Plugin struct {
	// methods mask
	hashPool    sync.Pool
	writersPool sync.Pool
	rspPool     sync.Pool

	log   *zap.Logger
	cfg   *Config
	cache cache.Cache
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

	l := new(zap.Logger)
	*l = *log
	p.log = l

	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectCaches,
	}
}

func (p *Plugin) CollectCaches(_ endure.Named, cache cache.HTTPCacheFromConfig) {
	c, err := cache.FromConfig(p.log)
	if err != nil {
		panic(err)
	}
	p.cache = c
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

		// get Cache-Control header
		switch r.Header.Get(cacheControl) {
		// https://datatracker.ietf.org/doc/html/rfc7234#section-5.2.1.5
		/*
			The "no-store" request directive indicates that a cache MUST NOT
			store any part of either this request or any response to it.
		*/
		case noStore, empty:
			next.ServeHTTP(w, r)
			return
		}

		wr := p.getWriter()
		defer p.putWriter(wr)

		switch r.Method {
		/*
			cacheable statuses by default: https://www.rfc-editor.org/rfc/rfc7231#section-6.1
			cacheable methods: https://www.rfc-editor.org/rfc/rfc7231#section-4.2.3 (GET, HEAD, POST (Responses to POST requests are only cacheable when they include explicit freshness information))
		*/
		case http.MethodGet:
			h := p.getHash()
			defer p.putHash(h)

			// write the data to the hash function
			_, err := h.Write(utils.AsBytes(r.RequestURI))
			if err != nil {
				http.Error(w, "failed to write the hash", http.StatusInternalServerError)
				return
			}

			// try to get the data from cache
			out, err := p.cache.Get(h.Sum64())
			if err != nil {
				// cache miss, no data
				if errors.Is(errors.EmptyItem, err) {
					// forward the
					next.ServeHTTP(wr, r)

					// send original data
					for k := range wr.HdrToSend {
						for kk := range wr.HdrToSend[k] {
							w.Header().Add(k, wr.HdrToSend[k][kk])
						}
					}

					w.WriteHeader(wr.Code)
					_, _ = w.Write(wr.Data)

					p.handleCacheMiss(wr, h.Sum64())
					return
				}

				http.Error(w, "get hash", http.StatusInternalServerError)
				return
			}

			msg := p.getRsp()
			defer p.putRsp(msg)

			err = proto.Unmarshal(out, msg)
			if err != nil {
				http.Error(w, "cache data unpack", http.StatusInternalServerError)
				return
			}

			// send original data
			for k := range msg.Headers {
				for i := 0; i < len(msg.Headers[k].Value); i++ {
					w.Header().Add(k, msg.Headers[k].Value[i])
				}
			}

			w.WriteHeader(int(msg.Code))
			_, _ = w.Write(msg.Data)
			return
		case http.MethodHead:
		case http.MethodPost:
		default:
			panic("non-cacheable")
		}

		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) Name() string {
	return name
}
