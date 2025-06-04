package handlers

import (
	"fmt"
	"fowergram/internal/core/ports"
	"fowergram/internal/domain"
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
	// Don't log the entire request body which contains credentials
	h.log.Info("Login request received",
		logger.NewField("ip", c.IP()),
		logger.NewField("user_agent", c.Get("User-Agent")),
	)

	req := new(domain.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		h.log.Warn("Error parsing login request",
			logger.NewField("error", err.Error()),
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

	user, token, refreshToken, err := h.authService.Login(req.Identifier, req.Password, deviceInfo)
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

	if user.ProfilePicture != "" {
		userResponse["profile_picture"] = user.ProfilePicture
	}

	return response.Success(c, "LOGIN_SUCCESS", "Login successful", map[string]interface{}{
		"user":          userResponse,
		"token":         token,
		"refresh_token": refreshToken,
		"device_info":   deviceInfo,
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

func (h *AuthHandler) SwitchAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	authHeader := c.Get("Authorization")
	if authHeader == "" || authHeader == "null" {
		refreshToken := c.Get("X-Refresh-Token")
		if refreshToken != "" {
			c.Request().Header.Set("Authorization", "Bearer "+refreshToken)
			authHeader = "Bearer " + refreshToken
		}
	}

	if userID == nil {
		return response.Unauthorized(c, "AUTH006", "Authentication required", nil)
	}

	var currentUserID uint
	switch v := userID.(type) {
	case uint:
		currentUserID = v
	case int:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		h.log.Error("Invalid user ID type", nil,
			logger.NewField("user_id_type", fmt.Sprintf("%T", userID)),
		)
		return response.Unauthorized(c, "AUTH006", "Invalid user ID", nil)
	}

	// Parse the request body
	req := new(domain.SwitchAccountRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidFormat(c, err, map[string]string{
			"switch_type":  "string ('token' or 'password')",
			"identifier":   "string (email or username)",
			"password":     "string (required for password type)",
			"stored_token": "string (required for token type)",
		})
	}

	// Validate request
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

	// Validate switch_type specific fields
	if req.SwitchType == "password" && req.Password == "" {
		return response.BadRequest(c, "AUTH009", "Password is required for password type switch", nil)
	}

	if req.SwitchType == "token" && req.StoredToken == "" {
		return response.BadRequest(c, "AUTH010", "Stored token is required for token type switch", nil)
	}

	// Get device info
	deviceInfo := &domain.DeviceSession{
		DeviceType: c.Get("User-Agent"),
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
	}

	// Get device ID from header or create a new one
	deviceID := c.Get("Device-ID")
	if deviceID != "" {
		deviceInfo.DeviceID = deviceID
	}

	// Switch to the target account
	user, token, refreshToken, err := h.authService.SwitchAccount(currentUserID, req, deviceInfo)
	if err != nil {
		switch e := err.(type) {
		case *errors.AuthError:
			switch e.Code {
			case "AUTH002":
				return response.Unauthorized(c, e.Code, "Account is locked due to too many failed attempts", map[string]interface{}{
					"locked": true,
				})
			case "AUTH007":
				return response.Unauthorized(c, e.Code, e.Message, map[string]interface{}{
					"require_password": true,
				})
			case "AUTH008":
				return response.BadRequest(c, e.Code, e.Message, nil)
			default:
				return response.Unauthorized(c, e.Code, e.Message, nil)
			}
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

	if user.ProfilePicture != "" {
		userResponse["profile_picture"] = user.ProfilePicture
	}

	return response.Success(c, "SWITCH_ACCOUNT_SUCCESS", "Successfully switched accounts", map[string]interface{}{
		"user":          userResponse,
		"token":         token,
		"refresh_token": refreshToken,
		"device_info":   deviceInfo,
	})
}

// RefreshToken refreshes the access token using a refresh token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Parse the request body
	type RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	req := new(RefreshTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidFormat(c, err, map[string]string{
			"refresh_token": "string",
		})
	}

	// Validate request
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

	// Call service to refresh the token
	accessToken, refreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.log.Error("Error refreshing token", err,
			logger.NewField("error", err.Error()),
		)

		if err == errors.ErrInvalidRefreshToken {
			return response.Unauthorized(c, "AUTH011", "Invalid or expired refresh token", nil)
		}

		if err == errors.ErrUserNotFound {
			return response.Unauthorized(c, "AUTH012", "User not found", nil)
		}

		return response.InternalError(c)
	}

	return response.Success(c, "TOKEN_REFRESHED", "Token refreshed successfully", map[string]interface{}{
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}
