package status_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/domain"
	"github.com/estrys/estrys/internal/domain/domainmodels"
	"github.com/estrys/estrys/internal/domain/mocks"
	"github.com/estrys/estrys/internal/domain/status"
	"github.com/estrys/estrys/internal/twitter/models"
	"github.com/estrys/estrys/internal/twitter/repository"
	mocks2 "github.com/estrys/estrys/internal/twitter/repository/mocks"
	"github.com/estrys/estrys/tests"
)

type StatusHandlerTestSuite struct {
	suite.Suite
	tests.HTTPTestSuite
}

func TestStatusHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(StatusHandlerTestSuite))
}

func (suite *StatusHandlerTestSuite) TestHandleStatus() {
	fakeProfileImage, _ := url.Parse("https://example.com/image.png")
	fakeUser := &domainmodels.User{
		Name:            "Foo Bar",
		Username:        "foobar",
		Description:     "foobar description",
		ProfileImageURL: fakeProfileImage,
	}
	fakeDate, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	fakeTweet := &models.Tweet{
		ID:             "1234",
		AuthorUsername: "foobar",
		Text:           "This is a fake tweet content",
		Published:      fakeDate,
		Sensitive:      false,
	}

	fakeRetweetedUser := &domainmodels.User{
		Name:            "Foo Bar",
		Username:        "fakeRTUser",
		Description:     "foobar description",
		ProfileImageURL: fakeProfileImage,
	}
	fakeRetweet := &models.Tweet{
		ID:             "4321",
		AuthorUsername: "fakeRTUser",
		ReferencedTweets: []models.Tweet{
			{
				ID:             "7654",
				AuthorUsername: "fakeRTUser",
				ReferencedType: models.ReferenceTypeRetweet,
				Text:           "Retweeted content",
				Published:      fakeDate,
			},
		},
	}

	cases := []tests.HTTPTestCase{
		{
			Name: "missing tweet id",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUser.Username}},
			},
			StatusCode: http.StatusBadRequest,
		},
		{
			Name: "missing username",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"id": fakeTweet.ID}},
			},
			StatusCode: http.StatusBadRequest,
		},
		{
			Name: "tweet not found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUser.Username, "id": fakeTweet.ID}},
			},
			Mock: func(t *testing.T) {
				fakeTweetRepo := mocks2.NewTweetRepository(t)
				fakeTweetRepo.On("GetTweet", mock.Anything, fakeTweet.ID).Return(
					nil, errors.New("tweet not found"),
				)
				_ = dic.Register[repository.TweetRepository](fakeTweetRepo)
			},
			StatusCode: http.StatusNotFound,
			GoldenFile: "errors/tweet_not_found.json",
		},
		{
			Name: "tweet username mismatch",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": "invalidUsername", "id": fakeTweet.ID}},
			},
			Mock: func(t *testing.T) {
				fakeTweetRepo := mocks2.NewTweetRepository(t)
				fakeTweetRepo.On("GetTweet", mock.Anything, fakeTweet.ID).Return(
					fakeTweet, nil,
				)
				_ = dic.Register[repository.TweetRepository](fakeTweetRepo)
			},
			StatusCode: http.StatusBadRequest,
			GoldenFile: "errors/username_mismatch.json",
		},
		{
			Name: "twitter user not found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUser.Username, "id": fakeTweet.ID}},
			},
			Mock: func(t *testing.T) {
				fakeTweetRepo := mocks2.NewTweetRepository(t)
				fakeTweetRepo.On("GetTweet", mock.Anything, fakeTweet.ID).Return(
					fakeTweet, nil,
				)
				_ = dic.Register[repository.TweetRepository](fakeTweetRepo)

				fakeUserService := mocks.NewUserService(t)
				fakeUserService.On("GetFullUser", mock.Anything, fakeUser.Username).Return(
					nil, errors.WithStack(domain.ErrUserDoesNotExist),
				)
				_ = dic.Register[domain.UserService](fakeUserService)
			},
			StatusCode: http.StatusNotFound,
			GoldenFile: "errors/user_not_found.json",
		},
		{
			Name: "ok",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUser.Username, "id": fakeTweet.ID}},
			},
			Mock: func(t *testing.T) {
				fakeUserService := mocks.NewUserService(t)
				fakeUserService.On("GetFullUser", mock.Anything, fakeUser.Username).Return(
					fakeUser, nil,
				)
				_ = dic.Register[domain.UserService](fakeUserService)

				fakeTweetRepo := mocks2.NewTweetRepository(t)
				fakeTweetRepo.On("GetTweet", mock.Anything, fakeTweet.ID).Return(
					fakeTweet, nil,
				)
				_ = dic.Register[repository.TweetRepository](fakeTweetRepo)
			},
			StatusCode: http.StatusOK,
			GoldenFile: "status.html",
		},
		{
			Name: "retweet_ok",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeRetweetedUser.Username, "id": fakeRetweet.ID}},
			},
			Mock: func(t *testing.T) {
				fakeTweetRepo := mocks2.NewTweetRepository(t)
				fakeTweetRepo.On("GetTweet", mock.Anything, fakeRetweet.ID).Return(
					fakeRetweet, nil,
				)

				fakeUserService := mocks.NewUserService(t)
				fakeUserService.On("GetFullUser", mock.Anything, fakeRetweetedUser.Username).Return(
					fakeRetweetedUser, nil,
				)
				fakeUserService.On("GetFullUser", mock.Anything, fakeRetweet.Retweet().AuthorUsername).Return(
					fakeRetweetedUser, nil,
				)

				_ = dic.Register[domain.UserService](fakeUserService)
				_ = dic.Register[repository.TweetRepository](fakeTweetRepo)
			},
			StatusCode: http.StatusOK,
			GoldenFile: "status_retweet.html",
		},
	}

	suite.RunHTTPCases(suite.T(), status.HandleStatus, cases)
}
