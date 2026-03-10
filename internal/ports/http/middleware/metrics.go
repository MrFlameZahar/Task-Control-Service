package middleware

import (
	"net/http"
	"strconv"
	"time"

	"TaskControlService/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func Metrics(m *metrics.HTTPMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := newStatusRecorder(w)

			next.ServeHTTP(recorder, r)

			status := strconv.Itoa(recorder.statusCode)
			method := r.Method
			path := r.URL.Path

			m.RequestsTotal.WithLabelValues(method, path, status).Inc()
			m.RequestDuration.WithLabelValues(method, path, status).Observe(time.Since(start).Seconds())

			if recorder.statusCode >= 400 {
				m.ErrorsTotal.WithLabelValues(method, path, status).Inc()
			}
		})
	}
}