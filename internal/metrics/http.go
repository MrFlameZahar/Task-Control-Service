package metrics

import "github.com/prometheus/client_golang/prometheus"

type HTTPMetrics struct {
	RequestsTotal   *prometheus.CounterVec
	ErrorsTotal     *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

func NewHTTPMetrics() *HTTPMetrics {
	m := &HTTPMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP error responses",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
	}

	prometheus.MustRegister(
		m.RequestsTotal,
		m.ErrorsTotal,
		m.RequestDuration,
	)

	return m
}