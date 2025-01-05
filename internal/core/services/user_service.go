package services

import (
	"fowergram/internal/core/domain"
	"fowergram/internal/core/ports"
)

type userService struct {
	userRepo  ports.UserRepository
	cacheRepo ports.CacheRepository
}

func NewUserService(ur ports.UserRepository, cr ports.CacheRepository) ports.UserService {
	return &userService{
		userRepo:  ur,
		cacheRepo: cr,
	}
}

func (s *userService) CreateUser(user *domain.User) error {
	return s.userRepo.Create(user)
}

func (s *userService) GetUserByID(id uint) (*domain.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) GetUserByEmail(email string) (*domain.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *userService) UpdateUser(user *domain.User) error {
	return s.userRepo.Update(user)
}
