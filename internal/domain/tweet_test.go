package domain

import (
	"context"
	"net/url"
	"testing"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/estrys/estrys/internal/domain/mocks"
	loggermock "github.com/estrys/estrys/internal/logger/mocks"
	"github.com/estrys/estrys/internal/models"
	mockstwitter "github.com/estrys/estrys/internal/twitter/mocks"
	twittermodels "github.com/estrys/estrys/internal/twitter/models"
	mockstwitterrepo "github.com/estrys/estrys/internal/twitter/repository/mocks"
)

func TestTweetService_SaveTweetAndReferences(t *testing.T) {
	fakeDateStr := "2006-01-02T15:04:05Z"
	fakeDate, _ := time.Parse(time.RFC3339, fakeDateStr)

	fakeCompleteTweet := &gotwitter.TweetObj{
		ID:                "1234",
		AuthorID:          "mainAuthorID",
		Text:              `RT @someone: this text is gonna be truncated https://t.co/XaDNSVVB9l https://t.co/kdkjgnLWo5`,
		CreatedAt:         fakeDateStr,
		PossiblySensitive: true,
		Attachments: &gotwitter.TweetAttachmentsObj{
			MediaKeys: []string{"photo1"},
		},
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
		AuthorID:  "authorID1",
	}
	fakeReferencedTweet7654 := &gotwitter.TweetObj{
		ID:        "7654",
		CreatedAt: fakeDateStr,
		AuthorID:  "authorID2",
	}

	fakeReferencedTweetModel4321 := &twittermodels.Tweet{
		ID:             "4321",
		AuthorID:       "authorID1",
		AuthorUsername: "author1",
		ReferencedType: twittermodels.ReferenceTypeRetweet,
		Published:      fakeDate,
	}
	fakeReferencedTweetModel7654 := &twittermodels.Tweet{
		ID:             "7654",
		AuthorID:       "authorID2",
		AuthorUsername: "author2",
		ReferencedType: twittermodels.ReferenceTypeRetweet,
		Published:      fakeDate,
	}
	fakeMainAuthor := &models.User{
		Username:  "mainauthor",
		ID:        "mainAuthorID",
		CreatedAt: fakeDate,
	}
	fakeAuthor1 := &models.User{
		Username:  "author1",
		ID:        "authorID1",
		CreatedAt: fakeDate,
	}
	fakeAuthor2 := &models.User{
		Username:  "author2",
		ID:        "authorID2",
		CreatedAt: fakeDate,
	}

	fakeMediaURL, _ := url.Parse("https://example.com/photo.jpeg")
	fakeMedia := twittermodels.TweetMedia{
		Type: twittermodels.MediaTypePhoto,
		URL:  fakeMediaURL,
	}

	expectedTweetLookupOpts := gotwitter.TweetLookupOpts{
		Expansions: []gotwitter.Expansion{
			gotwitter.ExpansionAttachmentsMediaKeys,
			gotwitter.ExpansionReferencedTweetsID,
			gotwitter.ExpansionReferencedTweetsIDAuthorID,
		},
		MediaFields: []gotwitter.MediaField{
			gotwitter.MediaFieldType,
			gotwitter.MediaFieldURL,
			gotwitter.MediaFieldWidth,
			gotwitter.MediaFieldHeight,
		},
		TweetFields: []gotwitter.TweetField{
			gotwitter.TweetFieldID,
			gotwitter.TweetFieldAuthorID,
			gotwitter.TweetFieldText,
			gotwitter.TweetFieldCreatedAt,
			gotwitter.TweetFieldPossiblySensitve,
			gotwitter.TweetFieldReferencedTweets,
		},
	}

	cases := []struct {
		name    string
		tweetID string
		output  *twittermodels.Tweet
		mocks   func(*mocks.UserService, *mockstwitter.TwitterClient, *mockstwitterrepo.TweetRepository)
		err     string
	}{
		{
			name: "tweet already known",
			mocks: func(_ *mocks.UserService, _ *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(&twittermodels.Tweet{
					ID: "1234",
				}, nil)
			},
			output: &twittermodels.Tweet{
				ID: "1234",
			},
			tweetID: "1234",
		},
		{
			name: "error fetching tweet",
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(nil, errors.New("tweet lookup error"))
			},
			err:     `error while fetching tweet details from twitter: tweet lookup error`,
			tweetID: "1234",
		},
		{
			name: "no tweets returned",
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{},
					}, nil)
			},
			err:     `expected 1 tweets to be returned`,
			tweetID: "1234",
		},
		{
			name:    "all referenced tweets already exist",
			tweetID: "1234",
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Includes: &gotwitter.TweetRawIncludes{
								Users: []*gotwitter.UserObj{
									{
										ID:       fakeMainAuthor.ID,
										UserName: fakeMainAuthor.Username,
									},
								},
							},
							Tweets: []*gotwitter.TweetObj{fakeCompleteTweet},
						},
					}, nil)
				repository.EXPECT().GetTweet(mock.Anything, "4321").Times(1).Return(&twittermodels.Tweet{
					ID:       "4321",
					AuthorID: "author1",
				}, nil)
				repository.EXPECT().GetTweet(mock.Anything, "7654").Times(1).Return(&twittermodels.Tweet{
					ID:       "7654",
					AuthorID: "author2",
				}, nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"author1", "author2"}).
					Return([]*models.User{}, nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{fakeCompleteTweet.AuthorID}).
					Return([]*models.User{fakeMainAuthor}, nil)
				// Not super usefull to test the content of the stored tweet here, we are gonna
				// assert it with the output of this case
				repository.EXPECT().Store(mock.Anything, mock.Anything).Return(nil)
			},
			output: &twittermodels.Tweet{
				ID:             fakeCompleteTweet.ID,
				AuthorID:       fakeCompleteTweet.AuthorID,
				AuthorUsername: fakeMainAuthor.Username,
				Text:           "RT @someone: this text is gonna be truncated https://t.co/XaDNSVVB9l",
				Published:      fakeDate,
				Sensitive:      true,
				ReferencedTweets: []twittermodels.Tweet{
					{
						ID:       "4321",
						AuthorID: "author1",
					},
					{
						ID:       "7654",
						AuthorID: "author2",
					},
				},
			},
		},
		{
			name:    "batch create users from IDs failed",
			tweetID: "1234",
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{Tweets: []*gotwitter.TweetObj{fakeCompleteTweet}},
					}, nil)
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
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"authorID1", "authorID2"}).Return([]*models.User{fakeAuthor1, fakeAuthor2}, errors.New("user service error"))
			},
			err: "unable to create users for referenced tweets: user service error",
		},
		{
			name:    "unable to store referenced tweet",
			tweetID: "1234",
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{Tweets: []*gotwitter.TweetObj{fakeCompleteTweet}},
					}, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(&twittermodels.Tweet{
					ID: "fake",
				}, nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"", ""}).
					Return([]*models.User{}, nil)
				repository.EXPECT().Store(mock.Anything, mock.Anything).Times(1).Return(errors.New("repo error"))
			},
			err: "unable to save referenced tweet: repo error",
		},
		{
			name:    "unable to store tweet",
			tweetID: "1234",
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{Tweets: []*gotwitter.TweetObj{fakeCompleteTweet}},
					}, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(&twittermodels.Tweet{
					ID: "fake",
				}, nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"", ""}).
					Return([]*models.User{}, nil)
				repository.EXPECT().Store(mock.Anything, mock.Anything).Times(2).Return(nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{fakeCompleteTweet.AuthorID}).
					Return([]*models.User{fakeMainAuthor}, nil)
				repository.EXPECT().Store(mock.Anything, mock.Anything).Times(1).Return(errors.New("repo error"))
			},
			err: "unable to save tweet: repo error",
		},
		{
			name:    "tweeter client error while fetching referenced tweets",
			tweetID: "1234",
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{Tweets: []*gotwitter.TweetObj{fakeCompleteTweet}},
					}, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(nil, errors.New("twitter client error"))
			},
			err: "unable to fetch referenced tweets: error while fetching tweet details from twitter: twitter client error",
		},
		{
			name:    "tweeter client error while fetching referenced tweets",
			tweetID: "1234",
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{Tweets: []*gotwitter.TweetObj{fakeCompleteTweet}},
					}, nil)
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
			err: "unable to fetch referenced tweets: expected 2 tweets to be returned",
		},
		{
			name:    "invalid date in returned referenced tweet",
			tweetID: "1234",
			mocks: func(_ *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{Tweets: []*gotwitter.TweetObj{fakeCompleteTweet}},
					}, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Tweets: []*gotwitter.TweetObj{
								{
									CreatedAt: "invalid_date",
								},
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
			name:    "unable to fetch author for tweet",
			tweetID: "1234",
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{Tweets: []*gotwitter.TweetObj{fakeCompleteTweet}},
					}, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Includes: &gotwitter.TweetRawIncludes{
								Users: []*gotwitter.UserObj{
									{
										ID:       fakeAuthor1.ID,
										UserName: fakeAuthor1.Username,
									},
									{
										ID:       fakeAuthor2.ID,
										UserName: fakeAuthor2.Username,
									},
								},
							},
							Tweets: []*gotwitter.TweetObj{
								fakeReferencedTweet4321,
								fakeReferencedTweet7654,
							},
						},
					}, nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel4321).Return(nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel7654).Return(nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"authorID1", "authorID2"}).Return([]*models.User{fakeAuthor1, fakeAuthor2}, nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{fakeCompleteTweet.AuthorID}).Return(nil, errors.New("author create failed"))
			},
			err: "unable to fetch author for tweet: author create failed",
		},
		{
			name:    "ok",
			tweetID: "1234",
			mocks: func(userService *mocks.UserService, client *mockstwitter.TwitterClient, repository *mockstwitterrepo.TweetRepository) {
				repository.EXPECT().GetTweet(mock.Anything, "1234").Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"1234"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Includes: &gotwitter.TweetRawIncludes{
								Media: []*gotwitter.MediaObj{
									{
										Key:  "photo1",
										Type: string(fakeMedia.Type),
										URL:  fakeMedia.URL.String(),
									},
								},
								Users: []*gotwitter.UserObj{
									{
										ID:       fakeMainAuthor.ID,
										UserName: fakeMainAuthor.Username,
									},
								},
							},
							Tweets: []*gotwitter.TweetObj{fakeCompleteTweet},
						},
					}, nil)
				repository.EXPECT().GetTweet(mock.Anything, mock.Anything).Times(2).Return(nil, nil)
				client.EXPECT().GetTweets(mock.Anything, []string{"4321", "7654"}, expectedTweetLookupOpts).
					Return(&gotwitter.TweetLookupResponse{
						Raw: &gotwitter.TweetRaw{
							Includes: &gotwitter.TweetRawIncludes{
								Users: []*gotwitter.UserObj{
									{
										ID:       fakeAuthor1.ID,
										UserName: fakeAuthor1.Username,
									},
									{
										ID:       fakeAuthor2.ID,
										UserName: fakeAuthor2.Username,
									},
								},
							},
							Tweets: []*gotwitter.TweetObj{
								fakeReferencedTweet4321,
								fakeReferencedTweet7654,
							},
						},
					}, nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel4321).Return(nil)
				repository.EXPECT().Store(mock.Anything, fakeReferencedTweetModel7654).Return(nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{"authorID1", "authorID2"}).Return([]*models.User{fakeAuthor1, fakeAuthor2}, nil)
				userService.EXPECT().BatchCreateUsersFromIDs(mock.Anything, []string{fakeCompleteTweet.AuthorID}).Return([]*models.User{fakeMainAuthor}, nil)
				// Not super usefull to test the content of the stored tweet here, we are gonna
				// assert it with the output of this case
				repository.EXPECT().Store(mock.Anything, mock.Anything).Return(nil)
			},
			output: &twittermodels.Tweet{
				ID:             fakeCompleteTweet.ID,
				Text:           "RT @someone: this text is gonna be truncated https://t.co/XaDNSVVB9l",
				AuthorID:       fakeCompleteTweet.AuthorID,
				AuthorUsername: fakeMainAuthor.Username,
				Published:      fakeDate,
				Sensitive:      fakeCompleteTweet.PossiblySensitive,
				Medias: []twittermodels.TweetMedia{
					fakeMedia,
				},
				ReferencedTweets: []twittermodels.Tweet{
					*fakeReferencedTweetModel4321,
					*fakeReferencedTweetModel7654,
				},
			},
		},
	}

	nullLogger := loggermock.NewNullLogger()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fakeUserService := mocks.NewUserService(t)
			fakeTwitterClient := mockstwitter.NewTwitterClient(t)
			fateTweetRepo := mockstwitterrepo.NewTweetRepository(t)

			if c.mocks != nil {
				c.mocks(fakeUserService, fakeTwitterClient, fateTweetRepo)
			}

			tweetSvc := NewTweetService(
				nullLogger,
				fakeUserService,
				fakeTwitterClient,
				fateTweetRepo,
			)

			tweet, err := tweetSvc.SaveTweetAndReferences(context.TODO(), c.tweetID)
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
