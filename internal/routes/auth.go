package routes

import (
	"fowergram/internal/handlers"
	"fowergram/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(api fiber.Router, authHandler *handlers.AuthHandler, jwtSecret string) {
	auth := api.Group("/auth")

	// Public routes (no authentication required)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh-token", authHandler.RefreshToken)

	// Protected routes (require authentication)
	authProtected := auth.Group("/")
	authProtected.Use(middleware.ValidateAuth(jwtSecret))
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Post("/switch-account", authHandler.SwitchAccount)
}
