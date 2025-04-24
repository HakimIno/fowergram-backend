package ports

import (
	"fowergram/internal/core/domain"
	"fowergram/pkg/email"
	"fowergram/pkg/geolocation"
	"fowergram/pkg/logger"
)

// AuthService handles authentication related operations
type AuthService interface {
	Register(user *domain.User) error
	Login(identifier, password string, deviceInfo *domain.DeviceSession) (*domain.User, string, error)
	ValidateToken(token string) (*domain.User, error)
	RefreshToken(refreshToken string) (string, error)
	ValidateLoginCode(userID uint, code string) error
	GetActiveSessions(userID uint) ([]*domain.DeviceSession, error)
	RevokeSession(userID uint, deviceID string) error
	GetLoginHistory(userID uint) ([]*domain.LoginHistory, error)
	InitiateAccountRecovery(email string) error
	ValidateRecoveryCode(email, code string) error
	ResetPassword(email, code, newPassword string) error
	UpdateRecoveryEmail(userID uint, email string) error
}

// UserService handles user related operations
type UserService interface {
	CreateUser(user *domain.User) error
	GetUserByID(id uint) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByUsername(username string) (*domain.User, error)
	UpdateUser(user *domain.User) error
	DeleteUser(id uint) error
	GetUsers(page, limit int) ([]*domain.User, error)
	GetUsersFromCache(cacheKey string) ([]*domain.User, error)
	CacheUsers(cacheKey string, users []*domain.User) error
}

// PostService handles post related operations
type PostService interface {
	CreatePost(post *domain.Post) error
	GetPostByID(id uint) (*domain.Post, error)
	GetAllPosts() ([]*domain.Post, error)
	UpdatePost(post *domain.Post) error
	DeletePost(id uint) error
}

// Factory function for AuthService
func NewAuthService(
	ar AuthRepository,
	es email.Service,
	gs geolocation.Service,
	cr CacheRepository,
	secret string,
	log *logger.ZerologService,
) AuthService {
	return nil // actual implementation will be in services package
}
