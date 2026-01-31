package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GinMiddleware returns a Gin middleware for automatic metrics collection
func (m *Metrics) GinMiddleware() gin.HandlerFunc {
	if !m.config.EnableHTTPMetrics {
		// Return a no-op middleware if HTTP metrics are disabled
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Skip metrics endpoint itself
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		// Increment in-flight requests
		m.httpMetrics.RequestsInFlight.Inc()
		defer m.httpMetrics.RequestsInFlight.Dec()

		// Record request size
		if c.Request.ContentLength > 0 {
			m.httpMetrics.RequestSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		status := c.Writer.Status()

		// Record metrics
		labels := []string{c.Request.Method, c.FullPath(), http.StatusText(status)}
		
		m.httpMetrics.RequestsTotal.WithLabelValues(labels...).Inc()
		m.httpMetrics.RequestDuration.WithLabelValues(labels...).Observe(duration)

		// Record response size
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			m.httpMetrics.ResponseSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(responseSize))
		}
	}
}

// MetricsEndpoint returns a Gin handler for the /metrics endpoint
func (m *Metrics) MetricsEndpoint() gin.HandlerFunc {
	handler := m.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
