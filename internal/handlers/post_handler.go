package handlers

import (
	"fowergram/internal/core/domain"
	"fowergram/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type PostHandler struct {
	postService ports.PostService
}

func NewPostHandler(ps ports.PostService) *PostHandler {
	return &PostHandler{
		postService: ps,
	}
}

func (h *PostHandler) GetPosts(c *fiber.Ctx) error {
	posts, err := h.postService.GetAllPosts()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get posts",
		})
	}
	return c.JSON(posts)
}

func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	post := new(domain.Post)
	if err := c.BodyParser(post); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.postService.CreatePost(post); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create post",
		})
	}

	return c.JSON(post)
}
