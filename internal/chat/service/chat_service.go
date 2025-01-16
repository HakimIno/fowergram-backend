package service

import (
	"context"
	"fmt"
	"math/rand"
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
		return fmt.Errorf("failed to create chat in repository: %w", err)
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
			return fmt.Errorf("failed to add chat member %s: %w", userID, err)
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
	members, err := s.repo.GetUserChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	var chats []*domain.Chat
	for _, member := range members {
		chat, err := s.GetChat(ctx, member.ChatID)
		if err != nil {
			return nil, err
		}
		if chat != nil {
			chats = append(chats, chat)
		}
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

func (s *ChatService) CreateInviteLink(ctx context.Context, chatID string, createdBy string, maxUses int, expiresIn time.Duration) (*domain.ChatInviteLink, error) {
	// Generate a unique code
	code := generateInviteCode()

	link := &repository.ChatInviteLink{
		ChatID:    chatID,
		Code:      code,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expiresIn),
		MaxUses:   maxUses,
		Uses:      0,
	}

	if err := s.repo.CreateInviteLink(ctx, *link); err != nil {
		return nil, fmt.Errorf("failed to create invite link: %w", err)
	}

	return toDomainInviteLink(link), nil
}

func (s *ChatService) GetInviteLinkByCode(ctx context.Context, code string) (*domain.ChatInviteLink, error) {
	link, err := s.repo.GetInviteLinkByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite link: %w", err)
	}
	if link == nil {
		return nil, nil
	}
	return toDomainInviteLink(link), nil
}

func (s *ChatService) GetChatInviteLinks(ctx context.Context, chatID string) ([]*domain.ChatInviteLink, error) {
	links, err := s.repo.GetChatInviteLinks(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat invite links: %w", err)
	}
	return toDomainInviteLinks(links), nil
}

func (s *ChatService) JoinChatViaInvite(ctx context.Context, code string, userID string) (*domain.Chat, error) {
	link, err := s.repo.GetInviteLinkByCode(ctx, code)
	if err != nil {
		fmt.Printf("Error getting invite link: %v\n", err)
		return nil, fmt.Errorf("failed to get invite link: %w", err)
	}
	if link == nil {
		fmt.Printf("Invite link not found for code: %s\n", code)
		return nil, fmt.Errorf("invite link not found")
	}

	// Check if link is expired
	if time.Now().After(link.ExpiresAt) {
		fmt.Printf("Invite link expired: %s\n", code)
		return nil, fmt.Errorf("invite link has expired")
	}

	// Check if max uses reached
	if link.MaxUses > 0 && link.Uses >= link.MaxUses {
		fmt.Printf("Invite link max uses reached: %s\n", code)
		return nil, fmt.Errorf("invite link has reached maximum uses")
	}

	// Get chat
	chat, err := s.GetChat(ctx, link.ChatID)
	if err != nil {
		fmt.Printf("Error getting chat: %v\n", err)
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		fmt.Printf("Chat not found: %s\n", link.ChatID)
		return nil, fmt.Errorf("chat not found")
	}

	// Check if user is already a member
	for _, member := range chat.Members {
		if member == userID {
			fmt.Printf("User already a member: %s\n", userID)
			return chat, nil
		}
	}

	// Add user to chat
	member := &domain.ChatMember{
		ChatID:    chat.ID,
		UserID:    userID,
		Role:      string(domain.ChatRoleMember),
		JoinedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}
	repoMember := toRepositoryChatMember(member)
	if err := s.repo.AddChatMember(ctx, repoMember); err != nil {
		fmt.Printf("Error adding chat member: %v\n", err)
		return nil, fmt.Errorf("failed to add chat member: %w", err)
	}

	// Increment invite link uses
	if err := s.repo.IncrementInviteLinkUses(ctx, link.ChatID, link.Code); err != nil {
		fmt.Printf("Error incrementing invite link uses: %v\n", err)
		return nil, fmt.Errorf("failed to increment invite link uses: %w", err)
	}

	fmt.Printf("Successfully joined chat: %s, user: %s\n", chat.ID, userID)

	// Get updated chat
	return s.GetChat(ctx, link.ChatID)
}

func (s *ChatService) DeleteInviteLink(ctx context.Context, chatID string, code string) error {
	return s.repo.DeleteInviteLink(ctx, chatID, code)
}

func generateInviteCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 10
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}
