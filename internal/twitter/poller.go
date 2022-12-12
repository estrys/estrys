package twitter

import (
	"context"
	"time"

	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/repository"
	twittermodels "github.com/estrys/estrys/internal/twitter/models"
	"github.com/estrys/estrys/internal/worker/client"
	"github.com/estrys/estrys/internal/worker/tasks"
)

type TwitterPoller interface {
	Start(context.Context) error
}

type twitterPoller struct {
	log         logger.Logger
	twitter     TwitterClient
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
	client TwitterClient,
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

func (c *twitterPoller) FetchTweets(ctx context.Context) error {
	if len(c.users) == c.userIndex {
		c.log.WithField("index", c.userIndex).Trace("polled all twitter users from list, restarting ...")
		c.userIndex = 0
		err := c.RefreshUserList(ctx)
		if err != nil {
			return err
		}
	}
	user := c.users[c.userIndex]
	userLogger := c.log.WithField("user", user.Username)
	opt := twitter.UserTweetTimelineOpts{
		MaxResults: 100,
		Excludes: []twitter.Exclude{
			twitter.ExcludeReplies,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldID,
			twitter.TweetFieldText,
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPossiblySensitve,
		},
	}
	if cursor, exist := c.userCursors[c.users[c.userIndex].Username]; exist {
		opt.SinceID = cursor
	} else {
		opt.StartTime = c.startTime
	}
	userLogger.WithField("cursor", opt.SinceID).Trace("fetching user tweets")
	tweets, err := c.twitter.GetTweets(ctx, user.ID, opt)
	if err != nil {
		return err
	}
	userLogger.WithField("count", tweets.Meta.ResultCount).Trace("fetched tweets")
	if tweets.Meta.ResultCount > 0 {
		for _, tweet := range tweets.Raw.Tweets {
			createdAt, err := time.Parse(time.RFC3339, tweet.CreatedAt)
			if err != nil {
				c.log.WithError(err).Error("unable to decode tweet date")
				continue
			}
			actors, err := c.repo.GetFollowers(ctx, user)
			if err != nil {
				c.log.WithField("user", user.Username).WithError(err).Error("unable to retrieve followers for user")
				continue
			}
			for _, actor := range actors {
				sendTweetTask, err := tasks.NewSendTweet(user, actor, twittermodels.Tweet{
					ID:        tweet.ID,
					Text:      tweet.Text,
					Published: createdAt,
					Sensitive: tweet.PossiblySensitive,
				})
				if err != nil {
					c.log.WithError(err).Error("unable to create send tweet task")
					continue
				}
				_, err = c.worker.Enqueue(sendTweetTask)
				if err != nil {
					c.log.WithField("user", user.Username).WithError(err).Error("unable to schedule send tweet task")
				}
				c.log.WithField("tweet", tweet.ID).Info("scheduled new tweet send")
			}
		}
		c.userCursors[c.users[c.userIndex].Username] = tweets.Meta.NewestID
	}
	c.userIndex++
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
			}
		case <-ctx.Done():
			c.log.Info("Stopping poller")
			return nil
		}
	}
}
