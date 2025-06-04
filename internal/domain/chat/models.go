package domain

import "time"

type MessageType string

const (
	TextMessage     MessageType = "text"
	ImageMessage    MessageType = "image"
	VideoMessage    MessageType = "video"
	DocumentMessage MessageType = "document"
)

type Message struct {
	ID             string      `json:"id"`
	ConversationID string      `json:"conversation_id"`
	SenderID       string      `json:"sender_id"`
	Content        string      `json:"content"`
	Type           MessageType `json:"type"`
	CreatedAt      time.Time   `json:"created_at"`
}

type Chat struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Members   []string  `json:"members"`
	IsPrivate bool      `json:"is_private"`
}

type ChatMember struct {
	ChatID    string    `json:"chat_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
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
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatInviteLink struct {
	ChatID    string    `json:"chat_id"`
	Code      string    `json:"code"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	MaxUses   int       `json:"max_uses"`
	Uses      int       `json:"uses"`
}
