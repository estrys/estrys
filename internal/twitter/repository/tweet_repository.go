package repository

import (
	"context"
	"strings"

	"github.com/estrys/estrys/internal/cache"
	"github.com/estrys/estrys/internal/twitter/models"
)

//go:generate mockery --with-expecter --name=TweetRepository
type TweetRepository interface {
	GetTweet(context.Context, string) (*models.Tweet, error)
	Store(context.Context, *models.Tweet) error
}

type redisTweetRepository struct {
	cache cache.Cache[models.Tweet]
}

func NewRedisTweetRepository(cache cache.Cache[models.Tweet]) *redisTweetRepository {
	return &redisTweetRepository{
		cache: cache,
	}
}

func (r *redisTweetRepository) getTweetCacheKey(id string) string {
	return strings.Join([]string{"twitter", "tweets", id}, "/")
}

func (r *redisTweetRepository) GetTweet(ctx context.Context, tweetID string) (*models.Tweet, error) {
	tweet, err := r.cache.Get(ctx, r.getTweetCacheKey(tweetID))
	if err != nil {
		return nil, err
	}
	return tweet, nil
}

func (r *redisTweetRepository) Store(ctx context.Context, tweet *models.Tweet) error {
	err := r.cache.Set(ctx, r.getTweetCacheKey(tweet.ID), *tweet)
	if err != nil {
		return err
	}
	return nil
}
