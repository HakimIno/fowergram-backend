package routes

import (
	"github.com/gofiber/fiber/v2"
)

// AuthRoutes registers all authentication related routes
func AuthRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	// TODO: Implement auth routes
	auth.Post("/login", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	auth.Post("/register", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})
}
