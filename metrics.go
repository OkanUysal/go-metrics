package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics is the main metrics collector
type Metrics struct {
	config   *Config
	registry *prometheus.Registry

	// HTTP metrics
	httpMetrics *HTTPMetrics

	// Custom metrics storage
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec

	mu sync.RWMutex
}

// NewMetrics creates a new metrics collector
func NewMetrics(config *Config) *Metrics {
	if config == nil {
		config = DefaultConfig()
	} else {
		// Apply defaults for unset fields
		if config.Namespace == "" {
			config.Namespace = "app"
		}
		if config.HTTPBuckets == nil {
			config.HTTPBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
		}
		if config.PushInterval == 0 {
			config.PushInterval = 15 * time.Second
		}
		// Enable by default if not explicitly set
		if !config.EnableHTTPMetrics && config.ServiceName != "" {
			config.EnableHTTPMetrics = true
		}
		if !config.EnableMetricsEndpoint && config.ServiceName != "" {
			config.EnableMetricsEndpoint = true
		}
		if !config.EnableHealthEndpoint && config.ServiceName != "" {
			config.EnableHealthEndpoint = true
		}
	}

	registry := prometheus.NewRegistry()

	m := &Metrics{
		config:     config,
		registry:   registry,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}

	// Initialize HTTP metrics if enabled
	if config.EnableHTTPMetrics {
		m.initHTTPMetrics()
	}

	return m
}

// initHTTPMetrics initializes HTTP-related metrics
func (m *Metrics) initHTTPMetrics() {
	m.httpMetrics = &HTTPMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   m.config.Namespace,
				Subsystem:   m.config.Subsystem,
				Name:        "http_requests_total",
				Help:        "Total number of HTTP requests",
				ConstLabels: m.config.ConstLabels,
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   m.config.Namespace,
				Subsystem:   m.config.Subsystem,
				Name:        "http_request_duration_seconds",
				Help:        "HTTP request duration in seconds",
				Buckets:     m.config.HTTPBuckets,
				ConstLabels: m.config.ConstLabels,
			},
			[]string{"method", "path", "status"},
		),
		RequestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   m.config.Namespace,
				Subsystem:   m.config.Subsystem,
				Name:        "http_request_size_bytes",
				Help:        "HTTP request size in bytes",
				Buckets:     prometheus.ExponentialBuckets(100, 10, 7),
				ConstLabels: m.config.ConstLabels,
			},
			[]string{"method", "path"},
		),
		ResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   m.config.Namespace,
				Subsystem:   m.config.Subsystem,
				Name:        "http_response_size_bytes",
				Help:        "HTTP response size in bytes",
				Buckets:     prometheus.ExponentialBuckets(100, 10, 7),
				ConstLabels: m.config.ConstLabels,
			},
			[]string{"method", "path"},
		),
		RequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   m.config.Namespace,
				Subsystem:   m.config.Subsystem,
				Name:        "http_requests_in_flight",
				Help:        "Current number of HTTP requests being processed",
				ConstLabels: m.config.ConstLabels,
			},
		),
	}

	// Register HTTP metrics
	m.registry.MustRegister(
		m.httpMetrics.RequestsTotal,
		m.httpMetrics.RequestDuration,
		m.httpMetrics.RequestSize,
		m.httpMetrics.ResponseSize,
		m.httpMetrics.RequestsInFlight,
	)
}

// IncrementCounter increments a counter metric
func (m *Metrics) IncrementCounter(name string, labels MetricLabels) {
	m.IncrementCounterBy(name, 1, labels)
}

// IncrementCounterBy increments a counter by a specific value
func (m *Metrics) IncrementCounterBy(name string, value float64, labels MetricLabels) {
	counter := m.getOrCreateCounter(name, getLabelKeys(labels))
	counter.With(prometheus.Labels(labels)).Add(value)
}

// SetGauge sets a gauge metric value
func (m *Metrics) SetGauge(name string, value float64, labels MetricLabels) {
	gauge := m.getOrCreateGauge(name, getLabelKeys(labels))
	gauge.With(prometheus.Labels(labels)).Set(value)
}

// IncrementGauge increments a gauge metric
func (m *Metrics) IncrementGauge(name string, labels MetricLabels) {
	gauge := m.getOrCreateGauge(name, getLabelKeys(labels))
	gauge.With(prometheus.Labels(labels)).Inc()
}

// DecrementGauge decrements a gauge metric
func (m *Metrics) DecrementGauge(name string, labels MetricLabels) {
	gauge := m.getOrCreateGauge(name, getLabelKeys(labels))
	gauge.With(prometheus.Labels(labels)).Dec()
}

// RecordHistogram records a histogram observation
func (m *Metrics) RecordHistogram(name string, value float64, labels MetricLabels) {
	histogram := m.getOrCreateHistogram(name, getLabelKeys(labels))
	histogram.With(prometheus.Labels(labels)).Observe(value)
}

// getOrCreateCounter gets or creates a counter metric
func (m *Metrics) getOrCreateCounter(name string, labelKeys []string) *prometheus.CounterVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if counter, exists := m.counters[name]; exists {
		return counter
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   m.config.Namespace,
			Subsystem:   m.config.Subsystem,
			Name:        name,
			Help:        name + " counter",
			ConstLabels: m.config.ConstLabels,
		},
		labelKeys,
	)

	m.registry.MustRegister(counter)
	m.counters[name] = counter

	return counter
}

// getOrCreateGauge gets or creates a gauge metric
func (m *Metrics) getOrCreateGauge(name string, labelKeys []string) *prometheus.GaugeVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if gauge, exists := m.gauges[name]; exists {
		return gauge
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   m.config.Namespace,
			Subsystem:   m.config.Subsystem,
			Name:        name,
			Help:        name + " gauge",
			ConstLabels: m.config.ConstLabels,
		},
		labelKeys,
	)

	m.registry.MustRegister(gauge)
	m.gauges[name] = gauge

	return gauge
}

// getOrCreateHistogram gets or creates a histogram metric
func (m *Metrics) getOrCreateHistogram(name string, labelKeys []string) *prometheus.HistogramVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if histogram, exists := m.histograms[name]; exists {
		return histogram
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   m.config.Namespace,
			Subsystem:   m.config.Subsystem,
			Name:        name,
			Help:        name + " histogram",
			Buckets:     prometheus.DefBuckets,
			ConstLabels: m.config.ConstLabels,
		},
		labelKeys,
	)

	m.registry.MustRegister(histogram)
	m.histograms[name] = histogram

	return histogram
}

// Handler returns the Prometheus HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Registry returns the Prometheus registry
func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

// getLabelKeys extracts label keys from a label map
func getLabelKeys(labels MetricLabels) []string {
	if labels == nil {
		return []string{}
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	return keys
}
