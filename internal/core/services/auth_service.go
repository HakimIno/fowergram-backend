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
	"fowergram/pkg/logger"
	"fowergram/pkg/security"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	authRepo     ports.AuthRepository
	emailService email.Service
	geoService   geolocation.Service
	cacheRepo    ports.CacheRepository
	jwtSecret    string
	log          *logger.ZerologService
}

type AuthMetrics struct {
	CacheLookup        time.Duration
	DatabaseOperations time.Duration
	PasswordVerify     time.Duration
	TokenGeneration    time.Duration
	TotalTime          time.Duration
}

func NewAuthServiceWithLogger(
	ar ports.AuthRepository,
	es email.Service,
	gs geolocation.Service,
	cr ports.CacheRepository,
	secret string,
	log *logger.ZerologService,
) ports.AuthService {
	return &authService{
		authRepo:     ar,
		emailService: es,
		geoService:   gs,
		cacheRepo:    cr,
		jwtSecret:    secret,
		log:          log,
	}
}

func NewAuthService(
	ar ports.AuthRepository,
	es email.Service,
	gs geolocation.Service,
	cr ports.CacheRepository,
	secret string,
) ports.AuthService {
	return NewAuthServiceWithLogger(ar, es, gs, cr, secret, logger.NewLogger(logger.InfoLevel))
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
	s.log.Info("Password hashing completed",
		logger.NewField("duration", hashTime),
	)

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
	s.log.Info("User creation completed",
		logger.NewField("duration", createTime),
	)

	// Do all non-critical operations async
	go func() {
		// Cache user data
		cacheKey := fmt.Sprintf("user:%d", user.ID)
		if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
			s.log.Error("Failed to cache user data", err,
				logger.NewField("user_id", user.ID),
			)
		}

		// Generate verification code
		code, err := security.GenerateRandomCode(6)
		if err != nil {
			s.log.Error("Failed to generate verification code", err,
				logger.NewField("user_id", user.ID),
			)
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
			s.log.Error("Failed to create auth code", err,
				logger.NewField("user_id", user.ID),
			)
			return
		}

		// Send verification email
		if err := s.emailService.SendVerificationEmail(user.Email, code); err != nil {
			s.log.Error("Failed to send verification email", err,
				logger.NewField("user_id", user.ID),
			)
		}
	}()

	totalTime := time.Since(startTime)
	s.log.Info("Total registration completed",
		logger.NewField("duration", totalTime),
	)

	return nil
}

