package twitter

import (
	"context"
	"strconv"
	"time"

	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/repository"
)

type TwitterPoller interface {
	Start(context.Context) error
}

type twitterPoller struct {
	log         logger.Logger
	twitter     TwitterClient
	repo        repository.UserRepository
	users       models.UserSlice
	userCursors map[string]string
	userIndex   int
}

var (
	ErrNoUserToPoll = errors.New("no user to poll")
)

const (
	MaxRequests = 1500
	PeriodMins  = 15
)

func NewPoller(
	log logger.Logger,
	client TwitterClient,
	repo repository.UserRepository,
) *twitterPoller {
	return &twitterPoller{
		log:         log,
		twitter:     client,
		repo:        repo,
		userCursors: map[string]string{},
	}
}

func (c *twitterPoller) RefreshUserList(ctx context.Context) error {
	var err error
	c.users, err = c.repo.GetWithoutActor(ctx)
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
	}
	user := c.users[c.userIndex]
	userLogger := c.log.WithField("user", user.Username)
	idu64, _ := user.ID.Uint64()
	id := strconv.FormatUint(idu64, 10)
	opt := twitter.UserTweetTimelineOpts{}
	if cursor, exist := c.userCursors[c.users[c.userIndex].Username]; exist {
		opt.SinceID = cursor
	}
	userLogger.WithField("cursor", opt.SinceID).Trace("fetching user tweets")
	tweets, err := c.twitter.GetTweets(ctx, id, opt)
	if err != nil {
		return err
	}
	userLogger.WithField("count", tweets.Meta.ResultCount).Trace("fetched tweets")
	if tweets.Meta.ResultCount > 0 {
		// TODO Do something with the tweets ;)
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

	ticker := time.NewTicker(PeriodMins * time.Minute / MaxRequests)
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
