package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/mholt/acmez"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/http/attributes"
	httpConfig "github.com/spiral/roadrunner-plugins/v2/http/config"
	"github.com/spiral/roadrunner-plugins/v2/http/handler"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/status"
	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/state/process"
	"github.com/spiral/roadrunner/v2/worker"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	// PluginName declares plugin name.
	PluginName = "http"

	// RrMode RR_HTTP env variable key (internal) if the HTTP presents
	RrMode = "RR_MODE"

	Scheme = "https"
)

// Middleware interface
type Middleware interface {
	Middleware(f http.Handler) http.Handler
}

type middleware map[string]Middleware

// Plugin manages pool, http servers. The main http plugin structure
type Plugin struct {
	mu sync.RWMutex

	// plugins
	server server.Server
	log    logger.Logger
	// stdlog passed to the http/https/fcgi servers to log their internal messages
	stdLog *log.Logger

	// http configuration
	cfg *httpConfig.HTTP `mapstructure:"http"`

	// middlewares to chain
	mdwr middleware

	// Pool which attached to all servers
	pool pool.Pool

	// servers RR handler
	handler *handler.Handler

	// metrics
	workersExporter  *workersExporter
	requestsExporter *requestsExporter

	// servers
	http  *http.Server
	https *http.Server
	fcgi  *http.Server
}

// Init must return configure svc and return true if svc hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (p *Plugin) Init(cfg config.Configurer, rrLogger logger.Logger, server server.Server) error {
	const op = errors.Op("http_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.cfg.InitDefaults()
	if err != nil {
		return errors.E(op, err)
	}

	// rr logger (via plugin)
	p.log = rrLogger
	// use time and date in UTC format
	p.stdLog = log.New(logger.NewStdAdapter(p.log), "http_plugin: ", log.Ldate|log.Ltime|log.LUTC)

	p.mdwr = make(map[string]Middleware)

	if !p.cfg.EnableHTTP() && !p.cfg.EnableTLS() && !p.cfg.EnableFCGI() {
		return errors.E(op, errors.Disabled)
	}

	// init if nil
	if p.cfg.Env == nil {
		p.cfg.Env = make(map[string]string)
	}
	p.cfg.Env[RrMode] = "http"

	// initialize workersExporter
	p.workersExporter = newWorkersExporter()
	// initialize requests exporter
	p.requestsExporter = newRequestsExporter()
	p.server = server

	return nil
}

func (p *Plugin) logCallback(event interface{}) {
	if ev, ok := event.(handler.ResponseEvent); ok {
		if p.cfg.AccessLogs {
			p.log.Info(
				"",
				"method", ev.Method,
				"remote_addr", ev.ReqRemoteAddr,
				"bytes_sent", ev.BytesSent,
				"http_host", ev.Host,
				"request", ev.Query,
				"time_local", ev.TimeLocal,
				"request_length", ev.ReqLen,
				"request_time", ev.Elapsed.Seconds(),
				"status", ev.Status,
				"http_user_agent", ev.UserAgent,
				"http_referer", ev.Referer,
			)
		} else {
			p.log.Debug(fmt.Sprintf("%s %s %s", ev.Status, ev.Method, ev.URI),
				"remote", ev.ReqRemoteAddr,
				"elapsed", ev.Elapsed.String(),
			)
		}
	}
}

// Serve serves the svc.
func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 2)
	// run whole process in the goroutine
	go func() {
		// protect http initialization
		p.mu.Lock()
		p.serve(errCh)
		p.mu.Unlock()
	}()

	return errCh
}

