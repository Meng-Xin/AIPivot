package observability

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	registry        *prometheus.Registry
	httpRequests    *prometheus.CounterVec
	httpDuration    *prometheus.HistogramVec
	dependencyReady *prometheus.GaugeVec
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	metrics := &Metrics{
		registry: registry,
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "aipivot_http_requests_total",
			Help: "Total number of HTTP requests.",
		}, []string{"method", "path", "status"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "aipivot_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path", "status"}),
		dependencyReady: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "aipivot_dependency_ready",
			Help: "Dependency readiness state, 1 for ready and 0 for not ready.",
		}, []string{"dependency"}),
	}

	registry.MustRegister(metrics.httpRequests, metrics.httpDuration, metrics.dependencyReady)

	return metrics
}

func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	statusValue := strconv.Itoa(status)
	m.httpRequests.WithLabelValues(method, path, statusValue).Inc()
	m.httpDuration.WithLabelValues(method, path, statusValue).Observe(duration.Seconds())
}

func (m *Metrics) SetDependencyReady(name string, ready bool) {
	value := 0.0
	if ready {
		value = 1.0
	}

	m.dependencyReady.WithLabelValues(name).Set(value)
}
