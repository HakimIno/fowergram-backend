package service

import (
	"context"
	"errors"
	"fmt"
	"fowergram/internal/chat/broker"
	"fowergram/internal/chat/domain"
	"fowergram/internal/chat/repository"
	"sync"
	"time"
)

var (
	ErrChatNotFound = errors.New("chat not found")
	ErrInvalidInput = errors.New("invalid input")
)

type ChatService struct {
	repo      repository.ChatRepository
	wsManager *WebSocketManager
	msgBroker broker.MessageBroker
	cache     sync.Map // For caching chat data
}

func NewChatService(
	repo repository.ChatRepository,
	wsManager *WebSocketManager,
	msgBroker broker.MessageBroker,
) *ChatService {
	svc := &ChatService{
		repo:      repo,
		wsManager: wsManager,
		msgBroker: msgBroker,
	}

	if err := svc.subscribeToMessages(); err != nil {
		panic(fmt.Sprintf("Failed to subscribe to messages: %v", err))
	}
	return svc
}

func (s *ChatService) subscribeToMessages() error {
	return s.msgBroker.Subscribe(context.Background(), func(message *domain.Message) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try to get chat from cache first
		if cachedChat, ok := s.cache.Load(message.ChatID); ok {
			if chat, ok := cachedChat.(*domain.Chat); ok {
				s.broadcastMessage(chat, message)
				return
			}
		}

		// If not in cache, get from database
		chat, err := s.repo.GetChat(ctx, message.ChatID)
		if err != nil {
			fmt.Printf("Error getting chat: %v\n", err)
			return
		}

		// Cache the chat data
		s.cache.Store(message.ChatID, chat)
		s.broadcastMessage(chat, message)
	})
}

func (s *ChatService) broadcastMessage(chat *domain.Chat, message *domain.Message) {
	var wg sync.WaitGroup
	for _, memberID := range chat.Members {
		if memberID == message.SenderID {
			continue
		}
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := s.wsManager.SendToUser(id, message); err != nil {
				fmt.Printf("Error sending message to user %s: %v\n", id, err)
			}
		}(memberID)
	}
	wg.Wait()
}

func (s *ChatService) SendMessage(ctx context.Context, message *domain.Message) error {
	if message == nil || message.ChatID == "" || message.SenderID == "" {
		return ErrInvalidInput
	}

	// Save message to database
	if err := s.repo.SaveMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Publish message to message broker
	if err := s.msgBroker.PublishMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (s *ChatService) CreateChat(ctx context.Context, chat *domain.Chat) error {
	if chat == nil || len(chat.Members) < 2 {
		return ErrInvalidInput
	}

	if err := s.repo.CreateChat(ctx, chat); err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	// Cache the new chat
	s.cache.Store(chat.ID, chat)
	return nil
}

func (s *ChatService) GetMessages(ctx context.Context, chatID string, limit, offset int) ([]domain.Message, error) {
	if chatID == "" {
		return nil, ErrInvalidInput
	}

	if limit <= 0 {
		limit = 50
	}

	messages, err := s.repo.GetMessages(ctx, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID string) ([]domain.Chat, error) {
	if userID == "" {
		return nil, ErrInvalidInput
	}

	chats, err := s.repo.GetUserChats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}

	// Cache all chats
	for _, chat := range chats {
		s.cache.Store(chat.ID, &chat)
	}

	return chats, nil
}

func (s *ChatService) Shutdown() {
	s.wsManager.Shutdown()
}
