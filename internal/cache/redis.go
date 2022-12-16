package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/estrys/estrys/internal/observability"
)

type redisCache[T any] struct {
	client         *redis.Client
	defaultOptions []Option
}

func CreateRedisCache[T any](redisClient *RedisClient, defaultOpts ...Option) Cache[T] {
	return &redisCache[T]{
		defaultOptions: defaultOpts,
		client:         redisClient.Client(),
	}
}

func (r redisCache[T]) Set(ctx context.Context, key string, value T, opts ...Option) error {
	serializedValue, err := msgpack.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "unable to serialize cache value using msgpack")
	}

	var timeout time.Duration
	for _, v := range r.defaultOptions {
		if t, ok := any(v).(OptionDefaultTTL); ok {
			timeout = time.Duration(t)
		}
	}
	for _, v := range opts {
		if t, ok := any(v).(OptionDefaultTTL); ok {
			timeout = time.Duration(t)
		}
	}
	span := observability.StartSpan(ctx, "cache.save", map[string]any{"db.system": "redis", "cache.key": key})
	r.client.Set(ctx, key, serializedValue, timeout)
	observability.FinishSpan(span)
	return nil
}

func (r redisCache[T]) Get(ctx context.Context, key string) (*T, error) {
	span := observability.StartSpan(ctx, "cache.get_item", map[string]any{"db.system": "redis", "cache.key": key})
	var result T
	item, err := r.client.Get(ctx, key).Result()
	observability.FinishSpan(span)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrMiss
		}
		return nil, errors.Wrap(err, "redis get key error")
	}
	err = msgpack.Unmarshal([]byte(item), &result)
	if err != nil {
		return nil, errors.Wrap(err, "unable to deserialize cached value")
	}
	return &result, nil
}
