package repository

import (
	"context"
	"crypto"
	"crypto/x509"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/estrys/estrys/internal/database"
	"github.com/estrys/estrys/internal/models"
)

type CreateActorRequest struct {
	URL       *url.URL
	PublicKey crypto.PublicKey
}

//go:generate mockery --name ActorRepository
type ActorRepository interface {
	Get(ctx context.Context, url *url.URL) (*models.Actor, error)
	Create(context.Context, CreateActorRequest) (*models.Actor, error)
}

type actorRepo struct {
	db database.Database
}

func NewActorRepository(database database.Database) *actorRepo {
	return &actorRepo{db: database}
}

func (u *actorRepo) Create(ctx context.Context, input CreateActorRequest) (*models.Actor, error) {
	pubKey, err := x509.MarshalPKIXPublicKey(input.PublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode actor public key")
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate a valid UUIDv4 for actor")
	}
	actor := &models.Actor{
		ID:        id.String(),
		URL:       input.URL.String(),
		PublicKey: pubKey,
	}
	err = actor.Insert(ctx, u.db.DB(), boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "unable to save actor in db")
	}

	return actor, nil
}

func (u *actorRepo) Get(ctx context.Context, url *url.URL) (*models.Actor, error) {
	actor, err := models.Actors(models.ActorWhere.URL.EQ(url.String())).One(ctx, u.db.DB())
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch actor from db")
	}
	return actor, nil
}
