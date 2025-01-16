package scylladb

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type ChatMessage struct {
	ConversationID string    `json:"conversation_id"`
	MessageID      string    `json:"message_id"`
	SenderID       string    `json:"sender_id"`
	Content        string    `json:"content"`
	Type           string    `json:"type"`
	CreatedAt      time.Time `json:"created_at"`
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

var (
	chatMessageMetadata = table.Metadata{
		Name:    "chat_messages",
		Columns: []string{"conversation_id", "message_id", "sender_id", "content", "type", "created_at"},
		PartKey: []string{"conversation_id"},
		SortKey: []string{"created_at", "message_id"},
	}

	userStatusMetadata = table.Metadata{
		Name:    "user_status",
		Columns: []string{"user_id", "status", "last_seen", "updated_at"},
		PartKey: []string{"user_id"},
	}

	notificationMetadata = table.Metadata{
		Name:    "notifications",
		Columns: []string{"id", "user_id", "type", "content", "read", "created_at"},
		PartKey: []string{"user_id"},
		SortKey: []string{"created_at"},
	}

	chatMessagesTable  = table.New(chatMessageMetadata)
	userStatusTable    = table.New(userStatusMetadata)
	notificationsTable = table.New(notificationMetadata)
)

type ChatRepository struct {
	session gocqlx.Session
}

func NewChatRepository(hosts []string, keyspace string) (*ChatRepository, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		return nil, err
	}

	return &ChatRepository{
		session: session,
	}, nil
}

func (r *ChatRepository) SaveMessage(ctx context.Context, msg *ChatMessage) error {
	return r.session.Query(chatMessagesTable.Insert()).
		BindStruct(msg).
		ExecRelease()
}

func (r *ChatRepository) GetMessages(ctx context.Context, conversationID string, limit int, before time.Time) ([]*ChatMessage, error) {
	var messages []*ChatMessage
	stmt := qb.Select(chatMessagesTable.Name()).
		Where(qb.Eq("conversation_id"), qb.Lt("created_at")).
		Limit(uint(limit))

	q := stmt.Query(r.session)
	err := q.BindMap(qb.M{
		"conversation_id": conversationID,
		"created_at":      before,
	}).SelectRelease(&messages)

	return messages, err
}

func (r *ChatRepository) UpdateUserStatus(ctx context.Context, status *UserStatus) error {
	return r.session.Query(userStatusTable.Insert()).
		BindStruct(status).
		ExecRelease()
}

func (r *ChatRepository) GetUserStatus(ctx context.Context, userID string) (*UserStatus, error) {
	var status UserStatus
	err := r.session.Query(userStatusTable.Get()).
		BindMap(qb.M{"user_id": userID}).
		GetRelease(&status)

	return &status, err
}

func (r *ChatRepository) SaveNotification(ctx context.Context, notification *Notification) error {
	return r.session.Query(notificationsTable.Insert()).
		BindStruct(notification).
		ExecRelease()
}

func (r *ChatRepository) GetUserNotifications(ctx context.Context, userID string, limit int) ([]*Notification, error) {
	var notifications []*Notification
	stmt := qb.Select(notificationsTable.Name()).
		Where(qb.Eq("user_id")).
		Limit(uint(limit))

	q := stmt.Query(r.session)
	err := q.BindMap(qb.M{"user_id": userID}).
		SelectRelease(&notifications)

	return notifications, err
}

func (r *ChatRepository) MarkNotificationAsRead(ctx context.Context, userID string, notificationID string) error {
	stmt := qb.Update(notificationsTable.Name()).
		Set("read").
		Where(qb.Eq("user_id"), qb.Eq("id"))

	return stmt.Query(r.session).
		BindMap(qb.M{
			"user_id": userID,
			"id":      notificationID,
			"read":    true,
		}).ExecRelease()
}

func (r *ChatRepository) Close() {
	r.session.Close()
}
