package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path"})

	httpResponsesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_responses_total",
		Help: "Total number of HTTP responses",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})

	httpPanicsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_panics_total",
		Help: "Total number of panics occurred",
	}, []string{"method", "path"})

	httpRequestSizeBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_size_bytes",
		Help:    "Size of HTTP request bodies in bytes",
		Buckets: []float64{100, 1000, 10000, 100000, 1e6, 5e6},
	}, []string{"method", "path"})

	httpResponseSizeBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_size_bytes",
		Help:    "Size of HTTP response bodies in bytes",
		Buckets: []float64{100, 1000, 10000, 100000, 1e6, 5e6},
	}, []string{"method", "path"})
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Count incoming request
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

		// Record request size
		if r.ContentLength > 0 {
			httpRequestSizeBytes.WithLabelValues(r.Method, r.URL.Path).Observe(float64(r.ContentLength))
		}

		start := time.Now()
		rw := &responseWriter{
			ResponseWriter: w,
			status:         0,
			size:           0,
		}

		defer func() {
			if err := recover(); err != nil {
				httpPanicsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
				panic(err)
			}
		}()

		next.ServeHTTP(rw, r)

		// Count outgoing response
		status := http.StatusText(rw.status)
		httpResponsesTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()

		// Record response size
		if rw.size > 0 {
			httpResponseSizeBytes.WithLabelValues(r.Method, r.URL.Path).Observe(float64(rw.size))
		}

		// Record duration
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func Handler() http.Handler {
	return promhttp.Handler()
}