func (p *Plugin) serve(errCh chan error) {
	var err error
	p.pool, err = p.server.NewWorkerPool(context.Background(), &pool.Config{
		Debug:           p.cfg.Pool.Debug,
		NumWorkers:      p.cfg.Pool.NumWorkers,
		MaxJobs:         p.cfg.Pool.MaxJobs,
		AllocateTimeout: p.cfg.Pool.AllocateTimeout,
		DestroyTimeout:  p.cfg.Pool.DestroyTimeout,
		Supervisor:      p.cfg.Pool.Supervisor,
	}, p.cfg.Env)
	if err != nil {
		errCh <- err
		return
	}

	p.handler, err = handler.NewHandler(
		p.cfg.MaxRequestSize,
		p.cfg.InternalErrorCode,
		p.cfg.Uploads,
		p.cfg.Cidrs,
		p.pool,
		p.log,
		p.cfg.AccessLogs,
	)
	if err != nil {
		errCh <- err
		return
	}

	p.handler.AddListener(p.logCallback, p.metricsCallback)

	if p.cfg.EnableHTTP() {
		if p.cfg.EnableH2C() {
			p.http = &http.Server{
				Handler: h2c.NewHandler(p, &http2.Server{
					MaxHandlers:                  0,
					MaxConcurrentStreams:         0,
					MaxReadFrameSize:             0,
					PermitProhibitedCipherSuites: false,
					IdleTimeout:                  0,
					MaxUploadBufferPerConnection: 0,
					MaxUploadBufferPerStream:     0,
					NewWriteScheduler:            nil,
					CountError:                   nil,
				}),
				ErrorLog: p.stdLog,
			}
		} else {
			p.http = &http.Server{
				Handler:  p,
				ErrorLog: p.stdLog,
			}
		}
	}

	if p.cfg.EnableTLS() {
		p.https = p.initTLS()
		if p.cfg.SSLConfig.RootCA != "" {
			err = p.appendRootCa()
			if err != nil {
				errCh <- err
				return
			}
		}

		if p.cfg.EnableACME() {
			// for the first time - generate the certs
			tlsCfg, errObt := ObtainCertificates(
				p.cfg.SSLConfig.Acme.CacheDir,
				p.cfg.SSLConfig.Acme.Email,
				p.cfg.SSLConfig.Acme.ChallengeType,
				p.cfg.SSLConfig.Acme.Domains,
				p.cfg.SSLConfig.Acme.UseProductionEndpoint,
				p.cfg.SSLConfig.Acme.AltHTTPPort,
				p.cfg.SSLConfig.Acme.AltTLSALPNPort,
			)

			if errObt != nil {
				errCh <- errObt
				return
			}

			p.https.TLSConfig.GetCertificate = tlsCfg.GetCertificate
			p.https.TLSConfig.NextProtos = append(p.https.TLSConfig.NextProtos, acmez.ACMETLS1Protocol)
		}

		// if HTTP2Config not nil
		if p.cfg.HTTP2Config != nil {
			if err = p.initHTTP2(); err != nil {
				errCh <- err
				return
			}
		}
	}

	if p.cfg.EnableFCGI() {
		p.fcgi = &http.Server{Handler: p, ErrorLog: p.stdLog}
	}

	// start http, https and fcgi servers if requested in the config
	if p.http != nil {
		go p.serveHTTP(errCh)
		return
	}

	if p.https != nil {
		go p.serveHTTPS(errCh)
		return
	}

	if p.fcgi != nil {
		go p.serveFCGI(errCh)
		return
	}
}

// Stop stops the http.
func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.fcgi != nil {
		err := p.fcgi.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			p.log.Error("fcgi shutdown", "error", err)
		}
	}

	if p.https != nil {
		err := p.https.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			p.log.Error("https shutdown", "error", err)
		}
	}

	if p.http != nil {
		err := p.http.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			p.log.Error("http shutdown", "error", err)
		}
	}

	// check for safety
	if p.pool != nil {
		p.pool.Destroy(context.Background())
	}

	return nil
}

