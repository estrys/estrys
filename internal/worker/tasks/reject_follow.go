package tasks

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/estrys/estrys/internal/activitypub"
	activitypubclient "github.com/estrys/estrys/internal/activitypub/client"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/observability"
	"github.com/estrys/estrys/internal/repository"
	"github.com/estrys/estrys/internal/worker/queues"
)

type RejectFollowInput struct {
	TraceID  string `json:"trace_id,omitempty"`
	Username string `json:"user"`
	Activity map[string]interface{}
}

func NewRejectFollowTask(ctx context.Context, username string, act vocab.ActivityStreamsFollow) (*asynq.Task, error) {
	serializedActivity, err := streams.Serialize(act)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	payload, err := json.Marshal(RejectFollowInput{
		TraceID:  observability.GetTraceIDFromContext(ctx),
		Username: username,
		Activity: serializedActivity,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return asynq.NewTask(
		TypeRejectFollow,
		payload,
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
		asynq.Queue(queues.QueueFollows),
		asynq.Retention(24*time.Hour),
	), nil
}

func HandleRejectFollow(ctx context.Context, task *asynq.Task) error {
	log := dic.GetService[logger.Logger]()
	vocabService := dic.GetService[activitypub.VocabService]()
	userRepo := dic.GetService[repository.UserRepository]()
	actorRepo := dic.GetService[repository.ActorRepository]()
	activityPubClient := dic.GetService[activitypubclient.ActivityPubClient]()

	var input RejectFollowInput
	if err := json.Unmarshal(task.Payload(), &input); err != nil {
		log.WithError(err).Error("unable to deserialize task input")
		return errors.Errorf("unable to deserialize task input : %v: %s", err, asynq.SkipRetry)
	}

	var follow vocab.ActivityStreamsFollow
	resolver, err := streams.NewJSONResolver(func(_ context.Context, act vocab.ActivityStreamsFollow) error {
		follow = act
		return nil
	})
	if err != nil {
		return errors.Errorf("unable to create json resolver : %v: %s", err, asynq.SkipRetry)
	}
	err = resolver.Resolve(ctx, input.Activity)
	if err != nil {
		return errors.Errorf("unable to find a follow activity : %v: %s", err, asynq.SkipRetry)
	}

	user, err := userRepo.Get(ctx, input.Username)
	if err != nil {
		return errors.Wrap(err, "unable to fetch user from database")
	}

	actorURL, err := activitypub.GetActorURL(follow)
	if err != nil {
		return errors.Errorf("unable parse actor URL : %v: %s", err, asynq.SkipRetry)
	}
	actor, err := actorRepo.Get(ctx, actorURL)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve actor")
	}

	rejectFollow, err := vocabService.GetReject(user, follow)
	if err != nil {
		return errors.Errorf("unable to create an reject request : %v: %s", err, asynq.SkipRetry)
	}
	err = activityPubClient.PostInbox(ctx, actor, user, rejectFollow)
	if err != nil {
		var isNotAcceptedErr *activitypubclient.InboxNotAcceptedError
		if errors.As(err, &isNotAcceptedErr) {
			// If the error is not on our side, let's, try to retry
			if isNotAcceptedErr.StatusCode >= http.StatusInternalServerError {
				err = errors.Wrap(err, asynq.SkipRetry.Error())
			}
			return errors.Wrapf(err, "post to inbox was not accepted")
		}
		return errors.Wrapf(err, "unable to reject follow request: %s", asynq.SkipRetry)
	}

	log.WithFields(logrus.Fields{
		"user":  input.Username,
		"inbox": actorURL.String(),
	}).Info("sent reject follow to inbox")
	return nil
}
