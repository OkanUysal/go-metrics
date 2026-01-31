package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Config contains metrics configuration
type Config struct {
	ServiceName string // Service name for metrics
	Namespace   string // Prometheus namespace (e.g., "outcome")
	Subsystem   string // Prometheus subsystem (optional)

	// HTTP metrics configuration
	EnableHTTPMetrics     bool
	HTTPBuckets           []float64 // Custom histogram buckets for HTTP duration
	EnableMetricsEndpoint bool      // Auto-register /metrics endpoint
	EnableHealthEndpoint  bool      // Auto-register /health endpoint

	// Push gateway configuration (optional)
	PushGatewayURL string
	PushInterval   time.Duration

	// Grafana Cloud configuration (optional)
	GrafanaCloudURL    string
	GrafanaCloudUser   string
	GrafanaCloudAPIKey string

	// Custom labels for all metrics
	ConstLabels prometheus.Labels
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		ServiceName:           "app",
		Namespace:             "app",
		EnableHTTPMetrics:     true,
		EnableMetricsEndpoint: true,
		EnableHealthEndpoint:  true,
		HTTPBuckets:           []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		PushInterval:          15 * time.Second,
		ConstLabels:           prometheus.Labels{},
	}
}

// HTTPMetrics contains HTTP-related metrics
type HTTPMetrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestSize      *prometheus.HistogramVec
	ResponseSize     *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge
}

// Labels contains common label keys
type Labels struct {
	Method     string
	Path       string
	StatusCode string
	Error      string
}

// MetricLabels is a map of label key-value pairs
type MetricLabels map[string]string
