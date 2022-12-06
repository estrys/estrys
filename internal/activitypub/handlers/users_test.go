package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/estrys/estrys/internal/activitypub/auth"
	"github.com/estrys/estrys/internal/activitypub/handlers"
	"github.com/estrys/estrys/internal/cache"
	mockscache "github.com/estrys/estrys/internal/cache/mocks"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/repository"
	mocksuser "github.com/estrys/estrys/internal/repository/mocks"
	"github.com/estrys/estrys/internal/twitter"
	mockstwitter "github.com/estrys/estrys/internal/twitter/mocks"
	"github.com/estrys/estrys/internal/worker/client"
	"github.com/estrys/estrys/internal/worker/client/mocks"
	"github.com/estrys/estrys/internal/worker/tasks"
	"github.com/estrys/estrys/tests"
)

type UserHandlerTestSuite struct {
	suite.Suite
	tests.HTTPTestSuite
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

var (
	fakeUserName             = "foobar"
	fakeUserCreatedAtStr     = "2011-05-05T13:21:56.000Z"
	fakeUserProfileUrl, _    = url.Parse("https://example.com/image.jpg")
	fakeUserCreatedAt, _     = time.Parse(time.RFC3339, fakeUserCreatedAtStr)
	fakePrivateKey, _        = os.ReadFile(path.Join("testdata", "key.pem"))
	privKey, _               = pem.Decode(fakePrivateKey)
	authenticatedHTTPContext = context.WithValue(context.Background(),
		auth.HTTPSignature,
		func() *bool { v := true; return &v }(),
	)
)

func (suite *UserHandlerTestSuite) TestHandleUser() {
	cases := []tests.HTTPTestCase{
		{
			Name: "user found with cache miss",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get",
					mock.Anything,
					strings.Join([]string{"twitter", "user", fakeUserName}, "/")).
					Times(1).Return(nil, cache.ErrMiss)
				fakeCache.On(
					"Set",
					mock.Anything,
					strings.Join([]string{"twitter", "user", fakeUserName}, "/"),
					mock.Anything).
					Times(1).
					Return(nil)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeTwitterBackend := mockstwitter.NewBackend(t)
				fakeTwitterBackend.On(
					"UserNameLookup",
					mock.Anything,
					[]string{fakeUserName},
					gotwitter.UserLookupOpts{
						UserFields: []gotwitter.UserField{
							gotwitter.UserFieldDescription,
							gotwitter.UserFieldName,
							gotwitter.UserFieldProfileImageURL,
							gotwitter.UserFieldCreatedAt,
							gotwitter.UserFieldPublicMetrics,
						},
					},
				).Return(&gotwitter.UserLookupResponse{
					Raw: &gotwitter.UserRaw{
						Users: []*gotwitter.UserObj{
							{
								Name:            "Foo Bar",
								CreatedAt:       fakeUserCreatedAtStr,
								Description:     "This is a fake twitter user",
								ProfileImageURL: fakeUserProfileUrl.String(),
								PublicMetrics: &gotwitter.UserMetricsObj{
									Followers: 13,
									Following: 37,
									Tweets:    42,
								},
							},
						},
					},
				}, nil)
				_ = dic.Register[twitter.Backend](fakeTwitterBackend)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(
					&models.User{
						Username:   fakeUserName,
						PrivateKey: privKey.Bytes,
						CreatedAt:  fakeUserCreatedAt,
					}, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusOK,
			GoldenFile: "user/ok",
		},
		{
			Name: "user no found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, sql.ErrNoRows)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusNotFound,
		},
		{
			Name: "twitter user no found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get",
					mock.Anything,
					strings.Join([]string{"twitter", "user", fakeUserName}, "/")).
					Times(1).Return(nil, cache.ErrMiss)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeTwitterBackend := mockstwitter.NewBackend(t)
				fakeTwitterBackend.On(
					"UserNameLookup",
					mock.Anything,
					[]string{fakeUserName},
					gotwitter.UserLookupOpts{
						UserFields: []gotwitter.UserField{
							gotwitter.UserFieldDescription,
							gotwitter.UserFieldName,
							gotwitter.UserFieldProfileImageURL,
							gotwitter.UserFieldCreatedAt,
							gotwitter.UserFieldPublicMetrics,
						},
					},
				).Return(&gotwitter.UserLookupResponse{
					Raw: &gotwitter.UserRaw{
						Errors: []*gotwitter.ErrorObj{
							{
								Type: twitter.TwitterErrorTypeNotFound,
							},
						},
					},
				}, nil)
				_ = dic.Register[twitter.Backend](fakeTwitterBackend)
			},
			StatusCode: http.StatusNotFound,
		},
		{
			Name: "unexpected error",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get",
					mock.Anything,
					strings.Join([]string{"twitter", "user", fakeUserName}, "/")).
					Times(1).Return(nil, cache.ErrMiss)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeTwitterBackend := mockstwitter.NewBackend(t)
				fakeTwitterBackend.On(
					"UserNameLookup",
					mock.Anything,
					[]string{fakeUserName},
					gotwitter.UserLookupOpts{
						UserFields: []gotwitter.UserField{
							gotwitter.UserFieldDescription,
							gotwitter.UserFieldName,
							gotwitter.UserFieldProfileImageURL,
							gotwitter.UserFieldCreatedAt,
							gotwitter.UserFieldPublicMetrics,
						},
					},
				).Return(nil, errors.New("unexpected error"))
				_ = dic.Register[twitter.Backend](fakeTwitterBackend)
			},
			StatusCode: http.StatusInternalServerError,
		},
	}

	suite.RunHTTPCases(suite.T(), handlers.HandleUser, cases)
}

