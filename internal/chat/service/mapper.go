package service

import (
	"fowergram/internal/chat/domain"
	"fowergram/internal/chat/repository"
)

func toDomainChat(chat *repository.Chat) *domain.Chat {
	if chat == nil {
		return nil
	}
	return &domain.Chat{
		ID:        chat.ID,
		Name:      chat.Name,
		Type:      string(chat.Type),
		CreatedBy: chat.CreatedBy,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}
}

func toRepositoryChat(chat *domain.Chat) repository.Chat {
	return repository.Chat{
		ID:        chat.ID,
		Name:      chat.Name,
		Type:      repository.ChatType(chat.Type),
		CreatedBy: chat.CreatedBy,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}
}

func toDomainMessage(msg *repository.ChatMessage) *domain.Message {
	if msg == nil {
		return nil
	}
	return &domain.Message{
		ID:             msg.MessageID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		Type:           domain.MessageType(msg.Type),
		CreatedAt:      msg.CreatedAt,
	}
}

func toRepositoryMessage(msg *domain.Message) repository.ChatMessage {
	return repository.ChatMessage{
		ConversationID: msg.ConversationID,
		MessageID:      msg.ID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		Type:           repository.MessageType(msg.Type),
		CreatedAt:      msg.CreatedAt,
	}
}

func toDomainMessages(messages []repository.ChatMessage) []*domain.Message {
	result := make([]*domain.Message, len(messages))
	for i, msg := range messages {
		msg := msg // Create a new variable to avoid issues with the loop variable
		result[i] = toDomainMessage(&msg)
	}
	return result
}

func toDomainChatMember(member *repository.ChatMember) *domain.ChatMember {
	if member == nil {
		return nil
	}
	return &domain.ChatMember{
		ChatID:    member.ChatID,
		UserID:    member.UserID,
		Role:      string(member.Role),
		JoinedAt:  member.JoinedAt,
		UpdatedAt: member.UpdatedAt,
	}
}

func toRepositoryChatMember(member *domain.ChatMember) repository.ChatMember {
	return repository.ChatMember{
		ChatID:    member.ChatID,
		UserID:    member.UserID,
		Role:      repository.ChatRole(member.Role),
		JoinedAt:  member.JoinedAt,
		UpdatedAt: member.UpdatedAt,
	}
}

func toDomainChatMembers(members []repository.ChatMember) []*domain.ChatMember {
	result := make([]*domain.ChatMember, len(members))
	for i, member := range members {
		member := member // Create a new variable to avoid issues with the loop variable
		result[i] = toDomainChatMember(&member)
	}
	return result
}

func toDomainUserStatus(status *repository.UserStatus) *domain.UserStatus {
	if status == nil {
		return nil
	}
	return &domain.UserStatus{
		UserID:    status.UserID,
		Status:    status.Status,
		LastSeen:  status.LastSeen,
		UpdatedAt: status.UpdatedAt,
	}
}

func toRepositoryUserStatus(status *domain.UserStatus) repository.UserStatus {
	return repository.UserStatus{
		UserID:    status.UserID,
		Status:    status.Status,
		LastSeen:  status.LastSeen,
		UpdatedAt: status.UpdatedAt,
	}
}

func toDomainNotification(notification *repository.Notification) *domain.Notification {
	if notification == nil {
		return nil
	}
	return &domain.Notification{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Content:   notification.Content,
		Read:      notification.Read,
		CreatedAt: notification.CreatedAt,
	}
}

func toRepositoryNotification(notification *domain.Notification) repository.Notification {
	return repository.Notification{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Content:   notification.Content,
		Read:      notification.Read,
		CreatedAt: notification.CreatedAt,
	}
}

func toDomainNotifications(notifications []repository.Notification) []*domain.Notification {
	result := make([]*domain.Notification, len(notifications))
	for i, notification := range notifications {
		notification := notification // Create a new variable to avoid issues with the loop variable
		result[i] = toDomainNotification(&notification)
	}
	return result
}