func (s *authService) Login(identifier, password string, deviceInfo *domain.DeviceSession) (*domain.User, string, error) {
	startTime := time.Now()
	metrics := &AuthMetrics{}

	// Try to determine if identifier is email or username
	isEmail := strings.Contains(identifier, "@")

	// Try to get user from cache first
	var cacheKey string
	if isEmail {
		cacheKey = fmt.Sprintf("user:email:%s", identifier)
	} else {
		cacheKey = fmt.Sprintf("user:username:%s", identifier)
	}

	var user *domain.User
	cacheDone := make(chan bool, 1)

	go func() {
		cacheStart := time.Now()
		if cached, err := s.cacheRepo.Get(cacheKey); err == nil {
			if userData, ok := cached.(*domain.User); ok {
				user = userData
			}
		}
		metrics.CacheLookup = time.Since(cacheStart)
		cacheDone <- true
	}()

	select {
	case <-cacheDone:
		s.log.Debug("Cache lookup completed",
			logger.NewField("duration", metrics.CacheLookup),
			logger.NewField("cache_hit", user != nil),
		)
	case <-time.After(100 * time.Millisecond):
		s.log.Warn("Cache lookup timed out")
	}

	// Database operations
	dbStart := time.Now()
	if user == nil {
		var err error
		if isEmail {
			user, err = s.authRepo.FindUserByEmail(identifier)
		} else {
			user, err = s.authRepo.FindUserByUsername(identifier)
		}

		if err != nil {
			s.log.Error("User lookup failed", err,
				logger.NewField("identifier", identifier),
			)
			return nil, "", &errors.AuthError{
				Code:    "AUTH001",
				Message: "Invalid identifier or password",
			}
		}

		go func() {
			if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
				s.log.Error("Failed to cache user data", err,
					logger.NewField("user_id", user.ID),
				)
			}
		}()
	}
	metrics.DatabaseOperations = time.Since(dbStart)

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
		s.log.Warn("Password verification failed",
			logger.NewField("user_id", user.ID),
			logger.NewField("attempts", user.FailedLoginAttempts+1),
		)
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
			s.log.Error("Failed to update user failed attempts", err,
				logger.NewField("user_id", user.ID),
			)
		}

		// Update cache
		go func() {
			if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
				s.log.Error("Failed to update user cache", err,
					logger.NewField("user_id", user.ID),
				)
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
			Message: "Invalid identifier or password",
		}
	}

	// Reset failed login attempts on successful login
	user.FailedLoginAttempts = 0
	user.LastFailedLogin = nil
	user.AccountLockedUntil = nil
	if err := s.authRepo.UpdateUser(user); err != nil {
		s.log.Error("Failed to reset failed attempts", err) // logger.NewField("user_id", user.ID),

	}

	pwTime := time.Since(pwStart)
	metrics.PasswordVerify = pwTime
	// s.log.Info("Password verification completed",
	// 	logger.NewField("duration", pwTime),
	// )

	// Create a channel to receive location
	locationChan := make(chan string, 1)

	// Get location from IP fully async
	deviceInfo.SetLocation("Unknown") // Set default
	go func() {
		location, err := s.geoService.GetLocation(deviceInfo.IPAddress)
		if err != nil {
			s.log.Error("Failed to get location", err) // logger.NewField("ip_address", deviceInfo.IPAddress),

			locationChan <- "Unknown"
			return
		}
		locationChan <- location
	}()

	// Generate device ID if not provided
	if deviceInfo.DeviceID == "" {
		deviceID, err := security.GenerateDeviceID()
		if err != nil {
			s.log.Error("Failed to generate device ID", err) // logger.NewField("user_id", user.ID),

			deviceInfo.DeviceID = "unknown"
		} else {
			deviceInfo.DeviceID = deviceID
		}
	}

	// Token generation
	tokenStart := time.Now()
	token, err := s.generateJWT(user)
	if err != nil {
		s.log.Error("Token generation failed", err) // logger.NewField("user_id", user.ID),

		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}
	metrics.TokenGeneration = time.Since(tokenStart)

	// Wait for location with timeout
	select {
	case location := <-locationChan:
		deviceInfo.SetLocation(location)
	case <-time.After(100 * time.Millisecond):
		// Use default "Unknown" if timeout
	}

	// Log login and send notifications fully async
	go func() {
		// Log login
		loginHistory := &domain.LoginHistory{
			UserID:    user.ID,
			DeviceID:  deviceInfo.DeviceID,
			IPAddress: deviceInfo.IPAddress,
			Location:  deviceInfo.GetLocation(), // Safe to access now
			UserAgent: deviceInfo.UserAgent,
			Status:    "success",
		}
		if err := s.authRepo.LogLogin(loginHistory); err != nil {
			s.log.Error("Failed to log login", err,
				logger.NewField("user_id", user.ID),
			)
		}

		// Send notification
		if err := s.emailService.SendLoginNotification(user.Email, deviceInfo); err != nil {
			s.log.Error("Failed to send login notification", err,
				logger.NewField("user_id", user.ID),
			)
		}
	}()

	metrics.TotalTime = time.Since(startTime)
	s.logAuthMetrics(metrics)

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
			s.log.Error("Failed to revoke session", err,
				logger.NewField("user_id", userID),
				logger.NewField("device_id", deviceID),
			)
		}
	}()

	// Clear user cache immediately
	cacheKey := fmt.Sprintf("user:%d", userID)
	if err := s.cacheRepo.Delete(cacheKey); err != nil {
		s.log.Error("Failed to clear user cache", err,
			logger.NewField("user_id", userID),
		)
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

func (s *authService) logAuthMetrics(metrics *AuthMetrics) {
	s.log.Info("Auth Service Metrics") // logger.NewField("Operation", "login"),
	// logger.NewField("Cache Lookup", formatDuration(metrics.CacheLookup)),
	// logger.NewField("Database Operations", formatDuration(metrics.DatabaseOperations)),
	// logger.NewField("Password Verify", formatDuration(metrics.PasswordVerify)),
	// logger.NewField("Token Generation", formatDuration(metrics.TokenGeneration)),
	// logger.NewField("Total Time", formatDuration(metrics.TotalTime)),

}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000)
}
