package repository

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/estrys/estrys/internal/database"
	"github.com/estrys/estrys/internal/models"
)

type CreateUserRequest struct {
	Username   string
	ID         string
	CreatedAt  time.Time
	PrivateKey *rsa.PrivateKey
}

//go:generate mockery --with-expecter --name=UserRepository
type UserRepository interface {
	Get(context.Context, string) (*models.User, error)
	GetFollowers(context.Context, *models.User) (models.ActorSlice, error)
	Follow(context.Context, *models.User, *models.Actor) error
	UnFollow(context.Context, *models.User, *models.Actor) error
	CreateUser(context.Context, CreateUserRequest) (*models.User, error)
	GetWithFollowers(ctx context.Context) (models.UserSlice, error)
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
		ID:         input.ID,
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

func (u *userRepo) Get(ctx context.Context, usernameOrID string) (*models.User, error) {
	var err error
	var user *models.User
	if _, err = strconv.ParseInt(usernameOrID, 10, 64); err == nil {
		user, err = models.Users(models.UserWhere.ID.EQ(usernameOrID)).
			One(ctx, getExecutor(ctx, u.db.DB()))
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, "unable to fetch user from database")
		}
	}
	if user == nil {
		user, err = models.Users(models.UserWhere.Username.EQ(usernameOrID)).
			One(ctx, getExecutor(ctx, u.db.DB()))
		if err != nil {
			return nil, errors.Wrap(err, "unable to fetch user from database")
		}
	}
	return user, nil
}

func (u *userRepo) GetWithFollowers(ctx context.Context) (models.UserSlice, error) {
	mods := []qm.QueryMod{
		qm.InnerJoin(fmt.Sprintf("%[1]s on %[1]s.user = %s",
			models.TableNames.Followers,
			models.UserTableColumns.Username,
		)),
	}
	return models.Users(mods...).All(ctx, getExecutor(ctx, u.db.DB()))
}

func (u *userRepo) GetFollowers(ctx context.Context, user *models.User) (models.ActorSlice, error) {
	return user.Actors().All(ctx, getExecutor(ctx, u.db.DB()))
}
