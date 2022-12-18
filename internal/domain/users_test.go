package domain

import (
	"context"
	"database/sql"
	"testing"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/estrys/estrys/internal/crypto"
	httpmock "github.com/estrys/estrys/internal/http/mocks"
	"github.com/estrys/estrys/internal/logger/mocks"
	"github.com/estrys/estrys/internal/models"
	userrepository "github.com/estrys/estrys/internal/repository"
	mocksuser "github.com/estrys/estrys/internal/repository/mocks"
	"github.com/estrys/estrys/internal/twitter"
	mockstwitter "github.com/estrys/estrys/internal/twitter/mocks"
)

func Test_userService_BatchCreateUsers(t *testing.T) {
	tests := []struct {
		name                string
		allowedTwitterUsers []string
		mocks               func(*mockstwitter.TwitterClient, *mocksuser.UserRepository)
		err                 error
	}{
		{
			name:                "users already exist",
			allowedTwitterUsers: []string{"user1", "user2"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				repository.On("Get", mock.Anything, "user1").
					Once().
					Return(
						&models.User{}, nil,
					)
				twitterClient.On("GetUser", mock.Anything, "user1").Return(&gotwitter.UserObj{
					CreatedAt: "2006-01-02T15:04:05Z",
				}, nil)
				repository.On("Get", mock.Anything, "user2").
					Once().
					Return(
						&models.User{}, nil,
					)
				twitterClient.On("GetUser", mock.Anything, "user2").Return(&gotwitter.UserObj{
					CreatedAt: "2006-01-02T15:04:05Z",
				}, nil)
			},
		},
		{
			name:                "creation ok",
			allowedTwitterUsers: []string{"user1", "user2"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				repository.On("Get", mock.Anything, "user1").
					Once().
					Return(
						nil, sql.ErrNoRows,
					)
				repository.On("Get", mock.Anything, "user2").
					Once().
					Return(
						nil, sql.ErrNoRows,
					)

				twitterClient.On("GetUser", mock.Anything, "user1").Once().Return(&gotwitter.UserObj{
					UserName:  "user1",
					CreatedAt: "2006-01-02T15:04:05Z",
				}, nil)
				twitterClient.On("GetUser", mock.Anything, "user2").Once().Return(&gotwitter.UserObj{
					UserName:  "user2",
					CreatedAt: "2006-01-02T15:04:05Z",
				}, nil)

				repository.On(
					"CreateUser",
					mock.Anything,
					mock.MatchedBy(func(req userrepository.CreateUserRequest) bool {
						return req.Username == "user1"
					})).
					Once().
					Return(nil, nil)

				repository.On(
					"CreateUser",
					mock.Anything,
					mock.MatchedBy(func(req userrepository.CreateUserRequest) bool {
						return req.Username == "user2"
					})).
					Once().
					Return(nil, nil)
			},
		},
		{
			name:                "err checking user from db",
			allowedTwitterUsers: []string{"user1"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				repository.On("Get", mock.Anything, "user1").
					Once().
					Return(
						nil, errors.New("unexpected err"),
					)
			},
			err: errors.New("unable to fetch user from db: unexpected err"),
		},
		{
			name:                "unknown twitter user",
			allowedTwitterUsers: []string{"user1"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				repository.On("Get", mock.Anything, "user1").
					Once().
					Return(
						nil, sql.ErrNoRows,
					)
				twitterClient.On("GetUser", mock.Anything, "user1").Once().Return(
					nil, twitter.UserNotFoundError{Username: "user1"},
				)
			},
			err: errors.New("@user1 does not exist on twitter"),
		},
		{
			name:                "unexpected twitter error",
			allowedTwitterUsers: []string{"user1"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				repository.On("Get", mock.Anything, "user1").
					Once().
					Return(
						nil, sql.ErrNoRows,
					)
				twitterClient.On("GetUser", mock.Anything, "user1").Once().Return(
					nil, errors.New("unexpected err"),
				)
			},
			err: errors.New("unexpected err"),
		},
		{
			name:                "error creating user",
			allowedTwitterUsers: []string{"user1"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				repository.On("Get", mock.Anything, "user1").
					Once().
					Return(
						nil, sql.ErrNoRows,
					)
				twitterClient.On("GetUser", mock.Anything, "user1").Once().Return(&gotwitter.UserObj{
					UserName:  "user1",
					CreatedAt: "2006-01-02T15:04:05Z",
				}, nil)
				repository.On(
					"CreateUser",
					mock.Anything,
					mock.MatchedBy(func(req userrepository.CreateUserRequest) bool {
						return req.Username == "user1"
					})).
					Once().
					Return(nil, sql.ErrConnDone)
			},
			err: errors.New("unable to create user: sql: connection is already closed"),
		},
	}

	log := mocks.NewNullLogger()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTwitter := mockstwitter.NewTwitterClient(t)
			fakeUserRepo := mocksuser.NewUserRepository(t)
			tt.mocks(fakeTwitter, fakeUserRepo)
			u := NewUserService(
				log,
				crypto.NewKeyManager(log, httpmock.NewClient(t)),
				fakeUserRepo,
				fakeTwitter,
			)
			err := u.BatchCreateUsers(context.TODO(), tt.allowedTwitterUsers)
			if tt.err != nil {
				require.EqualError(t, err, tt.err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_userService_BatchCreateUsersFromIDs(t *testing.T) {
	tests := []struct {
		name          string
		IDs           []string
		mocks         func(*mockstwitter.TwitterClient, *mocksuser.UserRepository)
		expectedUsers []*models.User
		err           string
	}{
		{
			name: "unable to retrieve users",
			IDs:  []string{"123"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				twitterClient.EXPECT().GetUserByIDs(mock.Anything, []string{"123"}).Return(nil, errors.New("client error"))
			},
			err: "unable to retrieve twitter users from id list: client error",
		},
		{
			name: "error get user from db",
			IDs:  []string{"123", "456"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				twitterClient.EXPECT().GetUserByIDs(mock.Anything, []string{"123", "456"}).Return([]*gotwitter.UserObj{
					{
						ID: "123",
					},
					{
						ID: "456",
					},
				}, nil)
				repository.EXPECT().Get(mock.Anything, "123").Return(nil, errors.New("db error"))
			},
			err: "unable to fetch user from db: db error",
		},
		{
			name: "all users already exists",
			IDs:  []string{"123", "456"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				twitterClient.EXPECT().GetUserByIDs(mock.Anything, []string{"123", "456"}).Return([]*gotwitter.UserObj{
					{
						ID: "123",
					},
					{
						ID: "456",
					},
				}, nil)
				repository.EXPECT().Get(mock.Anything, "123").Return(&models.User{}, nil)
				repository.EXPECT().Get(mock.Anything, "456").Return(&models.User{}, nil)
			},
		},
		{
			name: "error on date",
			IDs:  []string{"123", "456"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				twitterClient.EXPECT().GetUserByIDs(mock.Anything, []string{"123", "456"}).Return([]*gotwitter.UserObj{
					{
						ID: "123",
					},
					{
						ID:        "456",
						CreatedAt: "invalid_date",
					},
				}, nil)
				repository.EXPECT().Get(mock.Anything, "123").Return(&models.User{}, nil)
				repository.EXPECT().Get(mock.Anything, "456").Return(nil, nil)
			},
			err: `unable to create user: unable to parse user creation date from twitter: parsing time "invalid_date" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid_date" as "2006"`,
		},
		{
			name: "error on repo",
			IDs:  []string{"123", "456"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				fakeDateStr := "2006-01-02T15:04:05Z"
				fakeDate, _ := time.Parse(time.RFC3339, fakeDateStr)
				twitterClient.EXPECT().GetUserByIDs(mock.Anything, []string{"123", "456"}).Return([]*gotwitter.UserObj{
					{
						ID: "123",
					},
					{
						ID:        "456",
						UserName:  "foobar",
						CreatedAt: fakeDateStr,
					},
				}, nil)
				repository.EXPECT().Get(mock.Anything, "123").Return(&models.User{}, nil)
				repository.EXPECT().Get(mock.Anything, "456").Return(nil, nil)
				repository.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(req userrepository.CreateUserRequest) bool {
					return req.ID == "456" && req.CreatedAt == fakeDate && req.Username == "foobar"
				})).Return(nil, errors.New("repo error"))
			},
			err: "unable to create user: unable to create user: repo error",
		},
		{
			name: "create one user",
			IDs:  []string{"123", "456"},
			mocks: func(twitterClient *mockstwitter.TwitterClient, repository *mocksuser.UserRepository) {
				fakeDateStr := "2006-01-02T15:04:05Z"
				fakeDate, _ := time.Parse(time.RFC3339, fakeDateStr)
				twitterClient.EXPECT().GetUserByIDs(mock.Anything, []string{"123", "456"}).Return([]*gotwitter.UserObj{
					{
						ID: "123",
					},
					{
						ID:        "456",
						UserName:  "foobar",
						CreatedAt: fakeDateStr,
					},
				}, nil)
				repository.EXPECT().Get(mock.Anything, "123").Return(&models.User{ID: "123"}, nil)
				repository.EXPECT().Get(mock.Anything, "456").Return(nil, nil)
				repository.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(req userrepository.CreateUserRequest) bool {
					return req.ID == "456" && req.CreatedAt == fakeDate && req.Username == "foobar"
				})).Return(&models.User{ID: "456"}, nil)
			},
			expectedUsers: []*models.User{
				{ID: "123"},
				{ID: "456"},
			},
		},
	}

	log := mocks.NewNullLogger()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTwitter := mockstwitter.NewTwitterClient(t)
			fakeUserRepo := mocksuser.NewUserRepository(t)
			tt.mocks(fakeTwitter, fakeUserRepo)
			u := NewUserService(
				log,
				crypto.NewKeyManager(log, httpmock.NewClient(t)),
				fakeUserRepo,
				fakeTwitter,
			)
			users, err := u.BatchCreateUsersFromIDs(context.TODO(), tt.IDs)
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
			if tt.expectedUsers != nil {
				require.Equal(t, tt.expectedUsers, users)
			}
		})
	}
}
