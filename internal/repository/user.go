package repository

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"time"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/estrys/estrys/internal/database"
	"github.com/estrys/estrys/internal/models"
)

type CreateUserRequest struct {
	Username   string
	CreatedAt  time.Time
	PrivateKey *rsa.PrivateKey
}

//go:generate mockery --name=UserRepository
type UserRepository interface {
	Get(context.Context, string) (*models.User, error)
	Follow(context.Context, *models.User, *models.Actor) error
	UnFollow(context.Context, *models.User, *models.Actor) error
	CreateUser(context.Context, CreateUserRequest) (*models.User, error)
}

type userRepo struct {
	db database.Database
}

func NewUserRepository(database database.Database) *userRepo {
	return &userRepo{db: database}
}

func (u *userRepo) CreateUser(ctx context.Context, input CreateUserRequest) (*models.User, error) {
	privKey := x509.MarshalPKCS1PrivateKey(input.PrivateKey)

	user := &models.User{
		Username:   input.Username,
		PrivateKey: privKey,
		CreatedAt:  input.CreatedAt,
	}
	err := user.Insert(ctx, getExecutor(ctx, u.db.DB()), boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "unable to save user")
	}

	return user, nil
}

func (u *userRepo) Follow(ctx context.Context, user *models.User, actor *models.Actor) error {
	err := user.AddActors(ctx, getExecutor(ctx, u.db.DB()), false, actor)
	if err != nil {
		return errors.Wrap(err, "unable to add actor to user")
	}
	return nil
}

func (u *userRepo) UnFollow(ctx context.Context, user *models.User, actor *models.Actor) error {
	err := user.RemoveActors(ctx, getExecutor(ctx, u.db.DB()), actor)
	if err != nil {
		return errors.Wrap(err, "unable to remove actor from user")
	}
	return nil
}

func (u *userRepo) Get(ctx context.Context, username string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.Username.EQ(username)).One(ctx, getExecutor(ctx, u.db.DB()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch user from database")
	}
	return user, nil
}
