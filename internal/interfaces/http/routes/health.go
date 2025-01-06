package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func RegisterHealthRoutes(app *fiber.App) {
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})
}
