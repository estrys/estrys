package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/estrys/estrys/internal/activitypub"
	activitypubclient "github.com/estrys/estrys/internal/activitypub/client"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/domain"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/repository"
	taskerrors "github.com/estrys/estrys/internal/worker/errors"
	"github.com/estrys/estrys/internal/worker/tasks"
)

func HandleSendTweet(ctx context.Context, task *asynq.Task) error {
	log := dic.GetService[logger.Logger]()
	vocabService := dic.GetService[activitypub.VocabService]()
	userRepo := dic.GetService[repository.UserRepository]()
	actorRepo := dic.GetService[repository.ActorRepository]()
	tweetService := dic.GetService[domain.TweetService]()
	activityPubClient := dic.GetService[activitypubclient.ActivityPubClient]()

	var input tasks.SendTweetInput
	if err := json.Unmarshal(task.Payload(), &input); err != nil {
		log.WithError(err).Error("unable to deserialize task input")
		return taskerrors.TaskError{
			SkipRetry: true,
			Err:       errors.Wrap(err, "unable to deserialize task input"),
		}
	}

	tweet, err := tweetService.SaveTweetAndReferences(ctx, input.TweetID)
	if err != nil {
		var errResponse *twitter.ErrorResponse
		if errors.As(err, &errResponse) {
			if errResponse.StatusCode == http.StatusTooManyRequests {
				return taskerrors.TaskError{
					Err: errors.Wrap(err, "got rate limited while fetching tweets amd references"),
				}
			}
		}
		return taskerrors.TaskError{
			SkipRetry: true,
			Err:       errors.Wrap(err, "unable to save tweets and references"),
		}
	}

	user, err := userRepo.Get(ctx, input.From)
	if err != nil {
		return taskerrors.TaskError{
			SkipRetry: true,
			Err:       errors.Wrap(err, "unable to fetch user from database"),
		}
	}

	actorURL, _ := url.Parse(input.To)
	actor, err := actorRepo.Get(ctx, actorURL)
	if err != nil {
		return taskerrors.TaskError{
			SkipRetry: true,
			Err:       errors.Wrap(err, "unable to fetch actor from database"),
		}
	}

	createNote, err := vocabService.GetCreateNoteFromTweet(user.Username, *tweet)
	if err != nil {
		return taskerrors.TaskError{
			SkipRetry: true,
			Err:       errors.Wrap(err, "unable to create an create tweet activity"),
		}
	}
	err = activityPubClient.PostInbox(ctx, actor, user, createNote)
	if err != nil {
		var isNotAcceptedErr *activitypubclient.InboxNotAcceptedError
		if errors.As(err, &isNotAcceptedErr) {
			return taskerrors.TaskError{
				SkipRetry: true,
				Err:       errors.Wrap(err, "post to inbox was not accepted"),
			}
		}
		return taskerrors.TaskError{
			SkipRetry: true,
			Err:       errors.Wrap(err, "unable to send create tweet"),
		}
	}

	log.WithFields(logrus.Fields{
		"from":  input.From,
		"to":    input.To,
		"tweet": tweet.ID,
	}).Info("tweet sent")
	return nil
}
