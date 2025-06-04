package services

import (
	"fmt"
	"testing"
	"time"

	"fowergram/internal/domain"

	"github.com/redis/go-redis/v9"
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

func (m *MockAuthRepo) FindUserByUsername(username string) (*domain.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
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

type MockCacheRepo struct {
	mock.Mock
}

func (m *MockCacheRepo) Set(key string, value interface{}, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheRepo) Get(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheRepo) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	mockEmail := new(MockEmailService)
	mockGeo := new(MockGeoService)
	mockCache := new(MockCacheRepo)
	service := NewAuthService(mockRepo, mockEmail, mockGeo, mockCache, "secret")

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
				mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("*domain.User"), mock.AnythingOfType("time.Duration")).Return(nil)
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
				mockRepo.On("CreateUser", mock.AnythingOfType("*domain.User")).Return(fmt.Errorf("duplicate key value"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockEmail.ExpectedCalls = nil
			mockCache.ExpectedCalls = nil
			tt.setup()

			err := service.Register(tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Wait for async operations
				time.Sleep(100 * time.Millisecond)
				mockRepo.AssertExpectations(t)
				mockEmail.AssertExpectations(t)
				mockCache.AssertExpectations(t)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	mockEmail := new(MockEmailService)
	mockGeo := new(MockGeoService)
	mockCache := new(MockCacheRepo)
	service := NewAuthService(mockRepo, mockEmail, mockGeo, mockCache, "secret")

	// Create test user with hashed password
	password := "Test123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	tests := []struct {
		name    string
		email   string
		pass    string
		wantErr bool
		setup   func()
	}{
		{
			name:    "successful login",
			email:   "test@example.com",
			pass:    "Test123!",
			wantErr: false,
			setup: func() {
				// Setup cache mock
				mockCache.On("Get", "user:email:test@example.com").Return(nil, redis.Nil)
				mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("*domain.User"), mock.AnythingOfType("time.Duration")).Return(nil)

				// Setup repository mock
				mockRepo.On("FindUserByEmail", "test@example.com").Return(testUser, nil)
				mockRepo.On("UpdateUser", mock.AnythingOfType("*domain.User")).Return(nil)
				mockRepo.On("CreateLoginHistory", mock.AnythingOfType("*domain.LoginHistory")).Return(nil)
				mockRepo.On("LogLogin", mock.AnythingOfType("*domain.LoginHistory")).Return(nil)
				mockRepo.On("CreateDeviceSession", mock.AnythingOfType("*domain.DeviceSession")).Return(nil)

				// Setup geo service mock
				mockGeo.On("GetLocation", mock.AnythingOfType("string")).Return("Test Location", nil)

				// Setup email service mock
				mockEmail.On("SendLoginNotification", mock.AnythingOfType("string"), mock.AnythingOfType("*domain.DeviceSession")).Return(nil)
			},
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockEmail.ExpectedCalls = nil
			mockCache.ExpectedCalls = nil
			mockGeo.ExpectedCalls = nil
			tt.setup()

			_, _, _, err := service.Login(tt.email, tt.pass, &domain.DeviceSession{
				DeviceType: "Browser",
				IPAddress:  "127.0.0.1",
			})
			if tt.wantErr {
				assert.Error(t, err)
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
	mockCache := new(MockCacheRepo)
	service := NewAuthService(mockRepo, mockEmail, mockGeo, mockCache, "secret")

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
	mockCache := new(MockCacheRepo)
	service := NewAuthService(mockRepo, mockEmail, mockGeo, mockCache, "secret")

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

func (m *MockGeoService) GetLocationFromIP(ip string) (string, error) {
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
