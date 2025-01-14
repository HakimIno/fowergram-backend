package broker

import (
	"context"
	"fowergram/internal/chat/domain"
)

type MessageBroker interface {
	PublishMessage(ctx context.Context, message *domain.Message) error
	Subscribe(ctx context.Context, handler func(*domain.Message)) error
	Close() error
}
