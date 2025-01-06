package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type cacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(client *redis.Client) *cacheRepository {
	return &cacheRepository{
		client: client,
	}
}

func (r *cacheRepository) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(context.Background(), key, data, ttl).Err()
}

func (r *cacheRepository) Get(key string) (interface{}, error) {
	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, err
	}

	return value, nil
}

func (r *cacheRepository) Delete(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
