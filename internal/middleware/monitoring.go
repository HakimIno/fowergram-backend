package middleware

import (
	"fmt"
	"time"

	"fowergram/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type RequestMetrics struct {
	Method      string
	Path        string
	Status      int
	Latency     time.Duration
	IP          string
	RequestID   string
	UserAgent   string
	QueryParams map[string]string
}

func RequestMonitoring(log *logger.ZerologService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Store request ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		c.Locals("requestID", requestID)

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Collect query params
		queryParams := make(map[string]string)
		c.Context().QueryArgs().VisitAll(func(key, value []byte) {
			queryParams[string(key)] = string(value)
		})

		// Create metrics
		metrics := RequestMetrics{
			Method:      c.Method(),
			Path:        c.Path(),
			Status:      c.Response().StatusCode(),
			Latency:     duration,
			IP:          c.IP(),
			RequestID:   requestID,
			UserAgent:   c.Get("User-Agent"),
			QueryParams: queryParams,
		}

		// Log request details
		logRequestMetrics(log, metrics)

		return err
	}
}

func logRequestMetrics(log *logger.ZerologService, metrics RequestMetrics) {
	msg := "HTTP Request"
	if metrics.Status >= 400 {
		msg = fmt.Sprintf("HTTP %d Error", metrics.Status)
	}

	fields := []logger.Field{
		logger.NewField("Request ID", metrics.RequestID),
		logger.NewField("Method", metrics.Method),
		logger.NewField("Status", metrics.Status),
		logger.NewField("Latency", fmt.Sprintf("%.2fms", float64(metrics.Latency.Microseconds())/1000)),
		logger.NewField("Path", metrics.Path),
		logger.NewField("IP", metrics.IP),
	}

	switch {
	case metrics.Status >= 500:
		log.Error(msg, fmt.Errorf("status code: %d", metrics.Status), fields...)
	case metrics.Status >= 400:
		log.Warn(msg, fields...)
	default:
		log.Info(msg, fields...)
	}
}
