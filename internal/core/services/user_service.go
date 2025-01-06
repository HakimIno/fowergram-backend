package services

import (
	"encoding/json"
	"fmt"
	"time"

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
	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	// Cache user data
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to cache user data: %v\n", err)
	}

	return nil
}

func (s *userService) GetUserByID(id uint) (*domain.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:%d", id)
	if cached, err := s.cacheRepo.Get(cacheKey); err == nil {
		if userData, ok := cached.(map[string]interface{}); ok {
			// Convert cached data to User struct
			userBytes, err := json.Marshal(userData)
			if err == nil {
				var user domain.User
				if err := json.Unmarshal(userBytes, &user); err == nil {
					return &user, nil
				}
			}
		}
	}

	// If not in cache or error, get from database
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Cache the user data
	if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to cache user data: %v\n", err)
	}

	return user, nil
}

func (s *userService) GetUserByEmail(email string) (*domain.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:email:%s", email)
	if cached, err := s.cacheRepo.Get(cacheKey); err == nil {
		if userData, ok := cached.(map[string]interface{}); ok {
			// Convert cached data to User struct
			userBytes, err := json.Marshal(userData)
			if err == nil {
				var user domain.User
				if err := json.Unmarshal(userBytes, &user); err == nil {
					return &user, nil
				}
			}
		}
	}

	// If not in cache or error, get from database
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	// Cache the user data
	if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to cache user data: %v\n", err)
	}

	return user, nil
}

func (s *userService) UpdateUser(user *domain.User) error {
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Update cache
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to update user cache: %v\n", err)
	}

	// Update email cache
	emailCacheKey := fmt.Sprintf("user:email:%s", user.Email)
	if err := s.cacheRepo.Set(emailCacheKey, user, 24*time.Hour); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to update user email cache: %v\n", err)
	}

	return nil
}
