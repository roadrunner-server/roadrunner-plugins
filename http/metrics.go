package http

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spiral/roadrunner-plugins/v2/informer"
)

func (p *Plugin) MetricsCollector() []prometheus.Collector {
	return []prometheus.Collector{p.statsExporter, requestCounter, requestDuration}
}

const (
	namespace string = "rr_http"
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

	workersMemory = prometheus.NewDesc("rr_http_workers_memory_bytes", "Memory usage by HTTP workers.", nil, nil)
)

type statsExporter struct {
	workers informer.Informer
}

func newWorkersExporter(stats informer.Informer) *statsExporter {
	return &statsExporter{
		workers: stats,
	}
}

func (s *statsExporter) Describe(d chan<- *prometheus.Desc) {
	// send description
	d <- workersMemory
}

func (s *statsExporter) Collect(ch chan<- prometheus.Metric) {
	// get the copy of the processes
	workers := s.workers.Workers()

	// cumulative RSS memory in bytes
	var cum uint64

	// collect the memory
	for i := 0; i < len(workers); i++ {
		cum += workers[i].MemoryUsage
	}

	// send the values to the prometheus
	ch <- prometheus.MustNewConstMetric(workersMemory, prometheus.GaugeValue, float64(cum))
}
