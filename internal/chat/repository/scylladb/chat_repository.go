package scylladb

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"

	"fowergram/internal/chat/repository"
)

var chatMessagesTable = table.New(table.Metadata{
	Name:    "chat_messages",
	Columns: []string{"conversation_id", "message_id", "sender_id", "content", "type", "created_at"},
	PartKey: []string{"conversation_id"},
	SortKey: []string{"created_at", "message_id"},
})

var userStatusTable = table.New(table.Metadata{
	Name:    "user_status",
	Columns: []string{"user_id", "status", "last_seen", "updated_at"},
	PartKey: []string{"user_id"},
})

var notificationsTable = table.New(table.Metadata{
	Name:    "notifications",
	Columns: []string{"user_id", "id", "type", "content", "read", "created_at"},
	PartKey: []string{"user_id"},
	SortKey: []string{"created_at"},
})

var chatsTable = table.New(table.Metadata{
	Name:    "chats",
	Columns: []string{"id", "name", "type", "created_by", "created_at", "updated_at"},
	PartKey: []string{"id"},
})

var chatMembersTable = table.New(table.Metadata{
	Name:    "chat_members",
	Columns: []string{"chat_id", "user_id", "role", "joined_at", "updated_at"},
	PartKey: []string{"chat_id"},
	SortKey: []string{"user_id"},
})

type ChatRepository struct {
	session *gocqlx.Session
}

func NewChatRepository(hosts []string, keyspace string) (*ChatRepository, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &ChatRepository{
		session: &session,
	}, nil
}

func (r *ChatRepository) CreateChat(ctx context.Context, chat repository.Chat) error {
	stmt, names := chatsTable.Insert()
	q := r.session.Query(stmt, names).BindStruct(chat)
	return q.ExecRelease()
}

func (r *ChatRepository) GetChat(ctx context.Context, chatID string) (*repository.Chat, error) {
	var chat repository.Chat
	stmt, names := chatsTable.Get()
	q := r.session.Query(stmt, names).BindMap(qb.M{"id": chatID})
	if err := q.GetRelease(&chat); err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &chat, nil
}

func (r *ChatRepository) GetChatMembers(ctx context.Context, chatID string) ([]repository.ChatMember, error) {
	var members []repository.ChatMember
	stmt, names := chatMembersTable.Select()
	q := r.session.Query(stmt, names).BindMap(qb.M{"chat_id": chatID})
	if err := q.SelectRelease(&members); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *ChatRepository) AddChatMember(ctx context.Context, member repository.ChatMember) error {
	stmt, names := chatMembersTable.Insert()
	q := r.session.Query(stmt, names).BindStruct(member)
	return q.ExecRelease()
}

func (r *ChatRepository) RemoveChatMember(ctx context.Context, chatID, userID string) error {
	stmt, names := chatMembersTable.Delete()
	q := r.session.Query(stmt, names).BindMap(qb.M{
		"chat_id": chatID,
		"user_id": userID,
	})
	return q.ExecRelease()
}

func (r *ChatRepository) UpdateChatMemberRole(ctx context.Context, chatID, userID string, role repository.ChatRole) error {
	stmt, names := qb.Update("chat_members").
		Set("role", "updated_at").
		Where(qb.Eq("chat_id"), qb.Eq("user_id")).
		ToCql()

	return r.session.Query(stmt, names).BindMap(qb.M{
		"chat_id":    chatID,
		"user_id":    userID,
		"role":       role,
		"updated_at": time.Now(),
	}).ExecRelease()
}

func (r *ChatRepository) SaveMessage(ctx context.Context, msg repository.ChatMessage) error {
	stmt, names := chatMessagesTable.Insert()
	q := r.session.Query(stmt, names).BindStruct(msg)
	return q.ExecRelease()
}

func (r *ChatRepository) GetMessages(ctx context.Context, conversationID string, limit int) ([]repository.ChatMessage, error) {
	var messages []repository.ChatMessage
	stmt, names := qb.Select("chat_messages").
		Where(qb.Eq("conversation_id")).
		OrderBy("created_at", qb.DESC).
		Limit(uint(limit)).
		ToCql()

	q := r.session.Query(stmt, names).BindMap(qb.M{
		"conversation_id": conversationID,
	})
	if err := q.SelectRelease(&messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *ChatRepository) UpdateUserStatus(ctx context.Context, status repository.UserStatus) error {
	stmt, names := userStatusTable.Insert()
	q := r.session.Query(stmt, names).BindStruct(status)
	return q.ExecRelease()
}

func (r *ChatRepository) GetUserStatus(ctx context.Context, userID string) (*repository.UserStatus, error) {
	var status repository.UserStatus
	stmt, names := userStatusTable.Get()
	q := r.session.Query(stmt, names).BindMap(qb.M{"user_id": userID})
	if err := q.GetRelease(&status); err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &status, nil
}

func (r *ChatRepository) SaveNotification(ctx context.Context, notification repository.Notification) error {
	stmt, names := notificationsTable.Insert()
	q := r.session.Query(stmt, names).BindStruct(notification)
	return q.ExecRelease()
}

func (r *ChatRepository) GetUserNotifications(ctx context.Context, userID string, limit int) ([]repository.Notification, error) {
	var notifications []repository.Notification
	stmt, names := qb.Select("notifications").
		Where(qb.Eq("user_id")).
		OrderBy("created_at", qb.DESC).
		Limit(uint(limit)).
		ToCql()

	q := r.session.Query(stmt, names).BindMap(qb.M{
		"user_id": userID,
	})
	if err := q.SelectRelease(&notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *ChatRepository) MarkNotificationAsRead(ctx context.Context, userID, notificationID string) error {
	stmt, names := qb.Update("notifications").
		Set("read").
		Where(qb.Eq("user_id"), qb.Eq("id")).
		ToCql()

	return r.session.Query(stmt, names).BindMap(qb.M{
		"user_id": userID,
		"id":      notificationID,
		"read":    true,
	}).ExecRelease()
}

func (r *ChatRepository) Close() error {
	r.session.Close()
	return nil
}
