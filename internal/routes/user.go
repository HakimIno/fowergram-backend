package routes

import (
	"fowergram/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(api fiber.Router, userHandler *handlers.UserHandler) {
	users := api.Group("/users")
	users.Get("/:id", userHandler.GetUser)
	users.Get("/", userHandler.GetUsers)

	// Add routes for checking username and email availability
	users.Get("/check/username", userHandler.CheckUsernameAvailability)
	users.Get("/check/email", userHandler.CheckEmailAvailability)
}
