package handlers

import (
	"fowergram/internal/core/domain"
	"fowergram/internal/core/ports"
	"fowergram/pkg/errors"
	"fowergram/pkg/security"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService ports.AuthService
}

func NewAuthHandler(as ports.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: as,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := new(domain.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := security.ValidatePassword(req.Password); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password, // จะถูก hash ใน service
	}

	if err := h.authService.Register(user); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to register user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Registration successful",
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create device info from request
	deviceInfo := &domain.DeviceSession{
		DeviceType: c.Get("User-Agent"),
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
	}

	user, token, err := h.authService.Login(req.Email, req.Password, deviceInfo)
	if err != nil {
		switch e := err.(type) {
		case *errors.AuthError:
			return c.Status(401).JSON(fiber.Map{
				"error": e.Message,
				"code":  e.Code,
			})
		default:
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	deviceID := c.Get("Device-ID")
	user := c.Locals("user").(*domain.User)

	if err := h.authService.RevokeSession(user.ID, deviceID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to logout",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
