package routes

import (
	"fowergram/internal/handlers"
	"fowergram/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(api fiber.Router, userHandler *handlers.UserHandler, jwtSecret string) {
	users := api.Group("/users")

	// Add routes for checking username and email availability
	users.Get("/check/username", userHandler.CheckUsernameAvailability)
	users.Get("/check/email", userHandler.CheckEmailAvailability)

	// Protected routes (auth required)
	// Add a /me endpoint that uses auth middleware
	// สำคัญ: ต้องลงทะเบียนเส้นทางเฉพาะก่อนเส้นทางที่มีพารามิเตอร์เพื่อหลีกเลี่ยงการชนกัน
	userAuth := users.Group("/me")
	userAuth.Use(middleware.ValidateAuth(jwtSecret))
	userAuth.Get("/", userHandler.GetMe)
	userAuth.Post("/profile-picture", userHandler.UpdateProfilePicture)

	// ลงทะเบียนเส้นทางที่มีพารามิเตอร์หลังจากเส้นทางเฉพาะ
	users.Get("/:id", userHandler.GetUser)
	users.Get("/", userHandler.GetUsers)
}
