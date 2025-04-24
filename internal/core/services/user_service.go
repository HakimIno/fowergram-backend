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
	// Try cache first
	cacheKey := fmt.Sprintf("user:%d", id)
	if cached, err := s.cacheRepo.Get(cacheKey); err == nil {
		if user, ok := cached.(*domain.User); ok {
			return user, nil
		}
	}

	// Get from database
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Cache async
	go func() {
		if err := s.cacheRepo.Set(cacheKey, user, 24*time.Hour); err != nil {
			fmt.Printf("failed to cache user: %v\n", err)
		}
	}()

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

func (s *userService) GetUsers(page, limit int) ([]*domain.User, error) {
	return s.userRepo.FindAll(page, limit)
}

func (s *userService) GetUsersFromCache(cacheKey string) ([]*domain.User, error) {
	cached, err := s.cacheRepo.Get(cacheKey)
	if err != nil {
		return nil, err
	}

	// Convert cached data back to users array
	if data, ok := cached.([]byte); ok {
		var users []*domain.User
		if err := json.Unmarshal(data, &users); err != nil {
			return nil, err
		}
		return users, nil
	}

	return nil, fmt.Errorf("invalid cache data type")
}

func (s *userService) CacheUsers(cacheKey string, users []*domain.User) error {
	// Cache for 5 minutes since user list might change frequently
	return s.cacheRepo.Set(cacheKey, users, 5*time.Minute)
}

func (s *userService) UpdateUser(user *domain.User) error {
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Clear cache async
	go func() {
		cacheKey := fmt.Sprintf("user:%d", user.ID)
		if err := s.cacheRepo.Delete(cacheKey); err != nil {
			fmt.Printf("failed to clear user cache: %v\n", err)
		}
	}()

	return nil
}

func (s *userService) DeleteUser(id uint) error {
	if err := s.userRepo.Delete(id); err != nil {
		return err
	}

	// Clear cache async
	go func() {
		cacheKey := fmt.Sprintf("user:%d", id)
		if err := s.cacheRepo.Delete(cacheKey); err != nil {
			fmt.Printf("failed to clear user cache: %v\n", err)
		}
	}()

	return nil
}

func (s *userService) GetUserByUsername(username string) (*domain.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:username:%s", username)
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
	user, err := s.userRepo.FindByUsername(username)
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
