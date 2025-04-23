package handlers

import (
	"fowergram/internal/core/domain"
	"fowergram/internal/core/ports"
	"fowergram/pkg/errors"
	"fowergram/pkg/logger"
	"fowergram/pkg/response"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService ports.AuthService
	validate    *validator.Validate
	log         *logger.ZerologService
}

func NewAuthHandler(as ports.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: as,
		validate:    validator.New(),
		log:         logger.NewLogger(logger.InfoLevel),
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := new(domain.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidFormat(c, err, map[string]string{
			"username":   "string",
			"email":      "string (optional)",
			"password":   "string",
			"birth_date": "string (optional, format: YYYY-MM-DD)",
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
		PasswordHash: req.Password,
	}

	// Set email if provided
	if req.Email != "" {
		user.Email = req.Email
	}

	// Parse birth date if provided
	if req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			return response.BadRequest(c, "AUTH005", "Invalid birth date format", map[string]interface{}{
				"field":  "birth_date",
				"reason": "invalid_format",
			})
		}

		// Check if birth date is in the future
		if birthDate.After(time.Now()) {
			return response.BadRequest(c, "AUTH006", "Birth date cannot be in the future", map[string]interface{}{
				"field":  "birth_date",
				"reason": "future_date",
			})
		}

		user.BirthDate = &birthDate
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

	// Create a response object that includes birth_date if it exists
	userResponse := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	}

	if user.Email != "" {
		userResponse["email"] = user.Email
	}

	if user.BirthDate != nil {
		userResponse["birth_date"] = user.BirthDate.Format("2006-01-02")
	}

	return response.Created(c, "REGISTRATION_SUCCESS", "Registration successful", map[string]interface{}{
		"user": userResponse,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Log request body
	requestBody := string(c.Body())
	h.log.Info("Login request received",
		logger.NewField("body", requestBody),
		logger.NewField("ip", c.IP()),
		logger.NewField("user_agent", c.Get("User-Agent")),
	)

	req := new(domain.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		h.log.Warn("Error parsing login request",
			logger.NewField("error", err.Error()),
			logger.NewField("body", requestBody),
		)
		return response.InvalidFormat(c, err, map[string]string{
			"identifier": "string (email or username)",
			"password":   "string",
		})
	}

	h.log.Info("Login attempt",
		logger.NewField("identifier", req.Identifier),
	)

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

	user, token, err := h.authService.Login(req.Identifier, req.Password, deviceInfo)
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

	// Create user response
	userResponse := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	}

	if user.Email != "" {
		userResponse["email"] = user.Email
	}

	return response.Success(c, "LOGIN_SUCCESS", "Login successful", map[string]interface{}{
		"user":        userResponse,
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
