package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(client *redis.Client) *CacheRepository {
	// Set default timeouts
	client.Options().ReadTimeout = 100 * time.Millisecond
	client.Options().WriteTimeout = 100 * time.Millisecond
	client.Options().PoolSize = 10
	client.Options().MinIdleConns = 5
	client.Options().PoolTimeout = 1 * time.Hour

	return &CacheRepository{
		client: client,
	}
}

func (r *CacheRepository) Get(key string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *CacheRepository) Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *CacheRepository) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	return r.client.Del(ctx, key).Err()
}
