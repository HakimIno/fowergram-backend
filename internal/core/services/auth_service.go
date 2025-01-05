package services

import (
	"fmt"
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
	jwtSecret    string
}

func NewAuthService(ar ports.AuthRepository, es email.Service, gs geolocation.Service, secret string) ports.AuthService {
	return &authService{
		authRepo:     ar,
		emailService: es,
		geoService:   gs,
		jwtSecret:    secret,
	}
}

func (s *authService) Register(user *domain.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	// Create user
	if err := s.authRepo.CreateUser(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Generate verification code
	code, err := security.GenerateRandomCode(6)
	if err != nil {
		return fmt.Errorf("failed to generate verification code: %w", err)
	}
	authCode := &domain.AuthCode{
		UserID:    user.ID,
		Code:      code,
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.authRepo.CreateAuthCode(authCode); err != nil {
		return fmt.Errorf("failed to create auth code: %w", err)
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(user.Email, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *authService) Login(email, password string, deviceInfo *domain.DeviceSession) (*domain.User, string, error) {
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return nil, "", errors.ErrInvalidCredentials
	}

	// Check if account is locked
	if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
		return nil, "", errors.ErrAccountLocked
	}

	// Verify password
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

		if err := s.authRepo.UpdateUser(user); err != nil {
			return nil, "", fmt.Errorf("failed to update user: %w", err)
		}

		return nil, "", errors.ErrInvalidCredentials
	}

	// Reset failed login attempts
	user.FailedLoginAttempts = 0
	user.LastFailedLogin = nil
	user.AccountLockedUntil = nil

	// Get location info
	location, err := s.geoService.GetLocation(deviceInfo.IPAddress)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get location: %w", err)
	}
	deviceInfo.Location = location

	// Generate device ID if not exists
	if deviceInfo.DeviceID == "" {
		deviceID, err := generateDeviceID()
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate device ID: %w", err)
		}
		deviceInfo.DeviceID = deviceID
	}

	// Create or update device session
	deviceInfo.UserID = user.ID
	deviceInfo.LastActive = time.Now()
	if err := s.authRepo.CreateDeviceSession(deviceInfo); err != nil {
		return nil, "", fmt.Errorf("failed to create device session: %w", err)
	}

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
		return nil, "", fmt.Errorf("failed to log login: %w", err)
	}

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Send login notification if new device
	if err := s.emailService.SendLoginNotification(user.Email, deviceInfo); err != nil {
		// Log error but don't fail the login
		fmt.Printf("failed to send login notification: %v", err)
	}

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

func generateDeviceID() (string, error) {
	return security.GenerateDeviceID(), nil
}

func (s *authService) ValidateLoginCode(userID uint, code string) error {
	return s.authRepo.ValidateAuthCode(userID, code, "login_verification")
}

func (s *authService) GetActiveSessions(userID uint) ([]*domain.DeviceSession, error) {
	return s.authRepo.GetActiveSessions(userID)
}

func (s *authService) RevokeSession(userID uint, deviceID string) error {
	return s.authRepo.RevokeSession(userID, deviceID)
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
