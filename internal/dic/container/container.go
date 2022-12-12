package container

import (
	"net/http"

	gotwitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/go-redis/redis/v9"
	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/activitypub"
	activitypubclient "github.com/estrys/estrys/internal/activitypub/client"
	"github.com/estrys/estrys/internal/authorization"
	"github.com/estrys/estrys/internal/authorization/voter"
	"github.com/estrys/estrys/internal/cache"
	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/crypto"
	"github.com/estrys/estrys/internal/database"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/domain"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/repository"
	"github.com/estrys/estrys/internal/router"
	"github.com/estrys/estrys/internal/router/urlgenerator"
	"github.com/estrys/estrys/internal/twitter"
	"github.com/estrys/estrys/internal/worker/client"
)

func BuildContainer() error {
	loader := config.NewLoader()
	err := loader.Load()
	if err != nil {
		return errors.Wrap(err, "unable to load config")
	}
	conf := loader.Get()
	_ = dic.Register[config.Config](conf)
	_ = dic.Register[logger.Logger](logger.CreateLogger(&conf))

	_ = dic.Register[*mux.Router](router.GetRouter())
	_ = dic.Register[urlgenerator.URLGenerator](urlgenerator.NewURLGenerator(conf,
		dic.GetService[*mux.Router](),
	))
	_ = dic.Register[activitypub.VocabService](activitypub.NewActivityPubVocabService(
		dic.GetService[urlgenerator.URLGenerator](),
	))

	_ = dic.Register[database.Database](database.NewPostgres(
		conf.DBURL,
		dic.GetService[logger.Logger](),
	))
	redisClient := cache.NewRedisClient(&redis.Options{Addr: conf.RedisAddress})
	_ = dic.Register[cache.RedisClient](*redisClient)
	_ = dic.Register[cache.Cache[twitter.User]](cache.CreateRedisCache[twitter.User](
		redisClient,
		cache.OptionDefaultTTL(conf.TwitterUserCacheTimeout),
	))
	_ = dic.Register[client.BackgroundWorkerClient](client.NewBackgroundWorkerClient(
		asynq.NewClient(asynq.RedisClientOpt{Addr: conf.RedisAddress}),
	))

	_ = dic.Register[twitter.Backend](&gotwitter.Client{
		Authorizer: twitter.Authorizer{
			Token: conf.Token,
		},
		Client: &http.Client{
			Transport: logger.GetResponseLogger(dic.GetService[logger.Logger]()),
		},
		Host: "https://api.twitter.com",
	})
	_ = dic.Register[twitter.TwitterClient](twitter.NewClient(
		dic.GetService[logger.Logger](),
		dic.GetService[cache.Cache[twitter.User]](),
		dic.GetService[twitter.Backend](),
	))
	activityPubClient, err := activitypubclient.NewActivityPubClient(
		&http.Client{},
		dic.GetService[logger.Logger](),
		dic.GetService[urlgenerator.URLGenerator](),
	)
	if err != nil {
		return errors.Wrap(err, "unable to create the activitypub client")
	}
	_ = dic.Register[activitypubclient.ActivityPubClient](activityPubClient)

	_ = dic.Register[repository.UserRepository](repository.NewUserRepository(
		dic.GetService[database.Database](),
	))
	_ = dic.Register[repository.ActorRepository](repository.NewActorRepository(
		dic.GetService[database.Database](),
	))
	_ = dic.Register[authorization.AuthorizationChecker](authorization.NewVoterAuthorizationChecker([]voter.Voter{
		voter.NewActivityVoter(conf.AllowedUsers),
	}))
	_ = dic.Register[crypto.KeyManager](crypto.NewKeyManager(
		dic.GetService[logger.Logger](),
		&http.Client{},
	))
	_ = dic.Register[domain.UserService](domain.NewUserService(
		dic.GetService[logger.Logger](),
		dic.GetService[crypto.KeyManager](),
		dic.GetService[repository.UserRepository](),
		dic.GetService[twitter.TwitterClient](),
	))
	_ = dic.Register[domain.InboxService](domain.NewInboxService(
		dic.GetService[logger.Logger](),
		dic.GetService[database.Database](),
		dic.GetService[repository.ActorRepository](),
		dic.GetService[repository.UserRepository](),
		dic.GetService[crypto.KeyManager](),
		dic.GetService[activitypubclient.ActivityPubClient](),
		dic.GetService[activitypub.VocabService](),
		dic.GetService[client.BackgroundWorkerClient](),
		dic.GetService[authorization.AuthorizationChecker](),
		conf,
	))

	_ = dic.Register[twitter.TwitterPoller](twitter.NewPoller(
		dic.GetService[logger.Logger](),
		dic.GetService[twitter.TwitterClient](),
		dic.GetService[repository.UserRepository](),
		dic.GetService[client.BackgroundWorkerClient](),
	))

	return nil
}
