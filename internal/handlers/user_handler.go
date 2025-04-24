package handlers

import (
	"fmt"
	"fowergram/internal/core/ports"
	"time"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService ports.UserService
}

func NewUserHandler(us ports.UserService) *UserHandler {
	return &UserHandler{
		userService: us,
	}
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	startTime := time.Now()

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	totalTime := time.Since(startTime)
	fmt.Printf("GetUser took: %v\n", totalTime)

	return c.JSON(user)
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get pagination params
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// Try to get from cache first
	cacheKey := fmt.Sprintf("users:page:%d:limit:%d", page, limit)
	users, err := h.userService.GetUsersFromCache(cacheKey)
	if err == nil {
		totalTime := time.Since(startTime)
		fmt.Printf("GetUsers from cache took: %v\n", totalTime)
		return c.JSON(users)
	}

	// If not in cache, get from database
	users, err = h.userService.GetUsers(page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get users",
		})
	}

	// Cache the result async
	go func() {
		if err := h.userService.CacheUsers(cacheKey, users); err != nil {
			fmt.Printf("failed to cache users: %v\n", err)
		}
	}()

	totalTime := time.Since(startTime)
	fmt.Printf("GetUsers from DB took: %v\n", totalTime)

	return c.JSON(users)
}

// CheckUsernameAvailability checks if a username is available
func (h *UserHandler) CheckUsernameAvailability(c *fiber.Ctx) error {
	startTime := time.Now()

	username := c.Query("username")
	if username == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Username is required",
		})
	}

	// Try to get the user with this username
	user, err := h.userService.GetUserByUsername(username)

	// If user is found, username is taken
	if err == nil && user != nil {
		totalTime := time.Since(startTime)
		fmt.Printf("CheckUsernameAvailability took: %v\n", totalTime)

		return c.JSON(fiber.Map{
			"available": false,
		})
	}

	// If error is not "record not found", it's a server error
	if err != nil && err.Error() != "record not found" {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check username availability",
		})
	}

	// Otherwise, username is available
	totalTime := time.Since(startTime)
	fmt.Printf("CheckUsernameAvailability took: %v\n", totalTime)

	return c.JSON(fiber.Map{
		"available": true,
	})
}

// CheckEmailAvailability checks if an email is available
func (h *UserHandler) CheckEmailAvailability(c *fiber.Ctx) error {
	startTime := time.Now()

	email := c.Query("email")
	if email == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	// Try to get the user with this email
	user, err := h.userService.GetUserByEmail(email)

	// If user is found, email is taken
	if err == nil && user != nil {
		totalTime := time.Since(startTime)
		fmt.Printf("CheckEmailAvailability took: %v\n", totalTime)

		return c.JSON(fiber.Map{
			"available": false,
		})
	}

	// If error is not "record not found", it's a server error
	if err != nil && err.Error() != "record not found" {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check email availability",
		})
	}

	// Otherwise, email is available
	totalTime := time.Since(startTime)
	fmt.Printf("CheckEmailAvailability took: %v\n", totalTime)

	return c.JSON(fiber.Map{
		"available": true,
	})
}
