package middleware

import (
	"fowergram/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func ValidateAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Remove "Bearer " prefix
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		user, err := security.ValidateToken(token, jwtSecret)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		c.Locals("user_id", user.ID)
		return c.Next()
	}
}
