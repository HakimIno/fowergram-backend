package services

import (
	"fmt"
	"strings"
	"time"

	"fowergram/internal/core/domain"
	"fowergram/internal/core/ports"
	"fowergram/pkg/email"
	"fowergram/pkg/errors"
	"fowergram/pkg/geolocation"
	"fowergram/pkg/security"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	authRepo     ports.AuthRepository
	emailService email.Service
	geoService   geolocation.Service
	cacheRepo    ports.CacheRepository
	jwtSecret    string
}

func NewAuthService(ar ports.AuthRepository, es email.Service, gs geolocation.Service, cr ports.CacheRepository, secret string) ports.AuthService {
	return &authService{
		authRepo:     ar,
		emailService: es,
		geoService:   gs,
		cacheRepo:    cr,
		jwtSecret:    secret,
	}
}

func (s *authService) Register(user *domain.User) error {
	startTime := time.Now()

	// Hash password with lower cost for faster registration
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), security.HashCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)
	hashTime := time.Since(startTime)
	fmt.Printf("Password hashing took: %v\n", hashTime)

	// Create user
	createStart := time.Now()
	if err := s.authRepo.CreateUser(user); err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return &errors.AuthError{
				Code:    "AUTH003",
				Message: "Email or username already exists",
			}
		}
		return &errors.AuthError{
			Code:    "AUTH004",
			Message: "Failed to create user",
		}
	}
	createTime := time.Since(createStart)
	fmt.Printf("User creation took: %v\n", createTime)

	// Do all non-critical operations async
	go func() {
		// Cache user data
		cacheKey := fmt.Sprintf("user:%d", user.ID)
		if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
			fmt.Printf("failed to cache user data: %v\n", err)
		}

		// Generate verification code
		code, err := security.GenerateRandomCode(6)
		if err != nil {
			fmt.Printf("failed to generate verification code: %v\n", err)
			return
		}

		// Create auth code
		authCode := &domain.AuthCode{
			UserID:    user.ID,
			Code:      code,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := s.authRepo.CreateAuthCode(authCode); err != nil {
			fmt.Printf("failed to create auth code: %v\n", err)
			return
		}

		// Send verification email
		if err := s.emailService.SendVerificationEmail(user.Email, code); err != nil {
			fmt.Printf("failed to send verification email: %v\n", err)
		}
	}()

	totalTime := time.Since(startTime)
	fmt.Printf("Total registration took: %v\n", totalTime)

	return nil
}

func (s *authService) Login(email, password string, deviceInfo *domain.DeviceSession) (*domain.User, string, error) {
	startTime := time.Now()

	// Try to get user from cache first with shorter timeout
	cacheKey := fmt.Sprintf("user:email:%s", email)
	var user *domain.User
	cacheDone := make(chan bool, 1)

	go func() {
		if cached, err := s.cacheRepo.Get(cacheKey); err == nil {
			if userData, ok := cached.(*domain.User); ok {
				user = userData
			}
		}
		cacheDone <- true
	}()

	// Wait for cache with short timeout
	select {
	case <-cacheDone:
	case <-time.After(100 * time.Millisecond):
		fmt.Printf("Cache lookup timed out\n")
	}

	cacheTime := time.Since(startTime)
	fmt.Printf("Cache lookup took: %v\n", cacheTime)

	// If not in cache, get from database
	dbStart := time.Now()
	if user == nil {
		var err error
		user, err = s.authRepo.FindUserByEmail(email)
		if err != nil {
			return nil, "", &errors.AuthError{
				Code:    "AUTH001",
				Message: "Invalid email or password",
			}
		}

		// Cache user data async
		go func() {
			if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
				fmt.Printf("failed to cache user data: %v\n", err)
			}
		}()
	}
	dbTime := time.Since(dbStart)
	fmt.Printf("Database operations took: %v\n", dbTime)

	// Check if account is locked
	if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
		return nil, "", &errors.AuthError{
			Code:    "AUTH002",
			Message: "Account is locked due to too many failed attempts",
		}
	}

	// Verify password with lower cost
	pwStart := time.Now()
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Increment failed login attempts
		user.FailedLoginAttempts++
		now := time.Now()
		user.LastFailedLogin = &now

		// Lock account if too many failed attempts
		if user.FailedLoginAttempts >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			user.AccountLockedUntil = &lockUntil
		}

		// Update user in database
		if err := s.authRepo.UpdateUser(user); err != nil {
			fmt.Printf("failed to update user failed attempts: %v\n", err)
		}

		// Update cache
		go func() {
			if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
				fmt.Printf("failed to update user cache: %v\n", err)
			}
		}()

		if user.AccountLockedUntil != nil {
			return nil, "", &errors.AuthError{
				Code:    "AUTH002",
				Message: "Account is locked due to too many failed attempts",
			}
		}

		return nil, "", &errors.AuthError{
			Code:    "AUTH001",
			Message: "Invalid email or password",
		}
	}

	// Reset failed login attempts on successful login
	user.FailedLoginAttempts = 0
	user.LastFailedLogin = nil
	user.AccountLockedUntil = nil
	if err := s.authRepo.UpdateUser(user); err != nil {
		fmt.Printf("failed to reset failed attempts: %v\n", err)
	}

	pwTime := time.Since(pwStart)
	fmt.Printf("Password verification took: %v\n", pwTime)

	// Get location from IP fully async (don't wait)
	deviceInfo.Location = "Unknown" // Set default
	go func() {
		location, err := s.geoService.GetLocation(deviceInfo.IPAddress)
		if err != nil {
			fmt.Printf("failed to get location: %v\n", err)
			return
		}
		// Update location in background
		deviceInfo.Location = location
	}()

	// Generate device ID if not provided
	if deviceInfo.DeviceID == "" {
		deviceID, err := security.GenerateDeviceID()
		if err != nil {
			fmt.Printf("failed to generate device ID: %v\n", err)
			deviceInfo.DeviceID = "unknown"
		} else {
			deviceInfo.DeviceID = deviceID
		}
	}

	// Generate JWT token
	tokenStart := time.Now()
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}
	tokenTime := time.Since(tokenStart)
	fmt.Printf("Token generation took: %v\n", tokenTime)

	// Log login and send notifications fully async
	go func() {
		// Log login
		loginHistory := &domain.LoginHistory{
			UserID:    user.ID,
			DeviceID:  deviceInfo.DeviceID,
			IPAddress: deviceInfo.IPAddress,
			Location:  deviceInfo.Location,
			UserAgent: deviceInfo.UserAgent,
			Status:    "success",
		}
		if err := s.authRepo.LogLogin(loginHistory); err != nil {
			fmt.Printf("failed to log login: %v\n", err)
		}

		// Send notification
		if err := s.emailService.SendLoginNotification(user.Email, deviceInfo); err != nil {
			fmt.Printf("failed to send login notification: %v\n", err)
		}
	}()

	totalTime := time.Since(startTime)
	fmt.Printf("Total auth service time: %v\n", totalTime)

	return user, token, nil
}

