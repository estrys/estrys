package domain

import (
	"context"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/observability"
	"github.com/estrys/estrys/internal/twitter"
	"github.com/estrys/estrys/internal/twitter/models"
	"github.com/estrys/estrys/internal/twitter/repository"
)

//go:generate mockery --with-expecter --name=TweetService
type TweetService interface {
	SaveTweetAndReferences(context.Context, *gotwitter.TweetObj) (*models.Tweet, error)
}

type tweetService struct {
	userService   UserService
	tweeterClient twitter.TwitterClient
	tweetRepo     repository.TweetRepository
}

func NewTweetService(
	userService UserService,
	tweeterClient twitter.TwitterClient,
	tweetRepo repository.TweetRepository,
) *tweetService {
	return &tweetService{
		userService:   userService,
		tweeterClient: tweeterClient,
		tweetRepo:     tweetRepo,
	}
}

func (t *tweetService) convertTweet(tweet *gotwitter.TweetObj) (*models.Tweet, error) {
	createdAt, err := time.Parse(time.RFC3339, tweet.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode tweet date")
	}
	return &models.Tweet{
		ID:        tweet.ID,
		Text:      tweet.Text,
		Published: createdAt,
		Sensitive: tweet.PossiblySensitive,
	}, nil
}

func (t *tweetService) saveReferencedTweets(
	ctx context.Context,
	rawTweet *gotwitter.TweetObj,
) ([]models.Tweet, error) {
	var result = make([]models.Tweet, 0, len(rawTweet.ReferencedTweets))
	var missingReferencedTweetsIDs = make([]string, 0, len(rawTweet.ReferencedTweets))
	for _, referencedTweet := range rawTweet.ReferencedTweets {
		// Check if a referenced tweet is already known and avoid to fetch it from twitter
		if tweet, err := t.tweetRepo.GetTweet(ctx, referencedTweet.ID); err == nil && tweet != nil {
			result = append(result, *tweet)
			continue
		}
		missingReferencedTweetsIDs = append(missingReferencedTweetsIDs, referencedTweet.ID)
	}

	if len(missingReferencedTweetsIDs) == 0 {
		return result, nil
	}

	resp, err := t.tweeterClient.GetTweets(ctx, missingReferencedTweetsIDs, gotwitter.TweetLookupOpts{
		TweetFields: []gotwitter.TweetField{
			gotwitter.TweetFieldText,
			gotwitter.TweetFieldCreatedAt,
			gotwitter.TweetFieldPossiblySensitve,
			gotwitter.TweetFieldAuthorID,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Raw.Errors) != 0 {
		return nil, errors.New("unable to fetch twitter user")
	}

	// We store user ids from referenced tweets to see if we also need to fetch their
	// author infos
	authorIDs := make([]string, 0, len(resp.Raw.Tweets))
	for _, rawReferencedTweet := range resp.Raw.Tweets {
		referencedTweet, err := t.convertTweet(rawReferencedTweet)
		if err != nil {
			return nil, err
		}
		err = t.tweetRepo.Store(ctx, referencedTweet)
		if err != nil {
			return nil, err
		}
		authorIDs = append(authorIDs, rawReferencedTweet.AuthorID)
		result = append(result, *referencedTweet)
	}

	err = t.userService.BatchCreateUsersFromIDs(ctx, authorIDs)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SaveTweetAndReferences save the tweet in database.
// Also resolve all referenced tweets and their authors and save them in the database too.
func (t *tweetService) SaveTweetAndReferences(
	ctx context.Context,
	rawTweet *gotwitter.TweetObj,
) (*models.Tweet, error) {
	span := observability.StartSpan(ctx, "tweet.save", map[string]any{"tweet.id": rawTweet.ID})
	if span != nil {
		span.Status = sentry.SpanStatusInternalError
		ctx = span.Context()
		defer span.Finish()
	}

	// Check if a tweet is already known and return it
	if tweet, err := t.tweetRepo.GetTweet(ctx, rawTweet.ID); err == nil && tweet != nil {
		if span != nil {
			span.Status = sentry.SpanStatusAlreadyExists
		}
		return tweet, nil
	}

	tweet, err := t.convertTweet(rawTweet)
	if err != nil {
		return nil, err
	}

	if len(rawTweet.ReferencedTweets) > 0 {
		referencedTweets, err := t.saveReferencedTweets(ctx, rawTweet)
		if err != nil {
			return nil, errors.Wrap(err, "unable to fetch referenced tweets")
		}
		tweet.ReferencedTweets = referencedTweets
	}

	err = t.tweetRepo.Store(ctx, tweet)
	if err != nil {
		return nil, errors.Wrap(err, "unable to save tweet with references")
	}

	if span != nil {
		span.Status = sentry.SpanStatusOK
	}
	return tweet, nil
}
