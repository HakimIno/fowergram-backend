package services

import (
	"fmt"
	"testing"
	"time"

	"fowergram/internal/core/domain"
	"fowergram/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockAuthRepo struct {
	mock.Mock
	users map[string]*domain.User
}

func NewMockAuthRepo() *MockAuthRepo {
	return &MockAuthRepo{
		users: make(map[string]*domain.User),
	}
}

func (m *MockAuthRepo) FindUserByID(id uint) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// Implement all AuthRepository methods...

type MockEmailService struct {
	mock.Mock
}

// Implement all EmailService methods...

type MockGeoService struct {
	mock.Mock
}

// Implement all GeoService methods...

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	mockEmail := new(MockEmailService)
	mockGeo := new(MockGeoService)
	service := NewAuthService(mockRepo, mockEmail, mockGeo, "secret")

	tests := []struct {
		name    string
		user    *domain.User
		wantErr bool
		setup   func()
	}{
		{
			name: "successful registration",
			user: &domain.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "Test123!",
			},
			wantErr: false,
			setup: func() {
				mockRepo.On("CreateUser", mock.AnythingOfType("*domain.User")).Return(nil)
				mockRepo.On("CreateAuthCode", mock.AnythingOfType("*domain.AuthCode")).Return(nil)
				mockEmail.On("SendVerificationEmail", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "duplicate email",
			user: &domain.User{
				Username:     "testuser2",
				Email:        "existing@example.com",
				PasswordHash: "Test123!",
			},
			wantErr: true,
			setup: func() {
				mockRepo.On("CreateUser", mock.AnythingOfType("*domain.User")).Return(fmt.Errorf("duplicate email"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockEmail.ExpectedCalls = nil
			tt.setup()

			err := service.Register(tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	mockEmail := new(MockEmailService)
	mockGeo := new(MockGeoService)
	service := NewAuthService(mockRepo, mockEmail, mockGeo, "secret")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name       string
		email      string
		password   string
		deviceInfo *domain.DeviceSession
		wantErr    bool
		setup      func()
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			deviceInfo: &domain.DeviceSession{
				DeviceType: "Browser",
				IPAddress:  "127.0.0.1",
			},
			wantErr: false,
			setup: func() {
				mockRepo.On("FindUserByEmail", "test@example.com").Return(&domain.User{
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
				}, nil)
				mockRepo.On("UpdateUser", mock.AnythingOfType("*domain.User")).Return(nil)
				mockGeo.On("GetLocation", "127.0.0.1").Return("Test Location", nil)
				mockRepo.On("CreateDeviceSession", mock.AnythingOfType("*domain.DeviceSession")).Return(nil)
				mockRepo.On("LogLogin", mock.AnythingOfType("*domain.LoginHistory")).Return(nil)
				mockEmail.On("SendLoginNotification", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:     "account locked",
			email:    "locked@example.com",
			password: "password123",
			deviceInfo: &domain.DeviceSession{
				DeviceType: "Browser",
				IPAddress:  "127.0.0.1",
			},
			wantErr: true,
			setup: func() {
				lockTime := time.Now().Add(15 * time.Minute)
				mockRepo.On("FindUserByEmail", "locked@example.com").Return(&domain.User{
					Email:              "locked@example.com",
					PasswordHash:       string(hashedPassword),
					AccountLockedUntil: &lockTime,
				}, nil)
				mockRepo.On("UpdateUser", mock.AnythingOfType("*domain.User")).Return(errors.ErrAccountLocked)
			},
		},
		{
			name:     "too many failed attempts",
			email:    "failing@example.com",
			password: "wrongpass",
			deviceInfo: &domain.DeviceSession{
				DeviceType: "Browser",
				IPAddress:  "127.0.0.1",
			},
			wantErr: true,
			setup: func() {
				mockRepo.On("FindUserByEmail", "failing@example.com").Return(&domain.User{
					Email:               "failing@example.com",
					PasswordHash:        "$2a$10$...",
					FailedLoginAttempts: 4,
				}, nil)
				mockRepo.On("UpdateUser", mock.AnythingOfType("*domain.User")).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockEmail.ExpectedCalls = nil
			mockGeo.ExpectedCalls = nil
			tt.setup()

			_, _, err := service.Login(tt.email, tt.password, tt.deviceInfo)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.name == "account locked" {
					assert.Equal(t, "Account is locked due to too many failed attempts", err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_ValidateLoginCode(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	mockEmail := new(MockEmailService)
	mockGeo := new(MockGeoService)
	service := NewAuthService(mockRepo, mockEmail, mockGeo, "secret")

	tests := []struct {
		name    string
		userID  uint
		code    string
		wantErr bool
		setup   func()
	}{
		{
			name:    "valid code",
			userID:  1,
			code:    "123456",
			wantErr: false,
			setup: func() {
				mockRepo.On("ValidateAuthCode", uint(1), "123456", "login_verification").Return(nil)
			},
		},
		{
			name:    "invalid code",
			userID:  1,
			code:    "000000",
			wantErr: true,
			setup: func() {
				mockRepo.On("ValidateAuthCode", uint(1), "000000", "login_verification").
					Return(fmt.Errorf("invalid code"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.setup()

			err := service.ValidateLoginCode(tt.userID, tt.code)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_InitiateAccountRecovery(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	mockEmail := new(MockEmailService)
	mockGeo := new(MockGeoService)
	service := NewAuthService(mockRepo, mockEmail, mockGeo, "secret")

	tests := []struct {
		name    string
		email   string
		wantErr bool
		setup   func()
	}{
		{
			name:    "successful initiation",
			email:   "test@example.com",
			wantErr: false,
			setup: func() {
				mockRepo.On("FindUserByEmail", "test@example.com").Return(&domain.User{
					ID:    1,
					Email: "test@example.com",
				}, nil)
				mockRepo.On("CreateAccountRecovery", mock.AnythingOfType("*domain.AccountRecovery")).Return(nil)
				mockEmail.On("SendPasswordResetEmail", "test@example.com", mock.Anything).Return(nil)
			},
		},
		{
			name:    "user not found",
			email:   "nonexistent@example.com",
			wantErr: true,
			setup: func() {
				mockRepo.On("FindUserByEmail", "nonexistent@example.com").
					Return(nil, fmt.Errorf("user not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockEmail.ExpectedCalls = nil
			tt.setup()

			err := service.InitiateAccountRecovery(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Add more test functions for other methods

// MockEmailService methods
func (m *MockEmailService) SendVerificationEmail(to, code string) error {
	args := m.Called(to, code)
	return args.Error(0)
}

func (m *MockEmailService) SendLoginNotification(to string, device *domain.DeviceSession) error {
	args := m.Called(to, device)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(to, code string) error {
	args := m.Called(to, code)
	return args.Error(0)
}

// MockGeoService methods
func (m *MockGeoService) GetLocation(ip string) (string, error) {
	args := m.Called(ip)
	return args.String(0), args.Error(1)
}

// MockAuthRepo methods
func (m *MockAuthRepo) CreateUser(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// Additional MockAuthRepo methods
func (m *MockAuthRepo) FindUserByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthRepo) UpdateUser(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockAuthRepo) CreateDeviceSession(session *domain.DeviceSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockAuthRepo) GetActiveSessions(userID uint) ([]*domain.DeviceSession, error) {
	args := m.Called(userID)
	return args.Get(0).([]*domain.DeviceSession), args.Error(1)
}

func (m *MockAuthRepo) RevokeSession(userID uint, deviceID string) error {
	args := m.Called(userID, deviceID)
	return args.Error(0)
}

func (m *MockAuthRepo) LogLogin(history *domain.LoginHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *MockAuthRepo) GetLoginHistory(userID uint) ([]*domain.LoginHistory, error) {
	args := m.Called(userID)
	return args.Get(0).([]*domain.LoginHistory), args.Error(1)
}

func (m *MockAuthRepo) CreateAuthCode(code *domain.AuthCode) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockAuthRepo) ValidateAuthCode(userID uint, code string, purpose string) error {
	args := m.Called(userID, code, purpose)
	return args.Error(0)
}

func (m *MockAuthRepo) CreateAccountRecovery(recovery *domain.AccountRecovery) error {
	args := m.Called(recovery)
	return args.Error(0)
}

func (m *MockAuthRepo) UpdateAccountRecovery(recovery *domain.AccountRecovery) error {
	args := m.Called(recovery)
	return args.Error(0)
}