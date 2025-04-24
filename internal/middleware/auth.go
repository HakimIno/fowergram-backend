package middleware

import (
	"fmt"
	"fowergram/internal/core/domain"
	"fowergram/pkg/security"

	"github.com/gofiber/fiber/v2"
)

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
		} else {
			fmt.Printf("DEBUG [ValidateAuth] Authorization header missing 'Bearer ' prefix: %s\n", authHeader)
		}

		var user *domain.User
		var err error

		// Choose validation method based on token type
		if isRefreshToken || c.Path() == "/api/v1/auth/switch-account" {
			fmt.Printf("DEBUG [ValidateAuth] Validating token as either type: %s...\n", token[:10])
			// Try both token types for switch-account
			user, err = security.ValidateRefreshTokenAsAccessToken(token, jwtSecret)
		} else {
			fmt.Printf("DEBUG [ValidateAuth] Validating as access token: %s...\n", token[:10])
			user, err = security.ValidateToken(token, jwtSecret)
		}

		if err != nil {
			fmt.Printf("DEBUG [ValidateAuth] Token validation failed: %v\n", err)
			return c.Status(401).JSON(fiber.Map{
				"error": "Authentication error: Please refresh your session or log in again.",
			})
		}

		fmt.Printf("DEBUG [ValidateAuth] Token validated successfully for user ID: %d\n", user.ID)
		c.Locals("user_id", user.ID)
		return c.Next()
	}
}
