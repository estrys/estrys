package twitter

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/cache"
	"github.com/estrys/estrys/internal/logger"
)

type Authorizer struct {
	Token string
}

func (a Authorizer) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

const (
	TwitterErrorTypeNotFound = "https://api.twitter.com/2/problems/resource-not-found"

	cacheKeyUsername = "twitter/user/by-username/%s"
	cacheKeyID       = "twitter/user/by-id/%s"
)

type UserNotFoundError struct {
	Username string
}

func (u UserNotFoundError) Error() string {
	return fmt.Sprintf("twitter user %s not found", u.Username)
}

//go:generate mockery --with-expecter --name=Backend
type Backend interface {
	UserNameLookup(
		ctx context.Context,
		usernames []string,
		opts twitter.UserLookupOpts,
	) (*twitter.UserLookupResponse, error)
	UserTweetTimeline(
		ctx context.Context,
		userID string,
		opts twitter.UserTweetTimelineOpts,
	) (*twitter.UserTweetTimelineResponse, error)
	TweetLookup(ctx context.Context, ids []string, opts twitter.TweetLookupOpts) (*twitter.TweetLookupResponse, error)
	UserLookup(ctx context.Context, ids []string, opts twitter.UserLookupOpts) (*twitter.UserLookupResponse, error)
}

//go:generate mockery --with-expecter --name=TwitterClient
type TwitterClient interface {
	GetUserTweets(context.Context, string, twitter.UserTweetTimelineOpts) (*twitter.UserTweetTimelineResponse, error)
	GetTweets(context.Context, []string, twitter.TweetLookupOpts) (*twitter.TweetLookupResponse, error)
	GetUser(ctx context.Context, username string) (*twitter.UserObj, error)
	GetUserByIDs(context.Context, []string) ([]*twitter.UserObj, error)
}

type twitterClient struct {
	twitter   Backend
	log       logger.Logger
	userCache cache.Cache[twitter.UserObj]
}

func NewClient(log logger.Logger, cache cache.Cache[twitter.UserObj], backend Backend) *twitterClient {
	return &twitterClient{
		userCache: cache,
		log:       log,
		twitter:   backend,
	}
}

func (c *twitterClient) GetUserTweets(
	ctx context.Context,
	id string,
	opt twitter.UserTweetTimelineOpts,
) (*twitter.UserTweetTimelineResponse, error) {
	timelineResponse, err := c.twitter.UserTweetTimeline(
		ctx,
		id,
		opt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch user tweets")
	}

	return timelineResponse, nil
}

func (c *twitterClient) GetTweets(
	ctx context.Context,
	ids []string,
	opt twitter.TweetLookupOpts,
) (*twitter.TweetLookupResponse, error) {
	response, err := c.twitter.TweetLookup(
		ctx,
		ids,
		opt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch tweets")
	}

	return response, nil
}

func (c *twitterClient) GetUser(ctx context.Context, username string) (*twitter.UserObj, error) {
	cacheKey := strings.ReplaceAll(cacheKeyUsername, "%s", username)
	fromCache, err := c.userCache.Get(ctx, cacheKey)
	if err == nil {
		c.log.WithField("key", cacheKey).Trace("twitter user cache hit")
		return fromCache, nil
	}
	if !errors.Is(err, cache.ErrMiss) {
		return nil, errors.Wrap(err, "error while retrieving twitter user from cache")
	}
	c.log.WithField("key", cacheKey).Trace("twitter user cache miss")

	lookup, err := c.twitter.UserNameLookup(ctx, []string{username}, twitter.UserLookupOpts{
		UserFields: []twitter.UserField{
			twitter.UserFieldID,
			twitter.UserFieldDescription,
			twitter.UserFieldName,
			twitter.UserFieldProfileImageURL,
			twitter.UserFieldCreatedAt,
			twitter.UserFieldPublicMetrics,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch twitter user")
	}

	for _, e := range lookup.Raw.Errors {
		if e.Type == TwitterErrorTypeNotFound {
			return nil, UserNotFoundError{username}
		}
	}
	// Not sure if that can happen that, but that does not cost a lot to secure
	// our code from a slice out of bounds error
	if len(lookup.Raw.Users) == 0 {
		return nil, UserNotFoundError{username}
	}

	if len(lookup.Raw.Errors) != 0 {
		return nil, errors.New("unable to fetch twitter user")
	}

	user := lookup.Raw.Users[0]
	err = c.userCache.Set(ctx, cacheKey, *user)
	if err != nil {
		c.log.WithError(err).Warn("unable to save twitter user to cache")
	}
	// We do not care about if this one is in error, this is only for perf optimization
	_ = c.userCache.Set(ctx, strings.ReplaceAll(cacheKeyID, "%s", user.ID), *user)

	return user, nil
}

func (c *twitterClient) GetUserByIDs(ctx context.Context, ids []string) ([]*twitter.UserObj, error) {
	var results = make([]*twitter.UserObj, 0, len(ids))
	var missingIds = make([]string, 0)
	for _, id := range ids {
		cacheKey := strings.ReplaceAll(cacheKeyID, "%s", id)
		fromCache, err := c.userCache.Get(ctx, cacheKey)
		if err == nil {
			c.log.WithField("key", cacheKey).Trace("twitter user cache hit")
			results = append(results, fromCache)
			continue
		}
		if !errors.Is(err, cache.ErrMiss) {
			return nil, errors.Wrap(err, "error while retrieving twitter user from cache")
		}
		missingIds = append(missingIds, id)
		c.log.WithField("key", cacheKey).Trace("twitter user cache miss")
	}

	if len(missingIds) > 0 {
		resp, err := c.twitter.UserLookup(
			ctx,
			missingIds,
			twitter.UserLookupOpts{
				UserFields: []twitter.UserField{
					twitter.UserFieldID,
					twitter.UserFieldDescription,
					twitter.UserFieldName,
					twitter.UserFieldProfileImageURL,
					twitter.UserFieldCreatedAt,
					twitter.UserFieldPublicMetrics,
				},
			},
		)
		if err != nil {
			return nil, errors.Wrap(err, "unable to fetch users from twitter")
		}
		if len(resp.Raw.Errors) != 0 {
			return nil, errors.New("unable to fetch users from twitter")
		}
		results = append(results, resp.Raw.Users...)
	}

	return results, nil
}
