package twitter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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
)

type UsernameNotFoundError struct {
	Username string
}

func (u UsernameNotFoundError) Error() string {
	return fmt.Sprintf("twitter user %s not found", u.Username)
}

type User struct {
	Username        string
	Name            string
	Description     string
	ProfileImageURL *url.URL
	CreatedAt       time.Time
	Followers       uint64
	Following       uint64
	Tweets          uint64
}

//go:generate mockery --name=Backend
type Backend interface {
	UserNameLookup(
		ctx context.Context,
		usernames []string,
		opts twitter.UserLookupOpts,
	) (*twitter.UserLookupResponse, error)
}

//go:generate mockery --name=TwitterClient
type TwitterClient interface {
	// GetTweets(ids []string) ([]*twitter.TweetObj, error)
	// StreamTweets(callback func(*twitter.TweetMessage)) error

	GetUser(ctx context.Context, username string) (*User, error)
}

type twitterClient struct {
	twitter Backend
	log     logger.Logger
	cache   cache.Cache[User]
}

func NewClient(log logger.Logger, cache cache.Cache[User], backend Backend) *twitterClient {
	return &twitterClient{
		cache:   cache,
		log:     log,
		twitter: backend,
	}
}

//nolint:gocritic
//func (c *twitterClient) GetTweets(ids []string) ([]*twitter.TweetObj, error) {
//	tweets, err := c.twitter.TweetLookup(
//		context.TODO(),
//		ids,
//		twitter.TweetLookupOpts{
//			TweetFields: []twitter.TweetField{
//				twitter.TweetFieldReferencedTweets,
//			},
//		},
//	)
//
//	if err != nil {
//		return nil, errors.Wrap(err, "unable to fetch tweets")
//	}
//
//	return tweets.Raw.Tweets, nil
//}

func (c *twitterClient) GetUser(ctx context.Context, username string) (*User, error) {
	cacheKey := strings.Join([]string{"twitter", "user", username}, "/")
	fromCache, err := c.cache.Get(ctx, cacheKey)
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
			return nil, UsernameNotFoundError{username}
		}
	}

	if len(lookup.Raw.Errors) != 0 {
		return nil, errors.New("unable to fetch twitter user")
	}

	createdAt, err := time.Parse(time.RFC3339, lookup.Raw.Users[0].CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse user creation date from twitter")
	}

	profileImage, err := url.Parse(strings.ReplaceAll(
		lookup.Raw.Users[0].ProfileImageURL,
		"_normal",
		"",
	))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse profile image url")
	}

	user := &User{
		Username:        username,
		Name:            lookup.Raw.Users[0].Name,
		Description:     lookup.Raw.Users[0].Description,
		ProfileImageURL: profileImage,
		CreatedAt:       createdAt,
		Following:       uint64(lookup.Raw.Users[0].PublicMetrics.Following),
		Followers:       uint64(lookup.Raw.Users[0].PublicMetrics.Followers),
		Tweets:          uint64(lookup.Raw.Users[0].PublicMetrics.Tweets),
	}

	err = c.cache.Set(ctx, cacheKey, *user)
	if err != nil {
		c.log.WithError(err).Warn("unable to save twitter user to cache")
	}

	return user, nil
}

//nolint:gocritic
//func (c *twitterClient) StreamTweets(callback func(*twitter.TweetMessage)) error {
//	stream, err := c.twitter.TweetSearchStream(context.TODO(), twitter.TweetSearchStreamOpts{
//		TweetFields: []twitter.TweetField{
//			twitter.TweetFieldReferencedTweets,
//		},
//	})
//	if err != nil {
//		return errors.Wrap(err, "unable to fetch tweet steam")
//	}
//
//	go func() {
//		for {
//			tweet := <-stream.Tweets()
//			if tweet != nil {
//				callback(tweet)
//			}
//		}
//	}()
//
//	return nil
//}
