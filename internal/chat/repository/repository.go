package repository

import (
	"context"
	"time"
)

type MessageType string
type ChatType string
type ChatRole string

const (
	TextMessage     MessageType = "text"
	ImageMessage    MessageType = "image"
	VideoMessage    MessageType = "video"
	DocumentMessage MessageType = "document"
)

const (
	DirectChat    ChatType = "direct"
	GroupChat     ChatType = "group"
	BroadcastChat ChatType = "broadcast"
)

const (
	ChatRoleOwner  ChatRole = "owner"
	ChatRoleMember ChatRole = "member"
	ChatRoleAdmin  ChatRole = "admin"
)

type ChatMessage struct {
	ConversationID string      `json:"conversation_id"`
	MessageID      string      `json:"message_id"`
	SenderID       string      `json:"sender_id"`
	Content        string      `json:"content"`
	Type           MessageType `json:"type"`
	CreatedAt      time.Time   `json:"created_at"`
}

type Chat struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      ChatType  `json:"type"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMember struct {
	ChatID    string    `json:"chat_id"`
	UserID    string    `json:"user_id"`
	Role      ChatRole  `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserStatus struct {
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"last_seen"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Notification struct {
	UserID    string    `json:"user_id"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatRepository interface {
	CreateChat(ctx context.Context, chat Chat) error
	GetChat(ctx context.Context, chatID string) (*Chat, error)
	GetChatMembers(ctx context.Context, chatID string) ([]ChatMember, error)
	AddChatMember(ctx context.Context, member ChatMember) error
	RemoveChatMember(ctx context.Context, chatID, userID string) error
	UpdateChatMemberRole(ctx context.Context, chatID, userID string, role ChatRole) error
	SaveMessage(ctx context.Context, msg ChatMessage) error
	GetMessages(ctx context.Context, conversationID string, limit int) ([]ChatMessage, error)
	UpdateUserStatus(ctx context.Context, status UserStatus) error
	GetUserStatus(ctx context.Context, userID string) (*UserStatus, error)
	SaveNotification(ctx context.Context, notification Notification) error
	GetUserNotifications(ctx context.Context, userID string, limit int) ([]Notification, error)
	MarkNotificationAsRead(ctx context.Context, userID, notificationID string) error
	Close() error
}
