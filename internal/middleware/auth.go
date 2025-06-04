package middleware

import (
	"fowergram/internal/domain"
	"fowergram/pkg/security"

	"github.com/gofiber/fiber/v2"
)

// Sanitize token before logging
func maskToken(token string) string {
	if len(token) > 10 {
		return token[:10] + "..."
	}
	return "****"
}

func ValidateAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")
		isRefreshToken := false

		if authHeader == "" || authHeader == "null" {
			refreshToken := c.Get("X-Refresh-Token")
			if refreshToken != "" {
				authHeader = "Bearer " + refreshToken
				isRefreshToken = true
			}
		}

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
		var user *domain.User
		var err error

		// Choose validation method based on token type
		if isRefreshToken || c.Path() == "/api/v1/auth/switch-account" {
			user, err = security.ValidateRefreshTokenAsAccessToken(token, jwtSecret)
		} else {
			user, err = security.ValidateToken(token, jwtSecret)
		}

		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authentication error: Please refresh your session or log in again.",
			})
		}

		c.Locals("user_id", user.ID)
		return c.Next()
	}
}
