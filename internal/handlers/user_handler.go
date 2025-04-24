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

// GetMe retrieves the currently authenticated user's profile
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	startTime := time.Now()

	// Try to get user_id from context in a safer way
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized - missing user ID",
		})
	}

	// Handle different numeric types
	var userID uint
	switch v := userIDRaw.(type) {
	case uint:
		userID = v
	case int:
		userID = uint(v)
	case float64:
		userID = uint(v)
	case int64:
		userID = uint(v)
	default:
		fmt.Printf("Unexpected type for user_id: %T\n", userIDRaw)
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized - invalid user ID format",
		})
	}

	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized - zero user ID",
		})
	}

	fmt.Printf("Processing GetMe for user ID: %d\n", userID)

	// Get user from service (it already handles caching internally)
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		fmt.Printf("Error getting user: %v\n", err)
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	totalTime := time.Since(startTime)
	fmt.Printf("GetMe took: %v\n", totalTime)

	return c.JSON(user)
}

// UpdateProfilePicture updates the current user's profile picture
func (h *UserHandler) UpdateProfilePicture(c *fiber.Ctx) error {
	startTime := time.Now()

	// Try to get user_id from context in a safer way
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized - missing user ID",
		})
	}

	// Handle different numeric types
	var userID uint
	switch v := userIDRaw.(type) {
	case uint:
		userID = v
	case int:
		userID = uint(v)
	case float64:
		userID = uint(v)
	case int64:
		userID = uint(v)
	default:
		fmt.Printf("Unexpected type for user_id: %T\n", userIDRaw)
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized - invalid user ID format",
		})
	}

	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized - zero user ID",
		})
	}

	fmt.Printf("Processing UpdateProfilePicture for user ID: %d\n", userID)

	// Parse request body
	type ProfilePictureRequest struct {
		ProfilePicture string `json:"profile_picture" validate:"required"`
	}

	req := new(ProfilePictureRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate request
	if req.ProfilePicture == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Profile picture URL is required",
		})
	}

	// Update profile picture
	if err := h.userService.UpdateProfilePicture(userID, req.ProfilePicture); err != nil {
		fmt.Printf("Error updating profile picture: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update profile picture",
		})
	}

	totalTime := time.Since(startTime)
	fmt.Printf("UpdateProfilePicture took: %v\n", totalTime)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile picture updated successfully",
	})
}
