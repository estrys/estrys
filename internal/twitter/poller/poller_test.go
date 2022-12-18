package poller_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocksdomain "github.com/estrys/estrys/internal/domain/mocks"
	"github.com/estrys/estrys/internal/logger/mocks"
	"github.com/estrys/estrys/internal/models"
	mocksuser "github.com/estrys/estrys/internal/repository/mocks"
	mockstwitter "github.com/estrys/estrys/internal/twitter/mocks"
	twittermodels "github.com/estrys/estrys/internal/twitter/models"
	"github.com/estrys/estrys/internal/twitter/poller"
	mocksworker "github.com/estrys/estrys/internal/worker/client/mocks"
	"github.com/estrys/estrys/internal/worker/tasks"
)

func Test_twitterPoller_Start(t *testing.T) {
	fakeDateStr := "2006-01-02T15:04:05Z"
	fakeDate, _ := time.Parse(time.RFC3339, fakeDateStr)
	cases := []struct {
		name      string
		assertErr func(*testing.T, error)
		mocks     func(*mockstwitter.TwitterClient, *mocksuser.UserRepository, *mocksdomain.TweetService, context.CancelFunc, *mocksworker.BackgroundWorkerClient)
	}{
		{
			name: "error while fetching users for the first time",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ *mocksdomain.TweetService, _ context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ *mocksdomain.TweetService, _ context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, tweetService *mocksdomain.TweetService, cancel context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
					"GetUserTweets",
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, tweetService *mocksdomain.TweetService, cancel context.CancelFunc, worker *mocksworker.BackgroundWorkerClient) {
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
				fakeTwitter.On("GetUserTweets", mock.Anything, "123", mock.Anything).
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
				fakeTwitter.On("GetUserTweets", mock.Anything, "123", mock.MatchedBy(func(opts gotwitter.UserTweetTimelineOpts) bool {
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

				tweetService.EXPECT().SaveTweetAndReferences(mock.Anything, &gotwitter.TweetObj{
					ID:                "1337",
					Text:              "tweet content",
					CreatedAt:         fakeDateStr,
					PossiblySensitive: true,
				}).Return(&twittermodels.Tweet{
					ID:        "1337",
					Text:      "tweet content",
					Published: fakeDate,
					Sensitive: true,
				}, nil)

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

				worker.On("Enqueue", mock.MatchedBy(func(task *asynq.Task) bool {
					payload := map[string]any{}
					_ = json.Unmarshal(task.Payload(), &payload)
					expectedPayload := map[string]any{
						"from":     "foobar",
						"to":       "https://example.com/actor_url",
						"tweet_id": "1337",
					}
					match := task.Type() == tasks.TypeSendTweet &&
						expectedPayload["from"] == payload["from"] &&
						expectedPayload["to"] == payload["to"] &&
						reflect.DeepEqual(expectedPayload["tweet"], payload["tweet"]) &&
						payload["trace_id"] != ""
					return match
				})).Return(nil, nil)
				worker.On("Enqueue", mock.MatchedBy(func(task *asynq.Task) bool {
					payload := map[string]any{}
					_ = json.Unmarshal(task.Payload(), &payload)
					expectedPayload := map[string]any{
						"from":     "foobar",
						"to":       "https://example.com/another_actor_url",
						"tweet_id": "1337",
					}
					match := task.Type() == tasks.TypeSendTweet &&
						expectedPayload["from"] == payload["from"] &&
						expectedPayload["to"] == payload["to"] &&
						reflect.DeepEqual(expectedPayload["tweet"], payload["tweet"]) &&
						payload["trace_id"] != ""
					return match
				})).Return(nil, nil)

				fakeTwitter.On("GetUserTweets", mock.Anything, "123", gotwitter.UserTweetTimelineOpts{
					MaxResults: 100,
					Excludes: []gotwitter.Exclude{
						gotwitter.ExcludeReplies,
					},
					TweetFields: []gotwitter.TweetField{
						gotwitter.TweetFieldID,
						gotwitter.TweetFieldText,
						gotwitter.TweetFieldCreatedAt,
						gotwitter.TweetFieldPossiblySensitve,
						gotwitter.TweetFieldReferencedTweets,
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
			name: "failed GetUserTweets during Polling user timeline",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ *mocksdomain.TweetService, cancel context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
				fakeTwitter.On("GetUserTweets", mock.Anything, "123", mock.Anything).
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
				fakeTwitter.On("GetUserTweets", mock.Anything, mock.Anything, mock.Anything).
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
			name: "Keep polling on the same user as we get errors on GetUserTweets",
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ *mocksdomain.TweetService, cancel context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
				fakeTwitter.On("GetUserTweets", mock.Anything, "123", mock.Anything).
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
				fakeTwitter.On("GetUserTweets", mock.Anything, "123", mock.Anything).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{},
						errors.New("unexpected error"),
					)
				fakeTwitter.On("GetUserTweets", mock.Anything, "123", mock.Anything).
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
			mocks: func(fakeTwitter *mockstwitter.TwitterClient, fakeRepo *mocksuser.UserRepository, _ *mocksdomain.TweetService, cancel context.CancelFunc, _ *mocksworker.BackgroundWorkerClient) {
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
				fakeTwitter.On("GetUserTweets", mock.Anything, "123", mock.Anything).
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
				fakeTwitter.On("GetUserTweets", mock.Anything, "123", gotwitter.UserTweetTimelineOpts{
					MaxResults: 100,
					Excludes: []gotwitter.Exclude{
						gotwitter.ExcludeReplies,
					},
					TweetFields: []gotwitter.TweetField{
						gotwitter.TweetFieldID,
						gotwitter.TweetFieldText,
						gotwitter.TweetFieldCreatedAt,
						gotwitter.TweetFieldPossiblySensitve,
						gotwitter.TweetFieldReferencedTweets,
					},
					SinceID: "1",
				}).
					Once().
					Return(
						&gotwitter.UserTweetTimelineResponse{Meta: &gotwitter.UserTimelineMeta{ResultCount: 0}},
						nil,
					)

				fakeTwitter.On("GetUserTweets", mock.Anything, "124", mock.MatchedBy(func(opts gotwitter.UserTweetTimelineOpts) bool { return !opts.StartTime.IsZero() })).
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
			fakeTweetService := mocksdomain.NewTweetService(t)
			fakeWorker := mocksworker.NewBackgroundWorkerClient(t)
			c.mocks(fakeTwitterClient, fakeUserRepo, fakeTweetService, cancel, fakeWorker)
			poller := poller.NewPoller(
				mocks.NewNullLogger(),
				fakeTwitterClient,
				fakeUserRepo,
				fakeTweetService,
				fakeWorker,
			)
			err := poller.Start(fakeContext)
			c.assertErr(t, err)
		})
	}

}
