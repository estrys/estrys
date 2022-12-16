package domain

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/estrys/estrys/internal/activitypub"
	activitypubclient "github.com/estrys/estrys/internal/activitypub/client"
	"github.com/estrys/estrys/internal/authorization"
	"github.com/estrys/estrys/internal/authorization/attributes"
	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/crypto"
	"github.com/estrys/estrys/internal/database"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/repository"
	"github.com/estrys/estrys/internal/worker/client"
	"github.com/estrys/estrys/internal/worker/tasks"
)

type UnsuportedUndoObjectError struct {
	VocabType vocab.Type
}

func (u *UnsuportedUndoObjectError) Error() string {
	actType := "unknown_type"
	if u.VocabType != nil {
		actType = u.VocabType.GetTypeName()
	}
	return fmt.Sprintf("do not support undo on '%s' activities", actType)
}

type ActorNotAllowedError struct {
	actor vocab.ActivityStreamsActorProperty
}

func (e *ActorNotAllowedError) Actor() string {
	data, _ := e.actor.Serialize()
	return fmt.Sprintf("%+v", data)
}

func (e *ActorNotAllowedError) Error() string {
	return "user not allowed on instance"
}

type InboxFollowRequest struct {
	Actor string
}

type InboxService interface {
	Follow(context.Context, vocab.ActivityStreamsFollow) error
	UnFollow(context.Context, vocab.ActivityStreamsUndo) error
}

type inboxService struct {
	log                  logger.Logger
	database             database.Database
	actorRepo            repository.ActorRepository
	userRepo             repository.UserRepository
	keyService           crypto.KeyManager
	activityPubClient    activitypubclient.ActivityPubClient
	vocabService         activitypub.VocabService
	worker               client.BackgroundWorkerClient
	authorizationChecker authorization.AuthorizationChecker
	config               config.Config
}

func NewInboxService(
	log logger.Logger,
	database database.Database,
	actorRepository repository.ActorRepository,
	userRepo repository.UserRepository,
	keyService crypto.KeyManager,
	activityPubClient activitypubclient.ActivityPubClient,
	vocabService activitypub.VocabService,
	worker client.BackgroundWorkerClient,
	authorizationChecker authorization.AuthorizationChecker,
	config config.Config,
) *inboxService {
	return &inboxService{
		log:                  log,
		database:             database,
		actorRepo:            actorRepository,
		userRepo:             userRepo,
		keyService:           keyService,
		activityPubClient:    activityPubClient,
		vocabService:         vocabService,
		worker:               worker,
		authorizationChecker: authorizationChecker,
		config:               config,
	}
}

func (a *inboxService) getUserFromURL(ctx context.Context, objectURL *url.URL) (*models.User, error) {
	splitObjectPath := strings.Split(objectURL.Path, "/")
	username := splitObjectPath[len(splitObjectPath)-1]

	user, err := a.userRepo.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (a *inboxService) Follow(ctx context.Context, follow vocab.ActivityStreamsFollow) error {
	objectURL, err := activitypub.GetObjectURL(follow)
	if err != nil {
		return errors.Wrap(err, "unable to get object url")
	}
	if objectURL.Host != a.config.Domain.Host {
		return errors.WithStack(ErrFollowMismatchDomain)
	}
	user, err := a.getUserFromURL(ctx, objectURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.WithStack(ErrUserDoesNotExist)
		}
		return err
	}

	if !a.authorizationChecker.IsGranted(follow.GetActivityStreamsActor(), attributes.CanFollow) {
		rejectFollowTask, err := tasks.NewRejectFollowTask(user.Username, follow)
		if err != nil {
			return errors.Wrap(err, "unable to create reject follow task")
		}
		_, err = a.worker.Enqueue(rejectFollowTask)
		if err != nil {
			return errors.Wrap(err, "unable to schedule reject follow task")
		}

		return errors.WithStack(&ActorNotAllowedError{actor: follow.GetActivityStreamsActor()})
	}

	actorURL, err := activitypub.GetActorURL(follow)
	if err != nil {
		return errors.Wrap(err, "unable to get actor url")
	}

	actor, err := a.createActorIfNoExist(ctx, actorURL)
	if err != nil {
		return errors.Wrap(err, "unable to create actor")
	}
	err = a.userRepo.Follow(ctx, user, actor)
	if err != nil {
		return errors.Wrap(err, "unable to follow actor")
	}

	acceptFollowTask, err := tasks.NewAcceptFollowTask(user.Username, follow)
	if err != nil {
		return errors.Wrap(err, "unable to create accept follow task")
	}
	_, err = a.worker.Enqueue(acceptFollowTask)
	if err != nil {
		return errors.Wrap(err, "unable to schedule accept follow task")
	}

	a.log.WithFields(logrus.Fields{
		"actor":  actorURL.String(),
		"object": objectURL.String(),
	}).Info("successfully handled follow request")

	return nil
}

func (a *inboxService) createActorIfNoExist(ctx context.Context, url *url.URL) (*models.Actor, error) {
	actor, err := a.actorRepo.Get(ctx, url)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to fetch actor from db")
	}

	if actor == nil {
		pubKey, err := a.keyService.FetchKey(ctx, url.String())
		if err != nil {
			return nil, errors.Wrap(err, "unable to fetch actor key")
		}
		actor, err = a.actorRepo.Create(ctx, repository.CreateActorRequest{
			URL:       url,
			PublicKey: pubKey,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to create actor")
		}
		a.log.WithField("actor", actor.URL).Debug("new actor created")
	}

	return actor, nil
}

func (a *inboxService) UnFollow(ctx context.Context, act vocab.ActivityStreamsUndo) error {
	actorURL, err := activitypub.GetActorURL(act)
	if err != nil {
		return errors.Wrap(err, "unable to get actor url")
	}

	object := act.GetActivityStreamsObject()
	if object.Len() == 0 {
		return errors.Wrap(err, "no undo object specified")
	}
	if object.Len() != 1 {
		return errors.Wrap(err, "unfollowing multiples object not supported yet")
	}
	obj := object.Begin()
	if !obj.IsActivityStreamsFollow() {
		return errors.WithStack(&UnsuportedUndoObjectError{obj.GetType()})
	}

	followActivity := obj.GetActivityStreamsFollow()
	userURL, err := activitypub.GetObjectURL(followActivity)
	if err != nil {
		return errors.Wrap(err, "unable to get object URL")
	}
	user, err := a.getUserFromURL(ctx, userURL)
	if err != nil {
		return err
	}

	actor, err := a.actorRepo.Get(ctx, actorURL)
	if err != nil {
		return errors.Wrap(err, "unable to fetch actor")
	}

	err = a.userRepo.UnFollow(ctx, user, actor)
	if err != nil {
		return errors.Wrap(err, "unable to unfollow user")
	}

	a.log.WithFields(logrus.Fields{
		"actor":  actorURL.String(),
		"object": userURL.String(),
	}).Info("successfully handled unfollow request")

	return nil
}
