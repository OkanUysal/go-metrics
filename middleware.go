package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware returns a Gin middleware that collects HTTP metrics
func (m *Metrics) Middleware() gin.HandlerFunc {
	if m.httpMetrics == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
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
		status := strconv.Itoa(c.Writer.Status())
		
		// Record metrics
		m.httpMetrics.RequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()
		
		m.httpMetrics.RequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Observe(duration)
		
		// Record response size
		m.httpMetrics.ResponseSize.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(float64(c.Writer.Size()))
	}
}

// MiddlewareWithSkipper returns a middleware with a skipper function
func (m *Metrics) MiddlewareWithSkipper(skipper func(*gin.Context) bool) gin.HandlerFunc {
	middleware := m.Middleware()
	
	return func(c *gin.Context) {
		if skipper(c) {
			c.Next()
			return
		}
		middleware(c)
	}
}
