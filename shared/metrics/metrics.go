package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	VehicleRegisteredTotal   prometheus.Counter
	VehicleQueriesTotal     prometheus.Counter
	VehicleErrorsTotal      prometheus.Counter

	TelemetryEventsProcessedTotal prometheus.Counter
	TelemetryEventsFailedTotal    prometheus.Counter
	TelemetryQueriesTotal         prometheus.Counter

	AlertsGeneratedTotal prometheus.Counter
	AlertsQueriedTotal   prometheus.Counter
}

var globalMetrics *Metrics

func NewMetrics() *Metrics {
	if globalMetrics != nil {
		return globalMetrics
	}

	globalMetrics = &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_server_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_server_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_server_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000},
			},
			[]string{"method", "path"},
		),
		VehicleRegisteredTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "vehicle_registered_total",
				Help: "Total number of vehicles registered",
			},
		),
		VehicleQueriesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "vehicle_queries_total",
				Help: "Total number of vehicle queries",
			},
		),
		VehicleErrorsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "vehicle_errors_total",
				Help: "Total number of vehicle operation errors",
			},
		),
		TelemetryEventsProcessedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "telemetry_events_processed_total",
				Help: "Total number of telemetry events processed",
			},
		),
		TelemetryEventsFailedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "telemetry_events_failed_total",
				Help: "Total number of telemetry events failed",
			},
		),
		TelemetryQueriesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "telemetry_queries_total",
				Help: "Total number of telemetry queries",
			},
		),
		AlertsGeneratedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "alerts_generated_total",
				Help: "Total number of alerts generated",
			},
		),
		AlertsQueriedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "alerts_queried_total",
				Help: "Total number of alert queries",
			},
		),
	}

	return globalMetrics
}