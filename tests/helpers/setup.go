package helpers

import (
	"fowergram/internal/interfaces/http/routes"

	"github.com/gofiber/fiber/v2"
)

// SetupTestApp creates a new Fiber app instance for testing
func SetupTestApp() *fiber.App {
	app := fiber.New()

	// Register all routes
	routes.HealthRoutes(app)
	routes.AuthRoutes(app)

	return app
}