func (s *authService) ValidateToken(token string) (*domain.User, error) {
	// Validate JWT token
	userID, err := security.ValidateJWT(token, s.jwtSecret)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	// Get user from database
	user, err := s.authRepo.FindUserByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	return user, nil
}

// RefreshToken creates a new access token if the refresh token is valid
func (s *authService) RefreshToken(refreshToken string) (string, error) {
	// Validate refresh token
	userID, err := security.ValidateRefreshToken(refreshToken, s.jwtSecret)
	if err != nil {
		return "", errors.ErrInvalidRefreshToken
	}

	// Get user from database
	user, err := s.authRepo.FindUserByID(userID)
	if err != nil {
		return "", errors.ErrUserNotFound
	}

	// Generate new access token
	newToken, err := s.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return newToken, nil
}

func (s *authService) generateJWT(user *domain.User) (string, error) {
	// Generate access token with 15 minutes expiration
	return security.GenerateJWT(user.ID, s.jwtSecret, 15*time.Minute)
}

func (s *authService) ValidateLoginCode(userID uint, code string) error {
	return s.authRepo.ValidateAuthCode(userID, code, "login_verification")
}

func (s *authService) GetActiveSessions(userID uint) ([]*domain.DeviceSession, error) {
	return s.authRepo.GetActiveSessions(userID)
}

func (s *authService) RevokeSession(userID uint, deviceID string) error {
	// Do the session revocation async since it's not critical for immediate logout
	go func() {
		if err := s.authRepo.RevokeSession(userID, deviceID); err != nil {
			fmt.Printf("failed to revoke session: %v\n", err)
		}
	}()

	// Clear user cache immediately
	cacheKey := fmt.Sprintf("user:%d", userID)
	if err := s.cacheRepo.Delete(cacheKey); err != nil {
		fmt.Printf("failed to clear user cache: %v\n", err)
	}

	return nil
}

func (s *authService) GetLoginHistory(userID uint) ([]*domain.LoginHistory, error) {
	return s.authRepo.GetLoginHistory(userID)
}

func (s *authService) InitiateAccountRecovery(email string) error {
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	code, err := security.GenerateRandomCode(6)
	if err != nil {
		return fmt.Errorf("failed to generate recovery code: %w", err)
	}
	recovery := &domain.AccountRecovery{
		UserID:      user.ID,
		RequestType: "password_reset",
		Status:      "pending",
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.authRepo.CreateAccountRecovery(recovery); err != nil {
		return fmt.Errorf("failed to create recovery request: %w", err)
	}

	return s.emailService.SendPasswordResetEmail(email, code)
}

func (s *authService) ValidateRecoveryCode(email, code string) error {
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	return s.authRepo.ValidateAuthCode(user.ID, code, "password_reset")
}

func (s *authService) ResetPassword(email, code, newPassword string) error {
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.authRepo.ValidateAuthCode(user.ID, code, "password_reset"); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	return s.authRepo.UpdateUser(user)
}

func (s *authService) UpdateRecoveryEmail(userID uint, email string) error {
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.RecoveryEmail = email
	return s.authRepo.UpdateUser(user)
}
