package redis

import (
	"context"
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
	return r.client.Set(context.Background(), key, value, ttl).Err()
}

func (r *cacheRepository) Get(key string) (interface{}, error) {
	return r.client.Get(context.Background(), key).Result()
}

func (r *cacheRepository) Delete(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
