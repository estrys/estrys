package twitter_test

import (
	"context"
	"testing"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/estrys/estrys/internal/logger/mocks"
	"github.com/estrys/estrys/internal/models"
	mocksuser "github.com/estrys/estrys/internal/repository/mocks"
	"github.com/estrys/estrys/internal/twitter"
	mockstwitter "github.com/estrys/estrys/internal/twitter/mocks"
	mocksworker "github.com/estrys/estrys/internal/worker/client/mocks"
	"github.com/estrys/estrys/internal/worker/tasks"
)

func Test_twitterPoller_Start(t *testing.T) {
	cases := []struct {
		name      string
		assertErr func(*testing.T, error)
		mocks     func(*mockstwitter.TwitterClient, *mocksuser.UserRepository, context.CancelFunc, *mocksworker.BackgroundWorkerClient)
	}{
		{
			name: "error while fetching users for the first time",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc, worker *mocksworker.BackgroundWorkerClient) {
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
				var startTime time.Time
				fakeTwitter.On("GetTweets", mock.Anything, "123", mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{
							Meta: &gotwitter.UserTimelineMeta{
								ResultCount: 0,
							},
						},
						nil,
					).Run(func(args mock.Arguments) {
					arg := args.Get(2).(gotwitter.UserTweetTimelineOpts)
					startTime = arg.StartTime
				})
				fakeTwitter.On("GetTweets", mock.Anything, "123", mock.MatchedBy(func(opts gotwitter.UserTweetTimelineOpts) bool {
					return opts.StartTime == startTime
				})).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{
							Raw: &gotwitter.TweetRaw{
								Tweets: []*gotwitter.TweetObj{
									{
										ID:                "1337",
										Text:              "tweet content",
										CreatedAt:         "2006-01-02T15:04:05Z",
										PossiblySensitive: true,
									},
								},
							},
							Meta: &gotwitter.UserTimelineMeta{
								NewestID:    "1",
								ResultCount: 1,
							},
						},
						nil,
					)

				fakeRepo.On("GetFollowers", mock.Anything, &models.User{
					ID:       "123",
					Username: "foobar",
				}).Once().Return(models.ActorSlice{
					{
						URL: "https://example.com/actor_url",
					},
					{
						URL: "https://example.com/another_actor_url",
					},
				}, nil)

				worker.On("Enqueue", mock.MatchedBy(func(t *asynq.Task) bool {
					payload := string(t.Payload())
					expectedPayload := `{"from":"foobar","to":"https://example.com/actor_url","tweet":{"id":"1337","text":"tweet content","published":"2006-01-02T15:04:05Z","sensitive":true}}`
					return t.Type() == tasks.TypeSendTweet && payload == expectedPayload
				})).Return(nil, nil)
				worker.On("Enqueue", mock.MatchedBy(func(t *asynq.Task) bool {
					payload := string(t.Payload())
					expectedPayload := `{"from":"foobar","to":"https://example.com/another_actor_url","tweet":{"id":"1337","text":"tweet content","published":"2006-01-02T15:04:05Z","sensitive":true}}`
					return t.Type() == tasks.TypeSendTweet && payload == expectedPayload
				})).Return(nil, nil)

				fakeTwitter.On("GetTweets", mock.Anything, "123", gotwitter.UserTweetTimelineOpts{
					MaxResults: 100,
					Excludes: []gotwitter.Exclude{
						gotwitter.ExcludeReplies,
					},
					TweetFields: []gotwitter.TweetField{
						gotwitter.TweetFieldID,
						gotwitter.TweetFieldText,
						gotwitter.TweetFieldCreatedAt,
						gotwitter.TweetFieldPossiblySensitve,
					},
					SinceID: "1",
				}).
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
							Raw: &gotwitter.TweetRaw{},
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
							Raw: &gotwitter.TweetRaw{},
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, cancel context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
							Raw: &gotwitter.TweetRaw{},
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
				fakeTwitter.On("GetTweets", mock.Anything, "123", gotwitter.UserTweetTimelineOpts{
					MaxResults: 100,
					Excludes: []gotwitter.Exclude{
						gotwitter.ExcludeReplies,
					},
					TweetFields: []gotwitter.TweetField{
						gotwitter.TweetFieldID,
						gotwitter.TweetFieldText,
						gotwitter.TweetFieldCreatedAt,
						gotwitter.TweetFieldPossiblySensitve,
					},
					SinceID: "1",
				}).
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
			fakeWorker := mocksworker.NewBackgroundWorkerClient(t)
			c.mocks(fakeTwitterClient, fakeUserRepo, cancel, fakeWorker)
			poller := twitter.NewPoller(
				mocks.NewNullLogger(),
				fakeTwitterClient,
				fakeUserRepo,
				fakeWorker,
			)
			err := poller.Start(fakeContext)
			c.assertErr(t, err)
		})
	}

}
