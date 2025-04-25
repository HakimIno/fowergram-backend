package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing Authorization header")
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid Authorization header format")
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Set user ID in context
			c.Locals("user_id", claims["user_id"])
			return c.Next()
		}

		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token claims")
	}
}
