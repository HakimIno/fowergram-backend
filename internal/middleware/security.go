package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// SecurityMiddleware contains security-related middleware configurations
type SecurityMiddleware struct{}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{}
}

// RateLimiter returns rate limiting middleware
func (m *SecurityMiddleware) RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,               // 5 requests
		Expiration: 1 * time.Minute, // per 1 minute
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address and path as key
			return c.IP() + ":" + c.Path()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many attempts, please try again later",
			})
		},
		SkipFailedRequests: false, // Count failed requests towards the limit
	})
}

// CORS returns CORS middleware
func (m *SecurityMiddleware) CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, https://api.fowergram.online/metrics",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
		MaxAge:           int(12 * time.Hour.Seconds()),
	})
}

// SecurityHeaders returns security headers middleware
func (m *SecurityMiddleware) SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set security headers
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("Content-Security-Policy", "default-src 'self'")
		return c.Next()
	}
}
