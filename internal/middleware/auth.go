package middleware

import (
	"fowergram/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		user, err := security.ValidateToken(token, jwtSecret)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}
