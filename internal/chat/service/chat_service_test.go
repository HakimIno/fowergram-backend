package service

import (
	"context"
	"testing"
	"time"

	"fowergram/internal/chat/domain"
	"fowergram/internal/chat/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChatRepository is a mock implementation of repository.ChatRepository
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) CreateChat(ctx context.Context, chat repository.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockChatRepository) GetChat(ctx context.Context, chatID string) (*repository.Chat, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Chat), args.Error(1)
}

func (m *MockChatRepository) GetChatMembers(ctx context.Context, chatID string) ([]repository.ChatMember, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).([]repository.ChatMember), args.Error(1)
}

func (m *MockChatRepository) AddChatMember(ctx context.Context, member repository.ChatMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockChatRepository) RemoveChatMember(ctx context.Context, chatID, userID string) error {
	args := m.Called(ctx, chatID, userID)
	return args.Error(0)
}

func (m *MockChatRepository) UpdateChatMemberRole(ctx context.Context, chatID, userID string, role repository.ChatRole) error {
	args := m.Called(ctx, chatID, userID, role)
	return args.Error(0)
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, msg repository.ChatMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockChatRepository) GetMessages(ctx context.Context, conversationID string, limit int) ([]repository.ChatMessage, error) {
	args := m.Called(ctx, conversationID, limit)
	return args.Get(0).([]repository.ChatMessage), args.Error(1)
}

func (m *MockChatRepository) UpdateUserStatus(ctx context.Context, status repository.UserStatus) error {
	args := m.Called(ctx, status)
	return args.Error(0)
}

func (m *MockChatRepository) GetUserStatus(ctx context.Context, userID string) (*repository.UserStatus, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.UserStatus), args.Error(1)
}

func (m *MockChatRepository) SaveNotification(ctx context.Context, notification repository.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockChatRepository) GetUserNotifications(ctx context.Context, userID string, limit int) ([]repository.Notification, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]repository.Notification), args.Error(1)
}

func (m *MockChatRepository) MarkNotificationAsRead(ctx context.Context, userID, notificationID string) error {
	args := m.Called(ctx, userID, notificationID)
	return args.Error(0)
}

func (m *MockChatRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockChatRepository) GetUserChats(ctx context.Context, userID string) ([]repository.ChatMember, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]repository.ChatMember), args.Error(1)
}

// MockMessageProducer is a mock implementation of broker.MessageProducer
type MockMessageProducer struct {
	mock.Mock
}

