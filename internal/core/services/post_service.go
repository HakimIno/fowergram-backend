package services

import (
	"fowergram/internal/domain"
	"fowergram/internal/core/ports"
)

type postService struct {
	postRepo  ports.PostRepository
	cacheRepo ports.CacheRepository
}

func NewPostService(pr ports.PostRepository, cr ports.CacheRepository) ports.PostService {
	return &postService{
		postRepo:  pr,
		cacheRepo: cr,
	}
}

func (s *postService) CreatePost(post *domain.Post) error {
	return s.postRepo.Create(post)
}

func (s *postService) GetPostByID(id uint) (*domain.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *postService) GetAllPosts() ([]*domain.Post, error) {
	return s.postRepo.FindAll()
}

func (s *postService) UpdatePost(post *domain.Post) error {
	return s.postRepo.Update(post)
}

func (s *postService) DeletePost(id uint) error {
	return s.postRepo.Delete(id)
}
