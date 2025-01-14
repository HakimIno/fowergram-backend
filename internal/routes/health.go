package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupHealthRoutes(app *fiber.App) {
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
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
