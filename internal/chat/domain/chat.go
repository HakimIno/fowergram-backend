package domain

import (
	"time"
)

type MessageType string

const (
	TextMessage    MessageType = "text"
	ImageMessage   MessageType = "image"
	StickerMessage MessageType = "sticker"
)

type Message struct {
	ID        string      `json:"id" bson:"_id,omitempty"`
	ChatID    string      `json:"chat_id" bson:"chat_id"`
	SenderID  string      `json:"sender_id" bson:"sender_id"`
	Content   string      `json:"content" bson:"content"`
	Type      MessageType `json:"type" bson:"type"`
	MediaURL  string      `json:"media_url,omitempty" bson:"media_url,omitempty"`
	CreatedAt time.Time   `json:"created_at" bson:"created_at"`
}

type Chat struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	Members     []string  `json:"members" bson:"members"`
	LastMessage *Message  `json:"last_message,omitempty" bson:"last_message,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}
