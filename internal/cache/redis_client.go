package cache

import (
	"context"

	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(opts *redis.Options) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(opts),
	}
}

func (r RedisClient) Client() *redis.Client {
	return r.client
}

func (r RedisClient) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return errors.Wrap(err, "unable to reach redis")
	}
	return nil
}
