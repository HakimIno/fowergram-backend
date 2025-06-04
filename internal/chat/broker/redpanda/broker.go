package redpanda

import (
	"context"
	"encoding/json"
	domain "fowergram/internal/domain/chat"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	topicMessages = "chat.messages"
)

type RedpandaBroker struct {
	client *kgo.Client
}

func NewRedpandaBroker(brokers []string) (*RedpandaBroker, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup("chat-service"),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return nil, err
	}

	return &RedpandaBroker{
		client: client,
	}, nil
}

func (b *RedpandaBroker) ProduceMessage(ctx context.Context, message *domain.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	record := &kgo.Record{
		Topic: topicMessages,
		Key:   []byte(message.ConversationID),
		Value: data,
	}

	return b.client.ProduceSync(ctx, record).FirstErr()
}

func (b *RedpandaBroker) Subscribe(ctx context.Context, handler func(*domain.Message)) error {
	go func() {
		for {
			fetches := b.client.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				log.Printf("errors polling: %v", errs)
				continue
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()
				var message domain.Message
				if err := json.Unmarshal(record.Value, &message); err != nil {
					log.Printf("error unmarshaling message: %v", err)
					continue
				}
				handler(&message)
			}
		}
	}()
	return nil
}

func (b *RedpandaBroker) Close() error {
	b.client.Close()
	return nil
}
