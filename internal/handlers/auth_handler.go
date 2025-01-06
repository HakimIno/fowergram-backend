package handlers

import (
	"fowergram/internal/core/domain"
	"fowergram/internal/core/ports"
	"fowergram/pkg/errors"

	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService ports.AuthService
	validate    *validator.Validate
}

func NewAuthHandler(as ports.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: as,
		validate:    validator.New(),
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := new(domain.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		fmt.Printf("BodyParser error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
		})
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password, // Service will hash this
	}

	if err := h.authService.Register(user); err != nil {
		fmt.Printf("Register error: %v\n", err)
		switch e := err.(type) {
		case *errors.AuthError:
			if e.Code == "AUTH003" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": e.Message,
					"code":  e.Code,
				})
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": e.Message,
				"code":  e.Code,
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to register user",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Registration successful",
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(domain.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
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
			if e.Code == "AUTH002" { // Account locked error
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Account is locked due to too many failed attempts",
					"code":  e.Code,
				})
			}
			if e.Code == "AUTH001" { // Invalid credentials
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid email or password",
					"code":  e.Code,
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": e.Message,
				"code":  e.Code,
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	deviceID := c.Get("Device-ID")
	user := c.Locals("user").(*domain.User)

	if err := h.authService.RevokeSession(user.ID, deviceID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to logout",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// ValidateToken handles token validation requests
func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
	// Token validation is handled by the auth middleware
	// If we reach here, the token is valid
	return c.SendStatus(fiber.StatusOK)
}