func (suite *UserHandlerTestSuite) TestHandleFollowers() {
	cases := []tests.HTTPTestCase{
		{
			Name: "user found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get", mock.Anything, mock.Anything).Return(&twitter.User{
					ProfileImageURL: fakeUserProfileUrl,
					Followers:       1337,
				}, nil)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(
					&models.User{
						Username:   fakeUserName,
						PrivateKey: privKey.Bytes,
						CreatedAt:  fakeUserCreatedAt,
					}, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusOK,
			GoldenFile: "followers/ok",
		},
		{
			Name: "user no found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, sql.ErrNoRows)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusNotFound,
		},
		{
			Name: "unexpected error",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get",
					mock.Anything,
					strings.Join([]string{"twitter", "user", fakeUserName}, "/")).
					Times(1).Return(nil, cache.ErrMiss)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeTwitterBackend := mockstwitter.NewBackend(t)
				fakeTwitterBackend.On(
					"UserNameLookup",
					mock.Anything,
					[]string{fakeUserName},
					gotwitter.UserLookupOpts{
						UserFields: []gotwitter.UserField{
							gotwitter.UserFieldDescription,
							gotwitter.UserFieldName,
							gotwitter.UserFieldProfileImageURL,
							gotwitter.UserFieldCreatedAt,
							gotwitter.UserFieldPublicMetrics,
						},
					},
				).Return(nil, errors.New("unexpected error"))
				_ = dic.Register[twitter.Backend](fakeTwitterBackend)
			},
			StatusCode: http.StatusInternalServerError,
		},
	}

	suite.RunHTTPCases(suite.T(), handlers.HandleFollowers, cases)
}

func (suite *UserHandlerTestSuite) TestHandleFollowing() {
	cases := []tests.HTTPTestCase{
		{
			Name: "user found with cache miss",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get", mock.Anything, mock.Anything).Return(&twitter.User{
					ProfileImageURL: fakeUserProfileUrl,
					Following:       1337,
				}, nil)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(
					&models.User{
						Username:   fakeUserName,
						PrivateKey: privKey.Bytes,
						CreatedAt:  fakeUserCreatedAt,
					}, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusOK,
			GoldenFile: "following/ok",
		},
		{
			Name: "twitter user no found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, sql.ErrNoRows)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusNotFound,
		},
		{
			Name: "unexpected error",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get",
					mock.Anything,
					strings.Join([]string{"twitter", "user", fakeUserName}, "/")).
					Times(1).Return(nil, cache.ErrMiss)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeTwitterBackend := mockstwitter.NewBackend(t)
				fakeTwitterBackend.On(
					"UserNameLookup",
					mock.Anything,
					[]string{fakeUserName},
					gotwitter.UserLookupOpts{
						UserFields: []gotwitter.UserField{
							gotwitter.UserFieldDescription,
							gotwitter.UserFieldName,
							gotwitter.UserFieldProfileImageURL,
							gotwitter.UserFieldCreatedAt,
							gotwitter.UserFieldPublicMetrics,
						},
					},
				).Return(nil, errors.New("unexpected error"))
				_ = dic.Register[twitter.Backend](fakeTwitterBackend)
			},
			StatusCode: http.StatusInternalServerError,
		},
	}

	suite.RunHTTPCases(suite.T(), handlers.HandleFollowing, cases)
}

