package domain

import (
	"context"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/observability"
	"github.com/estrys/estrys/internal/twitter"
	twittermodels "github.com/estrys/estrys/internal/twitter/models"
	"github.com/estrys/estrys/internal/twitter/repository"
)

//go:generate mockery --with-expecter --name=TweetService
type TweetService interface {
	SaveTweetAndReferences(context.Context, string) (*twittermodels.Tweet, error)
}

type tweetService struct {
	logger        logger.Logger
	userService   UserService
	tweeterClient twitter.TwitterClient
	tweetRepo     repository.TweetRepository
}

func NewTweetService(
	logger logger.Logger,
	userService UserService,
	tweeterClient twitter.TwitterClient,
	tweetRepo repository.TweetRepository,
) *tweetService {
	return &tweetService{
		logger:        logger,
		userService:   userService,
		tweeterClient: tweeterClient,
		tweetRepo:     tweetRepo,
	}
}

func (t *tweetService) convertTweet(
	tweet *gotwitter.TweetObj,
	referenceType twittermodels.ReferenceType,
) (*twittermodels.Tweet, error) {
	createdAt, err := time.Parse(time.RFC3339, tweet.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode tweet date")
	}
	return &twittermodels.Tweet{
		ID:             tweet.ID,
		ReferencedType: referenceType,
		AuthorID:       tweet.AuthorID,
		Text:           tweet.Text,
		Published:      createdAt,
		Sensitive:      tweet.PossiblySensitive,
	}, nil
}

func (t *tweetService) fetchReferencedTweets(
	ctx context.Context,
	rawTweet *gotwitter.TweetObj,
) ([]*twittermodels.Tweet, error) {
	var result = make([]*twittermodels.Tweet, 0, len(rawTweet.ReferencedTweets))
	var missingReferencedTweetsIDs = make([]string, 0, len(rawTweet.ReferencedTweets))
	for _, referencedTweet := range rawTweet.ReferencedTweets {
		// Check if a referenced tweet is already known and avoid to fetch it from twitter
		if tweet, err := t.tweetRepo.GetTweet(ctx, referencedTweet.ID); err == nil && tweet != nil {
			result = append(result, tweet)
			continue
		}
		missingReferencedTweetsIDs = append(missingReferencedTweetsIDs, referencedTweet.ID)
	}

	if len(missingReferencedTweetsIDs) == 0 {
		return result, nil
	}

	resp, err := t.tweeterClient.GetTweets(ctx, missingReferencedTweetsIDs, gotwitter.TweetLookupOpts{
		TweetFields: []gotwitter.TweetField{
			gotwitter.TweetFieldID,
			gotwitter.TweetFieldAuthorID,
			gotwitter.TweetFieldText,
			gotwitter.TweetFieldCreatedAt,
			gotwitter.TweetFieldPossiblySensitve,
			gotwitter.TweetFieldReferencedTweets,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Raw.Errors) != 0 {
		return nil, errors.New("unable to fetch tweets")
	}

	// We store user ids from referenced tweets to see if we also need to fetch their
	// author infos
	for _, rawReferencedTweet := range resp.Raw.Tweets {
		var referenceType twittermodels.ReferenceType
		for _, refTweet := range rawTweet.ReferencedTweets {
			if refTweet.ID == rawReferencedTweet.ID {
				referenceType = twittermodels.ReferenceType(refTweet.Type)
			}
		}
		referencedTweet, err := t.convertTweet(rawReferencedTweet, referenceType)
		if err != nil {
			return nil, err
		}
		result = append(result, referencedTweet)
		t.logger.WithField("id", referencedTweet.ID).Debug("saved referenced tweet")
	}

	return result, nil
}

func (t *tweetService) saveReferencedTweets(
	ctx context.Context,
	referencedTweets []*twittermodels.Tweet,
) error {
	authorIDs := make([]string, 0, len(referencedTweets))
	for _, referencedTweet := range referencedTweets {
		authorIDs = append(authorIDs, referencedTweet.AuthorID)
	}
	authors, err := t.userService.BatchCreateUsersFromIDs(ctx, authorIDs)
	if err != nil {
		return errors.Wrap(err, "unable to create users for referenced tweets")
	}
	for _, referencedTweet := range referencedTweets {
		for _, user := range authors {
			if user.ID == referencedTweet.AuthorID {
				referencedTweet.AuthorUsername = user.Username
				break
			}
		}
		err = t.tweetRepo.Store(ctx, referencedTweet)
		if err != nil {
			return errors.Wrap(err, "unable to save referenced tweet")
		}
	}
	return nil
}

func (t tweetService) fetchRawTweet(ctx context.Context, tweetID string) (*gotwitter.TweetObj, error) {
	tweetResponse, err := t.tweeterClient.GetTweets(ctx, []string{tweetID}, gotwitter.TweetLookupOpts{
		TweetFields: []gotwitter.TweetField{
			gotwitter.TweetFieldID,
			gotwitter.TweetFieldAuthorID,
			gotwitter.TweetFieldText,
			gotwitter.TweetFieldCreatedAt,
			gotwitter.TweetFieldPossiblySensitve,
			gotwitter.TweetFieldReferencedTweets,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "error while fetching tweet details from twitter")
	}
	if len(tweetResponse.Raw.Tweets) != 1 {
		return nil, errors.New("no tweets returned, expected one")
	}
	return tweetResponse.Raw.Tweets[0], nil
}

// SaveTweetAndReferences save the tweet in database.
// Also resolve all referenced tweets and their authors and save them in the database too.
func (t *tweetService) SaveTweetAndReferences(
	ctx context.Context,
	tweetID string,
) (*twittermodels.Tweet, error) {
	span := observability.StartSpan(ctx, "tweet.save", map[string]any{"tweet.id": tweetID})
	if span != nil {
		span.Status = sentry.SpanStatusInternalError
		ctx = span.Context()
		defer span.Finish()
	}

	// Check if a tweet is already known and return it
	if tweet, err := t.tweetRepo.GetTweet(ctx, tweetID); err == nil && tweet != nil {
		if span != nil {
			span.Status = sentry.SpanStatusAlreadyExists
		}
		return tweet, nil
	}

	rawTweet, err := t.fetchRawTweet(ctx, tweetID)
	if err != nil {
		return nil, err
	}

	tweet, err := t.convertTweet(rawTweet, "")
	if err != nil {
		return nil, err
	}

	if len(rawTweet.ReferencedTweets) > 0 {
		referencedTweets, err := t.fetchReferencedTweets(ctx, rawTweet)
		if err != nil {
			return nil, errors.Wrap(err, "unable to fetch referenced tweets")
		}
		err = t.saveReferencedTweets(ctx, referencedTweets)
		if err != nil {
			return nil, err
		}
		for _, referencedTweet := range referencedTweets {
			tweet.ReferencedTweets = append(tweet.ReferencedTweets, *referencedTweet)
		}
	}

	author, err := t.userService.BatchCreateUsersFromIDs(ctx, []string{tweet.AuthorID})
	if err != nil || len(author) != 1 {
		return nil, errors.Wrap(err, "unable to fetch author for tweet")
	}

	tweet.AuthorUsername = author[0].Username
	err = t.tweetRepo.Store(ctx, tweet)
	if err != nil {
		return nil, errors.Wrap(err, "unable to save tweet")
	}

	if span != nil {
		span.Status = sentry.SpanStatusOK
	}
	t.logger.WithField("id", tweet.ID).Debug("saved tweet")
	return tweet, nil
}
