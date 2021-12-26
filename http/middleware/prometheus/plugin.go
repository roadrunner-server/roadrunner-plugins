package prometheus

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	pluginName string = "http_metrics"
	namespace  string = "rr_http"

	// should be in sync with the http/handler.go constants
	noWorkers string = "No-Workers"
	trueStr   string = "true"
)

type Plugin struct {
	writersPool sync.Pool
	stopCh      chan struct{}

	queueSize       prometheus.Gauge
	noFreeWorkers   *prometheus.CounterVec
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func (p *Plugin) Init() error {
	p.writersPool = sync.Pool{
		New: func() interface{} {
			wr := new(writer)
			wr.code = -1
			return wr
		},
	}

	p.stopCh = make(chan struct{}, 1)
	p.queueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "requests_queue",
		Help:      "Total number of queued requests.",
	})

	p.noFreeWorkers = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "no_free_workers_total",
		Help:      "Total number of NoFreeWorkers occurrences.",
	}, nil)

	p.requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "request_total",
		Help:      "Total number of handled http requests after server restart.",
	}, []string{"status"})

	p.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration.",
		},
		[]string{"status"},
	)

	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	p.stopCh <- struct{}{}
	return nil
}

func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// overwrite original rw, because we need to delete sensitive rr_newrelic headers
		rrWriter := p.getWriter(w)
		defer p.putWriter(rrWriter)

		p.queueSize.Inc()

		next.ServeHTTP(rrWriter, r)

		if w.Header().Get(noWorkers) == trueStr {
			p.noFreeWorkers.With(nil).Inc()
		}

		p.requestCounter.With(prometheus.Labels{
			"status": strconv.Itoa(rrWriter.code),
		}).Inc()

		p.requestDuration.With(prometheus.Labels{
			"status": strconv.Itoa(rrWriter.code),
		}).Observe(time.Since(start).Seconds())

		p.queueSize.Dec()
	})
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) MetricsCollector() []prometheus.Collector {
	return []prometheus.Collector{p.requestCounter, p.requestDuration, p.queueSize, p.noFreeWorkers}
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
