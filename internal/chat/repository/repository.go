package repository

import (
	"context"
	"fowergram/internal/chat/domain"
)

type ChatRepository interface {
	CreateChat(ctx context.Context, chat *domain.Chat) error
	GetChat(ctx context.Context, chatID string) (*domain.Chat, error)
	SaveMessage(ctx context.Context, message *domain.Message) error
	GetMessages(ctx context.Context, chatID string, limit, offset int) ([]domain.Message, error)
	GetUserChats(ctx context.Context, userID string) ([]domain.Chat, error)
}
