package prometheus

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spiral/roadrunner/v2/events"
)

const (
	pluginName string = "http_metrics"
	namespace  string = "rr_http"
)

var (
	queueSize = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: namespace,
		Name:      "requests_queue",
		Help:      "Total number of queued requests.",
	})

	noFreeWorkers = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "no_free_workers_total",
		Help:      "Total number of NoFreeWorkers occurrences.",
	}, nil)

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
	stopCh      chan struct{}
	id          string
	bus         events.EventBus

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

	p.bus, p.id = events.Bus()
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
	errCh := make(chan error, 1)

	eventsCh := make(chan events.Event, 10)
	err := p.bus.SubscribeP(p.id, "pool.EventNoFreeWorkers", eventsCh)
	if err != nil {
		errCh <- err
		return errCh
	}

	go func() {
		for {
			select {
			case <-eventsCh:
				// increment no free workers event
				noFreeWorkers.With(nil).Inc()
			case <-p.stopCh:
				p.bus.Unsubscribe(p.id)
			}
		}
	}()

	return errCh
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

		requestCounter.With(prometheus.Labels{
			"status": strconv.Itoa(rrWriter.code),
		}).Inc()

		requestDuration.With(prometheus.Labels{
			"status": strconv.Itoa(rrWriter.code),
		}).Observe(time.Since(start).Seconds())

		p.queueSize.Dec()
	})
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) MetricsCollector() []prometheus.Collector {
	return []prometheus.Collector{requestCounter, requestDuration, queueSize, noFreeWorkers}
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
