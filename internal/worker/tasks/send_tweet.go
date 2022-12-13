package tasks

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/estrys/estrys/internal/activitypub"
	activitypubclient "github.com/estrys/estrys/internal/activitypub/client"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/repository"
	twittermodels "github.com/estrys/estrys/internal/twitter/models"
	"github.com/estrys/estrys/internal/worker/queues"
)

type SendTweetInput struct {
	From  string              `json:"from"`
	To    string              `json:"to"`
	Tweet twittermodels.Tweet `json:"tweet"`
}

func NewSendTweet(user *models.User, actor *models.Actor, tweet twittermodels.Tweet) (*asynq.Task, error) {
	payload, err := json.Marshal(SendTweetInput{
		From:  user.Username,
		To:    actor.URL,
		Tweet: tweet,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return asynq.NewTask(
		TypeSendTweet,
		payload,
		asynq.MaxRetry(5),
		asynq.Timeout(10*time.Second),
		asynq.Queue(queues.QueueTweets),
		asynq.Retention(1*time.Hour),
	), nil
}

func HandleSendTweet(ctx context.Context, task *asynq.Task) error {
	log := dic.GetService[logger.Logger]()
	vocabService := dic.GetService[activitypub.VocabService]()
	userRepo := dic.GetService[repository.UserRepository]()
	actorRepo := dic.GetService[repository.ActorRepository]()
	activityPubClient := dic.GetService[activitypubclient.ActivityPubClient]()

	var input SendTweetInput
	if err := json.Unmarshal(task.Payload(), &input); err != nil {
		log.WithError(err).Error("unable to deserialize task input")
		return errors.Errorf("unable to deserialize task input : %v: %s", err, asynq.SkipRetry)
	}

	user, err := userRepo.Get(ctx, input.From)
	if err != nil {
		return errors.Wrap(err, "unable to fetch user from database")
	}

	actorURL, _ := url.Parse(input.To)
	actor, err := actorRepo.Get(ctx, actorURL)
	if err != nil {
		return errors.Wrap(err, "unable to fetch actor from database")
	}

	createNote, err := vocabService.GetCreateNoteFromTweet(user.Username, input.Tweet)
	if err != nil {
		return errors.Wrapf(err, "unable to create an create tweet activity: %s", asynq.SkipRetry)
	}
	err = activityPubClient.PostInbox(ctx, actor, user, createNote)
	if err != nil {
		var isNotAcceptedErr *activitypubclient.InboxNotAcceptedError
		if errors.As(err, isNotAcceptedErr) {
			return errors.Wrapf(err, "post to inbox was not accepted: %s", asynq.SkipRetry)
		}
		return errors.Wrapf(err, "unable to send create tweet: %s", asynq.SkipRetry)
	}

	log.WithFields(logrus.Fields{
		"from":  input.From,
		"to":    input.To,
		"tweet": input.Tweet.ID,
	}).Info("tweet sent")
	return nil
}
