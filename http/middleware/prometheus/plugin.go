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
)

var (
	requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "request_total",
		Help:      "Total number of handled http requests after server restart.",
	}, []string{"status"})

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration.",
		},
		[]string{"status"},
	)
)

type Plugin struct {
	writersPool sync.Pool
}

func (p *Plugin) Init() error {
	p.writersPool = sync.Pool{
		New: func() interface{} {
			wr := new(writer)
			wr.code = -1
			return wr
		},
	}

	return nil
}

func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// overwrite original rw, because we need to delete sensitive rr_newrelic headers
		rrWriter := p.getWriter(w)
		defer p.putWriter(rrWriter)

		next.ServeHTTP(rrWriter, r)

		requestCounter.With(prometheus.Labels{
			"status": strconv.Itoa(rrWriter.code),
		}).Inc()

		requestDuration.With(prometheus.Labels{
			"status": strconv.Itoa(rrWriter.code),
		}).Observe(time.Since(start).Seconds())
	})
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) MetricsCollector() []prometheus.Collector {
	return []prometheus.Collector{requestCounter, requestDuration}
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
