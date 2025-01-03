package ports

import "fowergram/internal/core/domain"

type UserService interface {
	CreateUser(user *domain.User) error
	GetUserByID(id uint) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	UpdateUser(user *domain.User) error
}

type PostService interface {
	CreatePost(post *domain.Post) error
	GetPostByID(id uint) (*domain.Post, error)
	GetAllPosts() ([]*domain.Post, error)
	UpdatePost(post *domain.Post) error
	DeletePost(id uint) error
}

type AuthService interface {
	Login(email, password string) (string, error)
	Register(user *domain.User) error
	ValidateToken(token string) (*domain.User, error)
}
