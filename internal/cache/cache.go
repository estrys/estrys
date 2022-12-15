package cache

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type Option any
type OptionDefaultTTL time.Duration

type Error interface {
	error
}

var ErrMiss = errors.New("cache key not found")

//go:generate mockery --with-expecter --name Cache
type Cache[T any] interface {
	Set(ctx context.Context, key string, value T, opts ...Option) error
	Get(ctx context.Context, key string) (*T, error)
}
