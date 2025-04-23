package ports

import (
	"fowergram/internal/core/domain"
	"time"
)

type UserRepository interface {
	Create(user *domain.User) error
	FindByID(id uint) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	FindAll(page, limit int) ([]*domain.User, error)
	Update(user *domain.User) error
	Delete(id uint) error
}

type PostRepository interface {
	Create(post *domain.Post) error
	FindByID(id uint) (*domain.Post, error)
	FindAll() ([]*domain.Post, error)
	Update(post *domain.Post) error
	Delete(id uint) error
}

type CacheRepository interface {
	Set(key string, value interface{}, ttl time.Duration) error
	Get(key string) (interface{}, error)
	Delete(key string) error
}

type AuthRepository interface {
	CreateUser(user *domain.User) error
	FindUserByEmail(email string) (*domain.User, error)
	FindUserByUsername(username string) (*domain.User, error)
	FindUserByID(id uint) (*domain.User, error)
	UpdateUser(user *domain.User) error
	CreateDeviceSession(session *domain.DeviceSession) error
	GetActiveSessions(userID uint) ([]*domain.DeviceSession, error)
	RevokeSession(userID uint, deviceID string) error
	CreateAuthCode(code *domain.AuthCode) error
	ValidateAuthCode(userID uint, code string, purpose string) error
	LogLogin(history *domain.LoginHistory) error
	GetLoginHistory(userID uint) ([]*domain.LoginHistory, error)
	CreateAccountRecovery(recovery *domain.AccountRecovery) error
}
