// Package metrics provides Prometheus instrumentation for the cart service.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all Prometheus metrics.
// Uses a private registry so that tests can create multiple instances safely.
type Metrics struct {
	CartOperationsTotal *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	GRPCRequestDuration *prometheus.HistogramVec
	HTTPRequestsTotal   *prometheus.CounterVec
	GRPCRequestsTotal   *prometheus.CounterVec

	registry *prometheus.Registry
}

// New creates and registers all metrics under the given namespace.
func New(namespace string) *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	cartOps := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "cart_operations_total",
		Help:      "Total number of cart operations partitioned by operation type.",
	}, []string{"operation"})

	httpDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	grpcDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "grpc_request_duration_seconds",
		Help:      "gRPC request latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status"})

	httpTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	grpcTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "grpc_requests_total",
		Help:      "Total number of gRPC requests.",
	}, []string{"method", "status"})

	reg.MustRegister(cartOps, httpDuration, grpcDuration, httpTotal, grpcTotal)

	return &Metrics{
		CartOperationsTotal: cartOps,
		HTTPRequestDuration: httpDuration,
		GRPCRequestDuration: grpcDuration,
		HTTPRequestsTotal:   httpTotal,
		GRPCRequestsTotal:   grpcTotal,
		registry:            reg,
	}
}

// Registry returns the Prometheus registry, suitable for promhttp.HandlerFor.
func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}
