package routes

import (
	"fowergram/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(api fiber.Router, authHandler *handlers.AuthHandler) {
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
}
