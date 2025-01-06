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