func (m *MockMessageProducer) ProduceMessage(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func TestChatService_HandleMessage(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	msg := &domain.Message{
		ID:             "msg1",
		ConversationID: "chat1",
		SenderID:       "user1",
		Content:        "Hello, World!",
		Type:           domain.TextMessage,
		CreatedAt:      time.Now(),
	}

	// Set up expectations
	mockRepo.On("SaveMessage", ctx, mock.AnythingOfType("repository.ChatMessage")).Return(nil)
	mockProducer.On("ProduceMessage", ctx, msg).Return(nil)

	// Execute test
	err := service.HandleMessage(ctx, msg)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestChatService_CreateChat(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	chat := &domain.Chat{
		ID:        "chat1",
		Name:      "Test Chat",
		Type:      string(repository.DirectChat),
		CreatedBy: "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Members:   []string{"user1", "user2"},
	}

	// Set up expectations
	mockRepo.On("CreateChat", ctx, mock.AnythingOfType("repository.Chat")).Return(nil)
	mockRepo.On("AddChatMember", ctx, mock.AnythingOfType("repository.ChatMember")).Return(nil).Times(2)

	// Execute test
	err := service.CreateChat(ctx, chat)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChatService_GetMessages(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	chatID := "chat1"
	limit := 10
	messages := []repository.ChatMessage{
		{
			ConversationID: chatID,
			MessageID:      "msg1",
			SenderID:       "user1",
			Content:        "Hello",
			Type:           repository.TextMessage,
			CreatedAt:      time.Now(),
		},
		{
			ConversationID: chatID,
			MessageID:      "msg2",
			SenderID:       "user2",
			Content:        "Hi",
			Type:           repository.TextMessage,
			CreatedAt:      time.Now(),
		},
	}

	// Set up expectations
	mockRepo.On("GetMessages", ctx, chatID, limit).Return(messages, nil)

	// Execute test
	result, err := service.GetMessages(ctx, chatID, limit)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, messages[0].MessageID, result[0].ID)
	assert.Equal(t, messages[1].MessageID, result[1].ID)
	mockRepo.AssertExpectations(t)
}

func TestChatService_GetChat(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	chatID := "chat1"
	chat := &repository.Chat{
		ID:        chatID,
		Name:      "Test Chat",
		Type:      repository.DirectChat,
		CreatedBy: "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	members := []repository.ChatMember{
		{
			ChatID:    chatID,
			UserID:    "user1",
			Role:      repository.ChatRoleMember,
			JoinedAt:  time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ChatID:    chatID,
			UserID:    "user2",
			Role:      repository.ChatRoleMember,
			JoinedAt:  time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Set up expectations
	mockRepo.On("GetChat", ctx, chatID).Return(chat, nil)
	mockRepo.On("GetChatMembers", ctx, chatID).Return(members, nil)

	// Execute test
	result, err := service.GetChat(ctx, chatID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, chat.ID, result.ID)
	assert.Equal(t, chat.Name, result.Name)
	assert.Len(t, result.Members, 2)
	mockRepo.AssertExpectations(t)
}

func TestChatService_HandleMessage_Error(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	msg := &domain.Message{
		ID:             "msg1",
		ConversationID: "chat1",
		SenderID:       "user1",
		Content:        "Hello, World!",
		Type:           domain.TextMessage,
		CreatedAt:      time.Now(),
	}

	// Test case 1: Repository error
	mockRepo.On("SaveMessage", ctx, mock.AnythingOfType("repository.ChatMessage")).Return(assert.AnError).Once()
	err := service.HandleMessage(ctx, msg)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)

	// Test case 2: Producer error
	mockRepo.On("SaveMessage", ctx, mock.AnythingOfType("repository.ChatMessage")).Return(nil).Once()
	mockProducer.On("ProduceMessage", ctx, msg).Return(assert.AnError).Once()
	err = service.HandleMessage(ctx, msg)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestChatService_CreateChat_Error(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	chat := &domain.Chat{
		ID:        "chat1",
		Name:      "Test Chat",
		Type:      string(repository.DirectChat),
		CreatedBy: "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Members:   []string{"user1", "user2"},
	}

	// Test case: Repository error
	mockRepo.On("CreateChat", ctx, mock.AnythingOfType("repository.Chat")).Return(assert.AnError).Once()
	err := service.CreateChat(ctx, chat)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChatService_GetChat_NotFound(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	chatID := "non_existent_chat"

	// Set up expectations
	mockRepo.On("GetChat", ctx, chatID).Return(nil, nil)
	mockRepo.On("GetChatMembers", ctx, chatID).Return([]repository.ChatMember{}, nil)

	// Execute test
	result, err := service.GetChat(ctx, chatID)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestChatService_UpdateUserStatus(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	userID := "user1"
	status := "online"

	// Set up expectations
	mockRepo.On("UpdateUserStatus", ctx, mock.MatchedBy(func(s repository.UserStatus) bool {
		return s.UserID == userID && s.Status == status
	})).Return(nil)

	// Execute test
	err := service.UpdateUserStatus(ctx, userID, status)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChatService_GetUserChats(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	userID := "user1"
	members := []repository.ChatMember{
		{
			ChatID:    "chat1",
			UserID:    userID,
			Role:      repository.ChatRoleMember,
			JoinedAt:  time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ChatID:    "chat2",
			UserID:    userID,
			Role:      repository.ChatRoleMember,
			JoinedAt:  time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	chat1 := &repository.Chat{
		ID:        "chat1",
		Name:      "Chat 1",
		Type:      repository.DirectChat,
		CreatedBy: "user2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	chat2 := &repository.Chat{
		ID:        "chat2",
		Name:      "Chat 2",
		Type:      repository.GroupChat,
		CreatedBy: "user3",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set up expectations
	mockRepo.On("GetChatMembers", ctx, userID).Return(members, nil)
	mockRepo.On("GetChat", ctx, "chat1").Return(chat1, nil)
	mockRepo.On("GetChatMembers", ctx, "chat1").Return([]repository.ChatMember{members[0]}, nil)
	mockRepo.On("GetChat", ctx, "chat2").Return(chat2, nil)
	mockRepo.On("GetChatMembers", ctx, "chat2").Return([]repository.ChatMember{members[1]}, nil)

	// Execute test
	chats, err := service.GetUserChats(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, chats, 2)
	assert.Equal(t, chat1.ID, chats[0].ID)
	assert.Equal(t, chat2.ID, chats[1].ID)
	mockRepo.AssertExpectations(t)
}

func TestChatService_SaveNotification(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	notification := &domain.Notification{
		ID:        "notif1",
		UserID:    "user1",
		Type:      "message",
		Content:   "New message from user2",
		Read:      false,
		CreatedAt: time.Now(),
	}

	// Set up expectations
	mockRepo.On("SaveNotification", ctx, mock.AnythingOfType("repository.Notification")).Return(nil)

	// Execute test
	err := service.SaveNotification(ctx, notification)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChatService_GetUserNotifications(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	userID := "user1"
	limit := 10
	notifications := []repository.Notification{
		{
			ID:        "notif1",
			UserID:    userID,
			Type:      "message",
			Content:   "New message from user2",
			Read:      false,
			CreatedAt: time.Now(),
		},
		{
			ID:        "notif2",
			UserID:    userID,
			Type:      "friend_request",
			Content:   "New friend request from user3",
			Read:      false,
			CreatedAt: time.Now(),
		},
	}

	// Set up expectations
	mockRepo.On("GetUserNotifications", ctx, userID, limit).Return(notifications, nil)

	// Execute test
	result, err := service.GetUserNotifications(ctx, userID, limit)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, notifications[0].ID, result[0].ID)
	assert.Equal(t, notifications[1].ID, result[1].ID)
	mockRepo.AssertExpectations(t)
}

func TestChatService_RemoveChatMember(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	chatID := "chat1"
	userID := "user1"

	// Set up expectations
	mockRepo.On("RemoveChatMember", ctx, chatID, userID).Return(nil)

	// Execute test
	err := service.RemoveChatMember(ctx, chatID, userID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChatService_UpdateChatMemberRole(t *testing.T) {
	// Create mocks
	mockRepo := new(MockChatRepository)
	mockProducer := new(MockMessageProducer)
	wsManager := NewWebSocketManager()

	// Create service with mocks
	service := NewChatService(mockRepo, wsManager, mockProducer)

	// Test data
	ctx := context.Background()
	chatID := "chat1"
	userID := "user1"
	role := string(repository.ChatRoleAdmin)

	// Set up expectations
	mockRepo.On("UpdateChatMemberRole", ctx, chatID, userID, repository.ChatRole(role)).Return(nil)

	// Execute test
	err := service.UpdateChatMemberRole(ctx, chatID, userID, role)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
