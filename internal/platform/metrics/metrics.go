package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var durationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5}

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	CartOperationsTotal *prometheus.CounterVec

	registry *prometheus.Registry
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	httpTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	httpDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "platform_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: durationBuckets,
	}, []string{"method", "path"})

	cartOps := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cart_operations_total",
		Help: "Total cart operations by type.",
	}, []string{"operation"})

	reg.MustRegister(httpTotal, httpDuration, cartOps)

	return &Metrics{
		HTTPRequestsTotal:   httpTotal,
		HTTPRequestDuration: httpDuration,
		CartOperationsTotal: cartOps,
		registry:            reg,
	}
}

func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}
