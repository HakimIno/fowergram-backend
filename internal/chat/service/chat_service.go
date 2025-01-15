package service

import (
	"context"
	"time"

	"fowergram/internal/chat/broker"
	"fowergram/internal/chat/domain"
	"fowergram/internal/chat/repository"
)

type ChatService struct {
	repo     repository.ChatRepository
	wsm      *WebSocketManager
	producer broker.MessageProducer
}

func NewChatService(repo repository.ChatRepository, wsm *WebSocketManager, producer broker.MessageProducer) *ChatService {
	return &ChatService{
		repo:     repo,
		wsm:      wsm,
		producer: producer,
	}
}

func (s *ChatService) HandleMessage(ctx context.Context, msg *domain.Message) error {
	repoMsg := toRepositoryMessage(msg)
	if err := s.repo.SaveMessage(ctx, repoMsg); err != nil {
		return err
	}

	if err := s.producer.ProduceMessage(ctx, msg); err != nil {
		return err
	}

	s.wsm.BroadcastToChat(msg.ConversationID, msg)
	return nil
}

func (s *ChatService) CreateChat(ctx context.Context, chat *domain.Chat) error {
	repoChat := toRepositoryChat(chat)
	if err := s.repo.CreateChat(ctx, repoChat); err != nil {
		return err
	}

	for _, userID := range chat.Members {
		member := &domain.ChatMember{
			ChatID:    chat.ID,
			UserID:    userID,
			Role:      string(domain.ChatRoleMember),
			JoinedAt:  time.Now(),
			UpdatedAt: time.Now(),
		}
		repoMember := toRepositoryChatMember(member)
		if err := s.repo.AddChatMember(ctx, repoMember); err != nil {
			return err
		}
	}

	return nil
}

func (s *ChatService) GetMessages(ctx context.Context, conversationID string, limit int) ([]*domain.Message, error) {
	messages, err := s.repo.GetMessages(ctx, conversationID, limit)
	if err != nil {
		return nil, err
	}

	return toDomainMessages(messages), nil
}

func (s *ChatService) GetChat(ctx context.Context, chatID string) (*domain.Chat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	members, err := s.repo.GetChatMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	domainChat := toDomainChat(chat)
	if domainChat != nil {
		domainMembers := toDomainChatMembers(members)
		domainChat.Members = make([]string, len(domainMembers))
		for i, member := range domainMembers {
			domainChat.Members[i] = member.UserID
		}
	}
	return domainChat, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID string) ([]*domain.Chat, error) {
	members, err := s.repo.GetChatMembers(ctx, userID)
	if err != nil {
		return nil, err
	}

	var chats []*domain.Chat
	for _, member := range members {
		chat, err := s.GetChat(ctx, member.ChatID)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

func (s *ChatService) AddChatMember(ctx context.Context, chatID, userID string, role string) error {
	member := repository.ChatMember{
		ChatID:    chatID,
		UserID:    userID,
		Role:      repository.ChatRole(role),
		JoinedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.repo.AddChatMember(ctx, member)
}

func (s *ChatService) RemoveChatMember(ctx context.Context, chatID, userID string) error {
	return s.repo.RemoveChatMember(ctx, chatID, userID)
}

func (s *ChatService) UpdateChatMemberRole(ctx context.Context, chatID, userID string, role string) error {
	return s.repo.UpdateChatMemberRole(ctx, chatID, userID, repository.ChatRole(role))
}

func (s *ChatService) UpdateUserStatus(ctx context.Context, userID string, status string) error {
	userStatus := repository.UserStatus{
		UserID:    userID,
		Status:    status,
		LastSeen:  time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.repo.UpdateUserStatus(ctx, userStatus)
}

func (s *ChatService) GetUserStatus(ctx context.Context, userID string) (*domain.UserStatus, error) {
	status, err := s.repo.GetUserStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toDomainUserStatus(status), nil
}

func (s *ChatService) SaveNotification(ctx context.Context, notification *domain.Notification) error {
	repoNotification := toRepositoryNotification(notification)
	return s.repo.SaveNotification(ctx, repoNotification)
}

func (s *ChatService) GetUserNotifications(ctx context.Context, userID string, limit int) ([]*domain.Notification, error) {
	notifications, err := s.repo.GetUserNotifications(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	return toDomainNotifications(notifications), nil
}

func (s *ChatService) MarkNotificationAsRead(ctx context.Context, userID, notificationID string) error {
	return s.repo.MarkNotificationAsRead(ctx, userID, notificationID)
}
