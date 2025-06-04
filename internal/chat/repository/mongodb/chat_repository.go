package mongodb

import (
	"context"
	"fowergram/internal/domain/chat"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type chatRepository struct {
	db       *mongo.Database
	chats    *mongo.Collection
	messages *mongo.Collection
}

func NewChatRepository(db *mongo.Database) *chatRepository {
	return &chatRepository{
		db:       db,
		chats:    db.Collection("chats"),
		messages: db.Collection("messages"),
	}
}

func (r *chatRepository) GetChat(ctx context.Context, chatID string) (*domain.Chat, error) {
	var chat domain.Chat
	err := r.chats.FindOne(ctx, bson.M{"_id": chatID}).Decode(&chat)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *chatRepository) GetMessages(ctx context.Context, chatID string, limit, offset int) ([]domain.Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.messages.Find(ctx, bson.M{"chat_id": chatID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []domain.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *chatRepository) GetUserChats(ctx context.Context, userID string) ([]domain.Chat, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cursor, err := r.chats.Find(ctx, bson.M{"members": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chats []domain.Chat
	if err := cursor.All(ctx, &chats); err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *chatRepository) CreateChat(ctx context.Context, chat *domain.Chat) error {
	// Generate new ObjectID for the chat
	chat.ID = primitive.NewObjectID().Hex()

	_, err := r.chats.InsertOne(ctx, chat)
	return err
}

func (r *chatRepository) SaveMessage(ctx context.Context, message *domain.Message) error {
	_, err := r.messages.InsertOne(ctx, message)
	return err
}