func (suite *UserHandlerTestSuite) TestHandleOutbox() {
	cases := []tests.HTTPTestCase{
		{
			Name: "user found with cache miss",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get", mock.Anything, mock.Anything).Return(&twitter.User{
					ProfileImageURL: fakeUserProfileUrl,
					Tweets:          1337,
				}, nil)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(
					&models.User{
						Username:   fakeUserName,
						PrivateKey: privKey.Bytes,
						CreatedAt:  fakeUserCreatedAt,
					}, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusOK,
			GoldenFile: "outbox/ok",
		},
		{
			Name: "twitter user no found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, sql.ErrNoRows)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusNotFound,
		},
		{
			Name: "unexpected error",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
			},
			Mock: func(t *testing.T) {
				fakeCache := mockscache.NewCache[twitter.User](t)
				fakeCache.On(
					"Get",
					mock.Anything,
					strings.Join([]string{"twitter", "user", fakeUserName}, "/")).
					Times(1).Return(nil, cache.ErrMiss)
				_ = dic.Register[cache.Cache[twitter.User]](fakeCache)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, fakeUserName).Return(nil, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeTwitterBackend := mockstwitter.NewBackend(t)
				fakeTwitterBackend.On(
					"UserNameLookup",
					mock.Anything,
					[]string{fakeUserName},
					gotwitter.UserLookupOpts{
						UserFields: []gotwitter.UserField{
							gotwitter.UserFieldDescription,
							gotwitter.UserFieldName,
							gotwitter.UserFieldProfileImageURL,
							gotwitter.UserFieldCreatedAt,
							gotwitter.UserFieldPublicMetrics,
						},
					},
				).Return(nil, errors.New("unexpected error"))
				_ = dic.Register[twitter.Backend](fakeTwitterBackend)
			},
			StatusCode: http.StatusInternalServerError,
		},
	}

	suite.RunHTTPCases(suite.T(), handlers.HandleOutbox, cases)
}

func (suite *UserHandlerTestSuite) TestHandleInbox() {
	cases := []tests.HTTPTestCase{
		{
			Name:       "request not authenticated",
			StatusCode: http.StatusForbidden,
		},
		{
			Name: "twitter user no found",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, twitter.UsernameNotFoundError{fakeUserName},
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)
			},
			StatusCode: http.StatusNotFound,
		},
		//TODO add more error cases here
		{
			Name: "invalid activity",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
				tests.RequestBodyFromFile{FilePath: "inbox/input/unsupported_activity"},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, nil,
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)
			},
			StatusCode: http.StatusBadRequest,
		},
	}

	suite.RunHTTPCases(suite.T(), handlers.HandleInbox, cases)
}

