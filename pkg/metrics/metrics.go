package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"service", "method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path"},
	)

	// Health Check Metrics
	HealthChecksTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "health_checks_total",
			Help: "Total health checks performed",
		},
		[]string{"service_name", "result"}, // result: healthy, down, error
	)

	ServicesMonitored = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "services_monitored_total",
			Help: "Current count of registered services being monitored",
		},
	)

	// Kafka Metrics
	KafkaEventsPublished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_events_published_total",
			Help: "Total number of service events published to Kafka",
		},
		[]string{"event_type"},
	)

	KafkaEventsConsumed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_events_consumed_total",
			Help: "Total number of service events consumed from Kafka",
		},
		[]string{"event_type"},
	)

	// Incident Metrics
	IncidentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "incidents_total",
			Help: "Total number of incidents created or resolved",
		},
		[]string{"service_name", "action"}, // action: created, resolved, duplicate
	)

	// Notification & gRPC Metrics
	NotificationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_total",
			Help: "Total notifications processed",
		},
		[]string{"channel", "status"}, // status: sent, failed
	)

	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total gRPC requests processed",
		},
		[]string{"method", "status"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HealthChecksTotal,
		ServicesMonitored,
		KafkaEventsPublished,
		KafkaEventsConsumed,
		IncidentsTotal,
		NotificationsTotal,
		GRPCRequestsTotal,
	)
}

// Handler returns the HTTP handler for Prometheus metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}

// Middleware instruments HTTP handlers with request count and latency metrics.
func Middleware(serviceName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		HTTPRequestsTotal.WithLabelValues(serviceName, r.Method, r.URL.Path, status).Inc()
		HTTPRequestDuration.WithLabelValues(serviceName, r.Method, r.URL.Path).Observe(duration)
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
