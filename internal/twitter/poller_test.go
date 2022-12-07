package twitter_test

import (
	"context"
	"testing"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/estrys/estrys/internal/logger/mocks"
	"github.com/estrys/estrys/internal/models"
	mocksuser "github.com/estrys/estrys/internal/repository/mocks"
	"github.com/estrys/estrys/internal/twitter"
	mockstwitter "github.com/estrys/estrys/internal/twitter/mocks"
)

func Test_twitterPoller_Start(t *testing.T) {
	cases := []struct {
		name      string
		assertErr func(*testing.T, error)
		mocks     func(*mockstwitter.TwitterClient, *mocksuser.UserRepository, context.CancelFunc)
	}{
		{
			name: "error while fetching users for the first time",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ context.CancelFunc) {
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Return(
						nil,
						errors.New("unexpected error"),
					)
			},
			assertErr: func(t *testing.T, err error) {
				require.Error(t, err, "unexpected error")
			},
		},
		{
			name: "error while fetching users while polling for new users",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ context.CancelFunc) {
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						nil,
						nil,
					)
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						nil,
						errors.New("unexpected error"),
					)
			},
			assertErr: func(t *testing.T, err error) {
				require.Error(t, err, "unexpected error")
			},
		},
		{
			name: "error while fetching users while polling for new users",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc) {
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						models.UserSlice{
							{
								ID:       "123",
								Username: "foobar",
							},
						},
						nil,
					)
				fakeTwitter.On(
					"GetTweets",
					mock.Anything,
					"123",
					mock.MatchedBy(func(opts gotwitter.UserTweetTimelineOpts) bool { return !opts.StartTime.IsZero() }),
				).Once().Return(nil, errors.New("unexpected error")).
					Run(func(args mock.Arguments) {
						cancel()
					})
			},
			assertErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Polling user timeline",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc) {
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Return(
						models.UserSlice{
							{
								ID:       "123",
								Username: "foobar",
							},
						},
						nil,
					)
				fakeTwitter.On("GetTweets", mock.Anything, "123", mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{
							Meta: &gotwitter.UserTimelineMeta{
								NewestID:    "1",
								ResultCount: 1,
							},
						},
						nil,
					)
				fakeTwitter.On("GetTweets", mock.Anything, "123", gotwitter.UserTweetTimelineOpts{SinceID: "1"}).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{
							Meta: &gotwitter.UserTimelineMeta{
								ResultCount: 0,
							},
						},
						nil,
					).
					Run(func(args mock.Arguments) {
						cancel()
					})
			},
			assertErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "failed GetTweets during Polling user timeline",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc) {
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						models.UserSlice{
							{
								ID:       "123",
								Username: "foobar",
							},
						},
						nil,
					)
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						models.UserSlice{
							{
								ID:       "123",
								Username: "foobar",
							},
						},
						nil,
					)
				fakeTwitter.On("GetTweets", mock.Anything, "123", mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{
							Meta: &gotwitter.UserTimelineMeta{
								NewestID:    "1",
								ResultCount: 1,
							},
						},
						nil,
					)
				fakeTwitter.On("GetTweets", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{},
						errors.New("unexpected error"),
					).
					Run(func(args mock.Arguments) {
						cancel()
					})
			},
			assertErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Keep polling on the same user as we get errors on GetTweets",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc) {
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						models.UserSlice{
							{
								ID:       "123",
								Username: "foobar",
							},
						},
						nil,
					)
				fakeTwitter.On("GetTweets", mock.Anything, "123", mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{
							Meta: &gotwitter.UserTimelineMeta{
								NewestID:    "1",
								ResultCount: 1,
							},
						},
						nil,
					)
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						models.UserSlice{
							{
								ID:       "123",
								Username: "foobar",
							},
							{
								ID:       "124",
								Username: "barbaz",
							},
						},
						nil,
					)
				fakeTwitter.On("GetTweets", mock.Anything, "123", mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{},
						errors.New("unexpected error"),
					)
				fakeTwitter.On("GetTweets", mock.Anything, "123", mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{},
						errors.New("unexpected error"),
					).
					Run(func(args mock.Arguments) {
						cancel()
					})
			},
			assertErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Polling multiple users",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc) {
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						models.UserSlice{
							{
								ID:       "123",
								Username: "foobar",
							},
						},
						nil,
					)
				fakeTwitter.On("GetTweets", mock.Anything, "123", mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{
							Meta: &gotwitter.UserTimelineMeta{
								NewestID:    "1",
								ResultCount: 1,
							},
						},
						nil,
					)
				fakeRepo.On("GetWithFollowers", mock.Anything).
					Once().
					Return(
						models.UserSlice{
							{
								ID:       "123",
								Username: "foobar",
							},
							{
								ID:       "124",
								Username: "barbaz",
							},
						},
						nil,
					)
				fakeTwitter.On("GetTweets", mock.Anything, "123", gotwitter.UserTweetTimelineOpts{SinceID: "1"}).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{Meta: &gotwitter.UserTimelineMeta{ResultCount: 0}},
						nil,
					)

				fakeTwitter.On("GetTweets", mock.Anything, "124", mock.MatchedBy(func(opts gotwitter.UserTweetTimelineOpts) bool { return !opts.StartTime.IsZero() })).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{Meta: &gotwitter.UserTimelineMeta{ResultCount: 0}},
						nil,
					).
					Run(func(args mock.Arguments) {
						cancel()
					})
			},
			assertErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fakeContext, cancel := context.WithCancel(context.Background())
			fakeTwitterClient := mockstwitter.NewTwitterClient(t)
			fakeUserRepo := mocksuser.NewUserRepository(t)
			c.mocks(fakeTwitterClient, fakeUserRepo, cancel)
			poller := twitter.NewPoller(
				mocks.NewNullLogger(),
				fakeTwitterClient,
				fakeUserRepo,
			)
			err := poller.Start(fakeContext)
			c.assertErr(t, err)
		})
	}

}