func (suite *UserHandlerTestSuite) TestHandleInbox_Follow() {
	cases := []tests.HTTPTestCase{
		//TODO add more error cases here
		{
			Name: "object domain mismatch",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
				tests.RequestBodyFromFile{FilePath: "inbox/input/domain_mismatch"},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, nil,
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)

			},
			StatusCode: http.StatusBadRequest,
		},
		{
			Name: "no user in db",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
				tests.RequestBodyFromFile{FilePath: "inbox/input/no_user_in_db"},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, nil,
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "notfound").Return(
					nil, sql.ErrNoRows)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusBadRequest,
		},
		{
			Name: "actor not allowed",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
				tests.RequestBodyFromFile{FilePath: "inbox/input/valid_follow"},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, nil,
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "validuser").Return(
					&models.User{
						Username:   "validuser",
						PrivateKey: privKey.Bytes,
						CreatedAt:  fakeUserCreatedAt,
					}, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeWorker := mocks.NewBackgroundWorkerClient(t)
				fakeWorker.On("Enqueue", mock.MatchedBy(func(task *asynq.Task) bool {
					expectedAccept := tasks.RejectFollowInput{
						Username: "validuser",
						Activity: map[string]interface{}{
							"@context": "https://www.w3.org/ns/activitystreams",
							"type":     "Follow",
							"actor":    "https://another-instance.example.com/users/validactor",
							"id":       "https://another-instance.example.com/bdd01ced-d657-4847-a266-2c43e1cd8dc5",
							"object":   "https://example.com/users/validuser",
						},
					}
					accept := tasks.RejectFollowInput{}
					err := json.Unmarshal(task.Payload(), &accept)
					if err != nil {
						return assert.NoError(t, err)
					}
					return assert.Equal(t, expectedAccept, accept)
				})).Return(nil, nil)
				_ = dic.Register[client.BackgroundWorkerClient](fakeWorker)
			},
			StatusCode: http.StatusForbidden,
		},
		{
			Name: "valid follow",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
				tests.RequestBodyFromFile{FilePath: "inbox/input/valid_follow"},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, nil,
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)

				fakeUser := &models.User{
					Username:   "validuser",
					PrivateKey: privKey.Bytes,
					CreatedAt:  fakeUserCreatedAt,
				}
				fakeActor := &models.Actor{}
				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "validuser").Return(fakeUser, nil)
				fakeUserRepo.On("Follow", mock.Anything, fakeUser, fakeActor).Return(nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeActorUrl, _ := url.Parse("https://another-instance.example.com/users/validactor")
				fakeActorRepo := mocksuser.NewActorRepository(t)
				fakeActorRepo.On("Get", mock.Anything, fakeActorUrl).Return(fakeActor, nil)
				_ = dic.Register[repository.ActorRepository](fakeActorRepo)

				fakeWorker := mocks.NewBackgroundWorkerClient(t)
				fakeWorker.On("Enqueue", mock.MatchedBy(func(task *asynq.Task) bool {
					expectedAccept := tasks.AcceptFollowInput{
						Username: "validuser",
						Activity: map[string]interface{}{
							"@context": "https://www.w3.org/ns/activitystreams",
							"type":     "Follow",
							"actor":    "https://another-instance.example.com/users/validactor",
							"id":       "https://another-instance.example.com/bdd01ced-d657-4847-a266-2c43e1cd8dc5",
							"object":   "https://example.com/users/validuser",
						},
					}
					accept := tasks.AcceptFollowInput{}
					err := json.Unmarshal(task.Payload(), &accept)
					if err != nil {
						return assert.NoError(t, err)
					}
					return assert.Equal(t, expectedAccept, accept)
				})).Return(nil, nil)
				_ = dic.Register[client.BackgroundWorkerClient](fakeWorker)

				viper.Set("allowed_users", "validactor@another-instance.example.com")
			},
			StatusCode: http.StatusAccepted,
		},
	}

	suite.RunHTTPCases(suite.T(), handlers.HandleInbox, cases)
}

func (suite *UserHandlerTestSuite) TestHandleInbox_UnFollow() {
	cases := []tests.HTTPTestCase{
		{
			Name: "no user in db",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
				tests.RequestBodyFromFile{FilePath: "inbox/input/no_user_in_db"},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, nil,
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)

				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "notfound").Return(
					nil, sql.ErrNoRows)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
			StatusCode: http.StatusBadRequest,
		},
		{
			Name: "can only undo follow",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
				tests.RequestBodyFromFile{FilePath: "inbox/input/invalid_undo_reject"},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, nil,
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)
			},
			StatusCode: http.StatusBadRequest,
		},
		{
			Name: "valid undo follow",
			RequestOptions: []tests.RequestOption{
				tests.RequestParams{Params: map[string]string{"username": fakeUserName}},
				tests.RequestContext{Context: authenticatedHTTPContext},
				tests.RequestBodyFromFile{FilePath: "inbox/input/valid_undo_follow"},
			},
			Mock: func(t *testing.T) {
				fakeTwitterClient := mockstwitter.NewTwitterClient(t)
				fakeTwitterClient.On("GetUser", mock.Anything, fakeUserName).Return(
					nil, nil,
				)
				_ = dic.Register[twitter.TwitterClient](fakeTwitterClient)

				fakeUser := &models.User{
					Username:   "validuser",
					PrivateKey: privKey.Bytes,
					CreatedAt:  fakeUserCreatedAt,
				}
				fakeActor := &models.Actor{}
				fakeUserRepo := mocksuser.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "validuser").Return(fakeUser, nil)
				fakeUserRepo.On("UnFollow", mock.Anything, fakeUser, fakeActor).Return(nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeActorUrl, _ := url.Parse("https://another-instance.example.com/users/validactor")
				fakeActorRepo := mocksuser.NewActorRepository(t)
				fakeActorRepo.On("Get", mock.Anything, fakeActorUrl).Return(fakeActor, nil)
				_ = dic.Register[repository.ActorRepository](fakeActorRepo)

				viper.Set("allowed_users", "validactor@another-instance.example.com")
			},
			StatusCode: http.StatusAccepted,
		},
	}

	suite.RunHTTPCases(suite.T(), handlers.HandleInbox, cases)
}
