package main

import (
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/cmd"
	"github.com/estrys/estrys/internal"
	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/domain"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/twitter"
	"github.com/estrys/estrys/internal/worker"
	"github.com/estrys/estrys/migrations"
)

func main() {
	appContext, _, err := cmd.Bootstrap()
	if err != nil {
		panic(err)
	}
	log := dic.GetService[logger.Logger]()
	conf := dic.GetService[config.Config]()

	if conf.RunMigrations {
		d, err := iofs.New(migrations.FS, ".")
		if err != nil {
			log.WithError(err).Fatal("Could not read database migration files")
		}
		migration, err := migrate.NewWithSourceInstance("iofs", d, conf.DBURL.String())
		if err != nil {
			log.WithError(err).Fatal("Failed to init migration")
		}

		log.Info("Running database migrations ...")
		err = migration.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.WithError(err).Fatal("Could not run migrations")
		}
	}

	if conf.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              conf.SentryDSN,
			TracesSampleRate: 1.0,
			AttachStacktrace: true,
		})
		if err != nil {
			log.WithError(err).Error("unable to init sentry")
			os.Exit(1)
		}
		log.Info("sentry initialized, errors will be reported")
	}

	if !conf.DisableEmbedWorker {
		go func() {
			err := worker.StartBroker(appContext)
			if err != nil {
				log.WithError(err).Error("worker failed")
			}
		}()
	}

	userService := dic.GetService[domain.UserService]()
	err = userService.BatchCreateUsers(appContext, conf.TwitterAllowedUsers)
	if err != nil {
		log.WithError(err).Error("User initialization failed")
		os.Exit(1)
	}

	twp := dic.GetService[twitter.TwitterPoller]()
	go func() {
		err := twp.Start(appContext)
		if err != nil {
			log.Fatalf("Could not start poller: %s", err)
		}
	}()

	err = internal.StartServer(appContext, internal.Config{Address: conf.Address})
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.WithError(err).Error("server failed")
		os.Exit(1)
	}
	log.Info("http server stopped")
	os.Exit(0)
}
