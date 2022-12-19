package tasks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"

	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/observability"
	"github.com/estrys/estrys/internal/worker/queues"
)

type SendTweetInput struct {
	TraceID string `json:"trace_id"`
	From    string `json:"from"`
	To      string `json:"to"`
	TweetID string `json:"tweet_id"`
}

func NewSendTweet(
	ctx context.Context,
	user *models.User,
	actor *models.Actor,
	tweetID string,
) (*asynq.Task, error) {
	payload, err := json.Marshal(SendTweetInput{
		TraceID: observability.GetTraceIDFromContext(ctx),
		From:    user.Username,
		To:      actor.URL,
		TweetID: tweetID,
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
