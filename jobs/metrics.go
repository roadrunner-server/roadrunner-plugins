package jobs

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/roadrunner-server/api/v2/plugins/informer"
)

func (p *Plugin) MetricsCollector() []prometheus.Collector {
	// p - implements Exporter interface (workers)
	// other - request duration and count
	return []prometheus.Collector{p.statsExporter}
}

const (
	namespace = "rr_jobs"
)

type statsExporter struct {
	workers       informer.Informer
	workersMemory uint64
	jobsOk        *uint64
	pushOk        *uint64
	jobsErr       *uint64
	pushErr       *uint64

	worker      *prometheus.Desc
	pushOkDesc  *prometheus.Desc
	pushErrDesc *prometheus.Desc
	jobsErrDesc *prometheus.Desc
	jobsOkDesc  *prometheus.Desc
}

func newStatsExporter(stats informer.Informer, jobsOk, pushOk, jobsErr, pushErr *uint64) *statsExporter {
	return &statsExporter{
		workers:       stats,
		workersMemory: 0,
		jobsOk:        jobsOk,
		pushOk:        pushOk,
		jobsErr:       jobsErr,
		pushErr:       pushErr,

		worker:      prometheus.NewDesc("workers_memory_bytes", "Memory usage by JOBS workers.", nil, nil),
		pushOkDesc:  prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "push_ok"), "Number of job push.", nil, nil),
		pushErrDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "push_err"), "Number of jobs push which was failed.", nil, nil),
		jobsErrDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "jobs_err"), "Number of jobs error while processing in the worker.", nil, nil),
		jobsOkDesc:  prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "jobs_ok"), "Number of successfully processed jobs.", nil, nil),
	}
}

func (se *statsExporter) Describe(d chan<- *prometheus.Desc) {
	// send description
	d <- se.worker
	d <- se.pushErrDesc
	d <- se.pushOkDesc
	d <- se.jobsErrDesc
	d <- se.jobsOkDesc
}

func (se *statsExporter) Collect(ch chan<- prometheus.Metric) {
	// get the copy of the processes
	workers := se.workers.Workers()

	// cumulative RSS memory in bytes
	var cum uint64

	// collect the memory
	for i := 0; i < len(workers); i++ {
		cum += workers[i].MemoryUsage
	}

	// send the values to the prometheus
	ch <- prometheus.MustNewConstMetric(se.worker, prometheus.GaugeValue, float64(cum))
	// send the values to the prometheus
	ch <- prometheus.MustNewConstMetric(se.jobsOkDesc, prometheus.GaugeValue, float64(atomic.LoadUint64(se.jobsOk)))
	ch <- prometheus.MustNewConstMetric(se.jobsErrDesc, prometheus.GaugeValue, float64(atomic.LoadUint64(se.jobsErr)))
	ch <- prometheus.MustNewConstMetric(se.pushOkDesc, prometheus.GaugeValue, float64(atomic.LoadUint64(se.pushOk)))
	ch <- prometheus.MustNewConstMetric(se.pushErrDesc, prometheus.GaugeValue, float64(atomic.LoadUint64(se.pushErr)))
}
