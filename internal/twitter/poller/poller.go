package poller

import (
	"context"
	"runtime/debug"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/observability"
	"github.com/estrys/estrys/internal/repository"
	"github.com/estrys/estrys/internal/twitter"
	"github.com/estrys/estrys/internal/worker/client"
	"github.com/estrys/estrys/internal/worker/tasks"
)

type TwitterPoller interface {
	Start(context.Context) error
}

type twitterPoller struct {
	log         logger.Logger
	twitter     twitter.TwitterClient
	repo        repository.UserRepository
	worker      client.BackgroundWorkerClient
	users       models.UserSlice
	userCursors map[string]string
	userIndex   int
	startTime   time.Time
}

var (
	ErrNoUserToPoll = errors.New("no user to poll")
)

const (
	maxRequests = 1500
	periodMins  = 15
)

func NewPoller(
	log logger.Logger,
	client twitter.TwitterClient,
	repo repository.UserRepository,
	worker client.BackgroundWorkerClient,
) *twitterPoller {
	return &twitterPoller{
		log:         log,
		twitter:     client,
		repo:        repo,
		worker:      worker,
		userCursors: map[string]string{},
	}
}

func (c *twitterPoller) RefreshUserList(ctx context.Context) error {
	var err error
	c.users, err = c.repo.GetWithFollowers(ctx)
	if err != nil {
		return err
	}
	if c.users == nil {
		return ErrNoUserToPoll
	}
	return nil
}

func (c *twitterPoller) FetchTweets(ctx context.Context) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Errorf("got a panic during poller: %s: %s", rec, string(debug.Stack()))
		}
	}()
	tx := observability.StartTransaction(ctx, "tweets.poll", func(s *sentry.Span) {
		s.Sampled = sentry.SampledFalse
	})
	ctx = tx.Context()
	if len(c.users) == c.userIndex {
		c.log.WithField("index", c.userIndex).Trace("polled all twitter users from list, restarting ...")
		c.userIndex = 0
		err := c.RefreshUserList(ctx)
		if err != nil {
			if errors.Is(err, ErrNoUserToPoll) {
				return nil
			}
			return err
		}
	}
	user := c.users[c.userIndex]
	userLogger := c.log.WithField("user", user.Username)
	opt := gotwitter.UserTweetTimelineOpts{
		MaxResults: 100,
		Excludes: []gotwitter.Exclude{
			gotwitter.ExcludeReplies,
		},
		TweetFields: []gotwitter.TweetField{
			gotwitter.TweetFieldID,
		},
	}
	if cursor, exist := c.userCursors[c.users[c.userIndex].Username]; exist {
		opt.SinceID = cursor
	} else {
		opt.StartTime = c.startTime
	}
	userLogger.WithField("cursor", opt.SinceID).Trace("fetching user tweets")
	tweets, err := c.twitter.GetUserTweets(ctx, user.ID, opt)
	if err != nil {
		return err
	}
	userLogger.WithField("count", tweets.Meta.ResultCount).Trace("fetched tweets")
	if tweets.Meta.ResultCount > 0 {
		tx.Sampled = sentry.SampledTrue
		for _, tweet := range tweets.Raw.Tweets {
			err := c.handleTweet(ctx, user, tweet)
			if err != nil {
				return err
			}
			c.log.WithField("tweet", tweet.ID).Info("scheduled new tweet send")
		}
		c.userCursors[c.users[c.userIndex].Username] = tweets.Meta.NewestID
	}
	tx.Data = map[string]interface{}{
		"new_tweets_count": tweets.Meta.ResultCount,
		"users_count":      len(c.users),
	}
	tx.Finish()
	c.userIndex++
	return nil
}

func (c *twitterPoller) handleTweet(
	ctx context.Context,
	user *models.User,
	rawTweet *gotwitter.TweetObj,
) error {
	actors, err := c.repo.GetFollowers(ctx, user)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve followers for user")
	}
	for _, actor := range actors {
		sendTweetTask, err := tasks.NewSendTweet(ctx, user, actor, rawTweet.ID)
		if err != nil {
			return errors.Wrap(err, "unable to create send tweet task")
		}
		_, err = c.worker.Enqueue(sendTweetTask)
		if err != nil {
			return errors.Wrap(err, "unable to schedule send tweet task")
		}
	}
	return nil
}

func (c *twitterPoller) Start(ctx context.Context) error {
	c.log.Info("Starting poller")
	err := c.RefreshUserList(ctx)
	if err != nil {
		if !errors.Is(err, ErrNoUserToPoll) {
			return errors.Wrap(err, "unexpected error while fetching users")
		}
		if errors.Is(err, ErrNoUserToPoll) {
			c.log.Debug("no user to poll, polling for new users")
			for errors.Is(err, ErrNoUserToPoll) {
				c.log.Debug("polling for users to poll")
				err = c.RefreshUserList(ctx)
				if err != nil && !errors.Is(err, ErrNoUserToPoll) {
					return err
				}
				time.Sleep(5 * time.Second)
			}
		}
	}

	c.log.Debug("we have users to poll")

	ticker := time.NewTicker(periodMins * time.Minute / maxRequests)
	c.startTime = time.Now()

	for {
		select {
		case <-ticker.C:
			err := c.FetchTweets(ctx)
			if err != nil {
				c.log.WithError(err).Error("an unexpected error happened during tweets fetching")
				sentry.CaptureException(err)
			}
		case <-ctx.Done():
			c.log.Info("Stopping poller")
			return nil
		}
	}
}
