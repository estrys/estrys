package domain

import (
	"context"
	"crypto/x509"
	"database/sql"
	"net/url"
	"strings"
	"time"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/crypto"
	"github.com/estrys/estrys/internal/domain/domainmodels"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/repository"
	"github.com/estrys/estrys/internal/twitter"
)

//go:generate mockery --with-expecter --name=UserService
type UserService interface {
	GetFullUser(context.Context, string) (*domainmodels.User, error)
	BatchCreateUsersFromIDs(context.Context, []string) error
	BatchCreateUsers(ctx context.Context, allowedTwitterUsers []string) error
}

type userService struct {
	log           logger.Logger
	repo          repository.UserRepository
	keyManager    crypto.KeyManager
	twitterClient twitter.TwitterClient
}

func NewUserService(
	log logger.Logger,
	manager crypto.KeyManager,
	userRepo repository.UserRepository,
	client twitter.TwitterClient,
) *userService {
	return &userService{
		repo:          userRepo,
		log:           log,
		keyManager:    manager,
		twitterClient: client,
	}
}

// GetFullUser This method return a User from both database and twitter data.
func (u *userService) GetFullUser(ctx context.Context, username string) (*domainmodels.User, error) {
	user, err := u.repo.Get(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(ErrUserDoesNotExist)
		}
		return nil, err
	}

	// TODO Create a twitter user repo, move caching from the twitter client to the user repo
	twitterUser, err := u.twitterClient.GetUser(ctx, username)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch twitter user info")
	}

	profileImage, err := url.Parse(strings.ReplaceAll(
		twitterUser.ProfileImageURL,
		"_normal",
		"",
	))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse profile image url")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(user.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode private key")
	}

	domainUser := &domainmodels.User{
		Name:            twitterUser.Name,
		Username:        user.Username,
		Description:     twitterUser.Description,
		CreatedAt:       user.CreatedAt,
		ProfileImageURL: profileImage,
		Metrics: domainmodels.UserMetrics{
			Following: uint64(twitterUser.PublicMetrics.Following),
			Followers: uint64(twitterUser.PublicMetrics.Followers),
			Tweets:    uint64(twitterUser.PublicMetrics.Tweets),
		},
		PublicKey: privateKey.Public(),
	}

	return domainUser, nil
}

func (u *userService) BatchCreateUsersFromIDs(ctx context.Context, twitterIDs []string) error {
	twitterUsers, err := u.twitterClient.GetUserByIDs(ctx, twitterIDs)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve twitter users from id list")
	}
	for _, twitterUser := range twitterUsers {
		user, err := u.repo.Get(ctx, twitterUser.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(err, "unable to fetch user from db")
		}
		if user == nil {
			u.log.WithField("id", twitterUser.ID).Debug("user not found in database, creating it")
			err = u.createUserFromTwitter(ctx, twitterUser)
			if err != nil {
				return errors.Wrap(err, "unable to create user")
			}
			u.log.WithField("id", twitterUser.ID).Debug("new user created with new keypair")
		}
	}
	return nil
}

func (u *userService) BatchCreateUsers(ctx context.Context, usernames []string) error {
	for _, username := range usernames {
		user, err := u.repo.Get(ctx, username)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(err, "unable to fetch user from db")
		}
		twitterUser, err := u.twitterClient.GetUser(ctx, username)
		if err != nil && errors.Is(err, twitter.UserNotFoundError{Username: username}) {
			return &TwitterUserDoesNotExistError{username}
		}
		if err != nil {
			return err
		}
		if user == nil {
			u.log.WithField("username", username).Debug("user not found in database, creating it")
			err = u.createUserFromTwitter(ctx, twitterUser)
			if err != nil {
				return err
			}
			u.log.WithField("username", username).Debug("new user created with new keypair")
		}
	}
	return nil
}

func (u *userService) createUserFromTwitter(ctx context.Context, twitterUser *gotwitter.UserObj) error {
	privateKey, err := u.keyManager.GenerateKey()
	if err != nil {
		return errors.Wrap(err, "unable to generate private key for user")
	}
	createdAt, err := time.Parse(time.RFC3339, twitterUser.CreatedAt)
	if err != nil {
		return errors.Wrap(err, "unable to parse user creation date from twitter")
	}
	_, err = u.repo.CreateUser(ctx, repository.CreateUserRequest{
		Username:   twitterUser.UserName,
		ID:         twitterUser.ID,
		CreatedAt:  createdAt,
		PrivateKey: privateKey,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create user")
	}
	return nil
}
