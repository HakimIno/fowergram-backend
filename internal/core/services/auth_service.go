package services

import (
	"fowergram/internal/core/domain"
	"fowergram/internal/core/ports"
)

type authService struct {
	userRepo ports.UserRepository
}

func NewAuthService(ur ports.UserRepository) ports.AuthService {
	return &authService{
		userRepo: ur,
	}
}

func (s *authService) Login(email, password string) (string, error) {
	// TODO: Implement login logic
	return "", nil
}

func (s *authService) Register(user *domain.User) error {
	return s.userRepo.Create(user)
}

func (s *authService) ValidateToken(token string) (*domain.User, error) {
	// TODO: Implement token validation
	return nil, nil
}
