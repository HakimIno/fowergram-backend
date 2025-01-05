package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	authFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_failures_total",
			Help: "Total number of authentication failures",
		},
		[]string{"reason"},
	)
)

// MonitoringMiddleware contains monitoring-related middleware
type MonitoringMiddleware struct {
	logger *zap.Logger
}

// NewMonitoringMiddleware creates a new monitoring middleware
func NewMonitoringMiddleware(logger *zap.Logger) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		logger: logger,
	}
}

// RequestLogger returns request logging middleware
func (m *MonitoringMiddleware) RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()

		// Update Prometheus metrics
		httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)

		// Log request
		m.logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Float64("duration", duration),
			zap.String("ip", c.IP()),
			zap.String("user-agent", c.Get("User-Agent")),
		)

		return err
	}
}

// AuthFailureLogger logs authentication failures
func (m *MonitoringMiddleware) AuthFailureLogger(reason string) {
	authFailures.WithLabelValues(reason).Inc()
	m.logger.Warn("Authentication failure",
		zap.String("reason", reason),
		zap.String("ip", ""), // Add IP in actual implementation
	)
}

// MetricsHandler returns Prometheus metrics handler
func (m *MonitoringMiddleware) MetricsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Implementation for exposing Prometheus metrics
		return nil // TODO: Implement proper metrics handler
	}
}
