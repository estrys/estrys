package domain

import (
	"context"
	"database/sql"
	"testing"
	"time"

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
				repository.On("Get", mock.Anything, "user2").
					Once().
					Return(
						&models.User{}, nil,
					)
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

				twitterClient.On("GetUser", mock.Anything, "user1").Once().Return(&twitter.User{CreatedAt: time.Now()}, nil)
				twitterClient.On("GetUser", mock.Anything, "user2").Once().Return(&twitter.User{CreatedAt: time.Now()}, nil)

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
			err: errors.New("unable to batch create user: unable to fetch user from db: unexpected err"),
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
					nil, twitter.UsernameNotFoundError{Username: "user1"},
				)
			},
			err: errors.New("unable to batch create user: @user1 does not exist on twitter"),
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
			err: errors.New("unable to batch create user: unexpected err"),
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
				twitterClient.On("GetUser", mock.Anything, "user1").Once().Return(&twitter.User{CreatedAt: time.Now()}, nil)
				repository.On(
					"CreateUser",
					mock.Anything,
					mock.MatchedBy(func(req userrepository.CreateUserRequest) bool {
						return req.Username == "user1"
					})).
					Once().
					Return(nil, sql.ErrConnDone)
			},
			err: errors.New("unable to batch create user: unable to create user: sql: connection is already closed"),
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
