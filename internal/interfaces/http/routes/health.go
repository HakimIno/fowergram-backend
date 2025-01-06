package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthRoutes handles health check endpoints
func HealthRoutes(app *fiber.App) {
	// Simple ping endpoint for health checks
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// Detailed health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		// TODO: Add more detailed health checks (DB, Redis, etc.)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
			"services": fiber.Map{
				"api":   "up",
				"db":    "up",
				"redis": "up",
			},
		})
	})
}
