package domain

import (
	"context"
	"crypto/x509"
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/crypto"
	"github.com/estrys/estrys/internal/domain/domainmodels"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/repository"
	"github.com/estrys/estrys/internal/twitter"
)

type CreateUserRequest struct {
	Username  string
	CreatedAt time.Time
}

type UserService interface {
	GetFullUser(context.Context, string) (*domainmodels.User, error)
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
			return nil, ErrUserDoesNotExist
		}
		return nil, err
	}

	twitterUser, err := u.twitterClient.GetUser(ctx, username)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch twitter user info")
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
		ProfileImageURL: twitterUser.ProfileImageURL,
		Metrics: domainmodels.UserMetrics{
			Following: twitterUser.Following,
			Followers: twitterUser.Followers,
			Tweets:    twitterUser.Tweets,
		},
		PublicKey: privateKey.Public(),
	}

	return domainUser, nil
}

func (u *userService) CreateUser(ctx context.Context, request CreateUserRequest) (*models.User, error) {
	user, err := u.repo.Get(ctx, request.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to fetch user from db")
	}
	if user == nil {
		u.log.WithField("username", request.Username).Debug("user not found in database, creating it")
		privateKey, err := u.keyManager.GenerateKey()
		if err != nil {
			return nil, errors.Wrap(err, "unable to generate private key for user")
		}
		user, err = u.repo.CreateUser(ctx, repository.CreateUserRequest{
			Username:   request.Username,
			CreatedAt:  request.CreatedAt,
			PrivateKey: privateKey,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to create user")
		}
		u.log.WithField("username", request.Username).Debug("new user created with new keypair")
	}

	return user, nil
}
