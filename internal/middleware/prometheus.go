package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

func Prometheus() fiber.Handler {
	return monitor.New(monitor.Config{
		Title: "Fowergram Metrics",
	})
}
