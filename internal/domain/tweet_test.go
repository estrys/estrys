package domain

import (
	"context"
	"testing"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/estrys/estrys/internal/domain/mocks"
	mockstwitter "github.com/estrys/estrys/internal/twitter/mocks"
	"github.com/estrys/estrys/internal/twitter/models"
	mockstwitterrepo "github.com/estrys/estrys/internal/twitter/repository/mocks"
)

func TestTweetService_SaveTweetAndReferences(t *testing.T) {
	fakeDateStr := "2006-01-02T15:04:05Z"
	fakeDate, _ := time.Parse(time.RFC3339, fakeDateStr)

	fakeCompleteTweetInput := &gotwitter.TweetObj{
		ID:                "1234",
		Text:              "RT @someone: this text is gonna be truncated",
		CreatedAt:         fakeDateStr,
		PossiblySensitive: true,
		ReferencedTweets: []*gotwitter.TweetReferencedTweetObj{
			{
				Type: "retweeted",
				ID:   "4321",
			},
			{
				Type: "retweeted",
				ID:   "7654",
			},
		},
	}
	fakeReferencedTweet4321 := &gotwitter.TweetObj{
		ID:        "4321",
		CreatedAt: fakeDateStr,
		AuthorID:  "author1",
	}
	fakeReferencedTweet7654 := &gotwitter.TweetObj{
		ID:        "7654",
		CreatedAt: fakeDateStr,
		AuthorID:  "author2",
	}

	fakeReferencedTweetModel4321 := &models.Tweet{
		ID:        "4321",
		Published: fakeDate,
	}
	fakeReferencedTweetModel7654 := &models.Tweet{
		ID:        "7654",
		Published: fakeDate,
	}

	expectedTweetLookupOpts := gotwitter.TweetLookupOpts{
		TweetFields: []gotwitter.TweetField{
			gotwitter.TweetFieldText,
			gotwitter.TweetFieldCreatedAt,
			gotwitter.TweetFieldPossiblySensitve,
			gotwitter.TweetFieldAuthorID,
		},
	}

	cases := []struct {
		name       string
		tweetInput *gotwitter.TweetObj
		output     *models.Tweet
		mocks      func(*mocks.UserService, *mockstwitter.TwitterClient, *mockstwitterrepo.TweetRepository)
		err        string
	}{
		{
			name: "tweet already known",
			mocks: func(_ *mocks.UserService, _ *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(&models.Tweet{
					ID: "1234",
				}, nil)
			},
			output: &models.Tweet{
				ID: "1234",
			},
			tweetInput: &gotwitter.TweetObj{
				ID: "1234",
			},
		},
		{
			name: "invalid tweet date",
			mocks: func(_ *mocks.UserService, _ *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Return(nil, nil)
			},
			err: `unable to decode tweet date: parsing time "invalid_date" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid_date" as "2006"`,
			tweetInput: &gotwitter.TweetObj{
				CreatedAt: "invalid_date",
			},
		},
		{
			name:       "all tweets already exist",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(_ *mocks.UserService, _ *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, "4321").Times(1).Return(&models.Tweet{
					ID: "4321",
				}, nil)
				repository.EXPECT().GetTweet(mock.Anything, "7654").Times(1).Return(&models.Tweet{
					ID: "7654",
				}, nil)
				// Not super usefull to test the content of the stored tweet here, we are gonna
				// assert it with the output of this case
				repository.EXPECT().Store(mock.Anything, mock.Anything).Return(nil)
			},
			output: &models.Tweet{
				ID:        fakeCompleteTweetInput.ID,
				Text:      fakeCompleteTweetInput.Text,
				Published: fakeDate,
				Sensitive: true,
				ReferencedTweets: []models.Tweet{
					{
						ID: "4321",
					},
					{
						ID: "7654",
					},
				},
			},
		},
		{
			name:       "unable to store tweet",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(_ *mocks.UserService, _ *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(&models.Tweet{
					ID: "fake",
				}, nil)
				repository.EXPECT().Store(mock.Anything, mock.Anything).Return(errors.New("repo error"))
			},
			err: "unable to save tweet with references: repo error",
		},
		{
			name:       "tweeter client error while fetching referenced tweets",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(nil, errors.New("twitter client error"))
			},
			err: "unable to fetch referenced tweets: twitter client error",
		},
		{
			name:       "tweeter client error while fetching referenced tweets",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Errors: []*gotwitter.ErrorObj{
								{Type: "error in response"},
							},
						},
					}, nil)
			},
			err: "unable to fetch referenced tweets: unable to fetch twitter user",
		},
		{
			name:       "invalid date in returned referenced tweet",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Tweets: []*gotwitter.TweetObj{
								{
									CreatedAt: "invalid_date",
								},
							},
						},
					}, nil)
			},
			err: `unable to fetch referenced tweets: unable to decode tweet date: parsing time "invalid_date" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid_date" as "2006"`,
		},
		{
			name:       "unable to store referenced tweet",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Tweets: []*gotwitter.TweetObj{{
								CreatedAt: fakeDateStr,
							}},
						},
					}, nil)
				repository.EXPECT().Store(mock.Anything, mock.Anything).Return(errors.New("repo error"))
			},
			err: "unable to fetch referenced tweets: repo error",
		},
		{
			name:       "batch create tweets ID failed",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Tweets: []*gotwitter.TweetObj{
								fakeReferencedTweet4321,
								fakeReferencedTweet7654,
							},
						},
					}, nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel4321).Return(nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel7654).Return(nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"author1", "author2"}).Return(errors.New("user service error"))
			},
			err: "unable to fetch referenced tweets: user service error",
		},
		{
			name:       "ok",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Tweets: []*gotwitter.TweetObj{
								fakeReferencedTweet4321,
								fakeReferencedTweet7654,
							},
						},
					}, nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel4321).Return(nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel7654).Return(nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"author1", "author2"}).Return(nil)
				// Not super usefull to test the content of the stored tweet here, we are gonna
				// assert it with the output of this case
				repository.EXPECT().Store(mock.Anything, mock.Anything).Return(nil)
			},
			output: &models.Tweet{
				ID:        fakeCompleteTweetInput.ID,
				Text:      fakeCompleteTweetInput.Text,
				Published: fakeDate,
				Sensitive: fakeCompleteTweetInput.PossiblySensitive,
				ReferencedTweets: []models.Tweet{
					*fakeReferencedTweetModel4321,
					*fakeReferencedTweetModel7654,
				},
			},
		},
		{
			name:       "ok with 1 referenced tweet already exist",
			tweetInput: fakeCompleteTweetInput,
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				repository.EXPECT().GetTweet(mock.Anything, "4321").Return(&models.Tweet{
					ID:        "4321",
					Published: fakeDate,
				}, nil)
				repository.EXPECT().GetTweet(mock.Anything, "7654").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Tweets: []*gotwitter.TweetObj{
								fakeReferencedTweet7654,
							},
						},
					}, nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel7654).Return(nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"author2"}).Return(nil)
				// Not super usefull to test the content of the stored tweet here, we are gonna
				// assert it with the output of this case
				repository.EXPECT().Store(mock.Anything, mock.Anything).Return(nil)
			},
			output: &models.Tweet{
				ID:        fakeCompleteTweetInput.ID,
				Text:      fakeCompleteTweetInput.Text,
				Published: fakeDate,
				Sensitive: fakeCompleteTweetInput.PossiblySensitive,
				ReferencedTweets: []models.Tweet{
					*fakeReferencedTweetModel4321,
					*fakeReferencedTweetModel7654,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fakeUserService := mocks.NewUserService(t)
			fakeTwitterClient := mockstwitter.NewTwitterClient(t)
			fateTweetRepo := mockstwitterrepo.NewTweetRepository(t)

			if c.mocks != nil {
				c.mocks(fakeUserService, fakeTwitterClient, fateTweetRepo)
			}

			tweetSvc := NewTweetService(
				fakeUserService,
				fakeTwitterClient,
				fateTweetRepo,
			)

			tweet, err := tweetSvc.SaveTweetAndReferences(context.TODO(), c.tweetInput)
			if c.err != "" {
				require.EqualError(t, err, c.err)
			} else {
				require.NoError(t, err)
			}
			if c.output != nil {
				require.Equal(t, *c.output, *tweet)
			}
		})
	}
}
