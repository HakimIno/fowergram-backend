package handlers

import (
	"fowergram/internal/core/domain"
	"fowergram/internal/core/ports"
	"fowergram/pkg/errors"
	"fowergram/pkg/response"

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
		return response.InvalidFormat(c, err, map[string]string{
			"username": "string",
			"email":    "string",
			"password": "string",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorDetails := make([]map[string]string, 0)
		for _, e := range validationErrors {
			errorDetails = append(errorDetails, map[string]string{
				"field": e.Field(),
				"tag":   e.Tag(),
				"value": e.Value().(string),
			})
		}
		return response.ValidationError(c, errorDetails)
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password,
	}

	if err := h.authService.Register(user); err != nil {
		switch e := err.(type) {
		case *errors.AuthError:
			if e.Code == "AUTH003" {
				return response.BadRequest(c, e.Code, e.Message, map[string]interface{}{
					"field":  "email",
					"reason": "already_exists",
				})
			}
			return response.BadRequest(c, e.Code, e.Message, nil)
		default:
			return response.InternalError(c)
		}
	}

	return response.Created(c, "REGISTRATION_SUCCESS", "Registration successful", map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(domain.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidFormat(c, err, map[string]string{
			"email":    "string",
			"password": "string",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorDetails := make([]map[string]string, 0)
		for _, e := range validationErrors {
			errorDetails = append(errorDetails, map[string]string{
				"field": e.Field(),
				"tag":   e.Tag(),
				"value": e.Value().(string),
			})
		}
		return response.ValidationError(c, errorDetails)
	}

	deviceInfo := &domain.DeviceSession{
		DeviceType: c.Get("User-Agent"),
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
	}

	user, token, err := h.authService.Login(req.Email, req.Password, deviceInfo)
	if err != nil {
		switch e := err.(type) {
		case *errors.AuthError:
			if e.Code == "AUTH002" {
				return response.Unauthorized(c, e.Code, "Account is locked due to too many failed attempts", map[string]interface{}{
					"locked": true,
				})
			}
			return response.Unauthorized(c, e.Code, e.Message, nil)
		default:
			return response.InternalError(c)
		}
	}

	return response.Success(c, "LOGIN_SUCCESS", "Login successful", map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
		"token":       token,
		"device_info": deviceInfo,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	deviceID := c.Get("Device-ID")
	user := c.Locals("user").(*domain.User)

	if err := h.authService.RevokeSession(user.ID, deviceID); err != nil {
		return response.InternalError(c)
	}

	return response.Success(c, "LOGOUT_SUCCESS", "Logged out successfully", nil)
}

func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
	return response.Success(c, "TOKEN_VALID", "Token is valid", nil)
}
