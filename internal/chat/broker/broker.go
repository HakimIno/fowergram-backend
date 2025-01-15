package broker

import (
	"context"
	"fowergram/internal/chat/domain"
)

type MessageProducer interface {
	ProduceMessage(ctx context.Context, message *domain.Message) error
}

type MessageConsumer interface {
	Subscribe(ctx context.Context, handler func(*domain.Message)) error
	Close() error
}

type MessageBroker interface {
	MessageProducer
	MessageConsumer
}