// ServeHTTP handles connection using set of middleware and pool PSR-7 server.
func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			p.log.Error("body close", "error", err)
		}
	}()

	if headerContainsUpgrade(r) {
		http.Error(w, "server does not support upgrade header", http.StatusInternalServerError)
		return
	}

	if p.https != nil && r.TLS == nil && p.cfg.SSLConfig.Redirect {
		p.redirect(w, r)
		return
	}

	if p.https != nil && r.TLS != nil {
		w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	}

	r = attributes.Init(r)
	// protect the case, when user sendEvent Reset, and we are replacing handler with pool
	p.mu.RLock()
	p.handler.ServeHTTP(w, r)
	p.mu.RUnlock()
}

// Workers returns slice with the process states for the workers
func (p *Plugin) Workers() []*process.State {
	p.mu.RLock()
	defer p.mu.RUnlock()

	workers := p.workers()

	ps := make([]*process.State, 0, len(workers))
	for i := 0; i < len(workers); i++ {
		state, err := process.WorkerProcessState(workers[i])
		if err != nil {
			return nil
		}
		ps = append(ps, state)
	}

	return ps
}

// internal
func (p *Plugin) workers() []worker.BaseProcess {
	return p.pool.Workers()
}

// Name returns endure.Named interface implementation
func (p *Plugin) Name() string {
	return PluginName
}

// Reset destroys the old pool and replaces it with new one, waiting for old pool to die
func (p *Plugin) Reset() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	const op = errors.Op("http_plugin_reset")
	p.log.Info("HTTP plugin got restart request. Restarting...")
	p.pool.Destroy(context.Background())
	p.pool = nil

	var err error
	p.pool, err = p.server.NewWorkerPool(context.Background(), &pool.Config{
		Debug:           p.cfg.Pool.Debug,
		NumWorkers:      p.cfg.Pool.NumWorkers,
		MaxJobs:         p.cfg.Pool.MaxJobs,
		AllocateTimeout: p.cfg.Pool.AllocateTimeout,
		DestroyTimeout:  p.cfg.Pool.DestroyTimeout,
		Supervisor:      p.cfg.Pool.Supervisor,
	}, p.cfg.Env)
	if err != nil {
		return errors.E(op, err)
	}

	p.log.Info("HTTP workers Pool successfully restarted")

	p.handler, err = handler.NewHandler(
		p.cfg.MaxRequestSize,
		p.cfg.InternalErrorCode,
		p.cfg.Uploads,
		p.cfg.Cidrs,
		p.pool,
		p.log,
		p.cfg.AccessLogs,
	)

	if err != nil {
		return errors.E(op, err)
	}

	p.log.Info("HTTP handler listeners successfully re-added")
	p.handler.AddListener(p.logCallback, p.metricsCallback)

	p.log.Info("HTTP plugin successfully restarted")
	return nil
}

// Collects collecting http middlewares
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.AddMiddleware,
	}
}

// AddMiddleware is base requirement for the middleware (name and Middleware)
func (p *Plugin) AddMiddleware(name endure.Named, m Middleware) {
	p.mdwr[name.Name()] = m
}

// Status return status of the particular plugin
func (p *Plugin) Status() status.Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	workers := p.workers()
	for i := 0; i < len(workers); i++ {
		if workers[i].State().IsActive() {
			return status.Status{
				Code: http.StatusOK,
			}
		}
	}
	// if there are no workers, threat this as error
	return status.Status{
		Code: http.StatusServiceUnavailable,
	}
}

// Ready return readiness status of the particular plugin
func (p *Plugin) Ready() status.Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	workers := p.workers()
	for i := 0; i < len(workers); i++ {
		// If state of the worker is ready (at least 1)
		// we assume, that plugin's worker pool is ready
		if workers[i].State().Value() == worker.StateReady {
			return status.Status{
				Code: http.StatusOK,
			}
		}
	}
	// if there are no workers, threat this as no content error
	return status.Status{
		Code: http.StatusServiceUnavailable,
	}
}

// Available interface implementation
func (p *Plugin) Available() {}
