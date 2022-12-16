package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/cache"
	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/database"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/dic/container"
	"github.com/estrys/estrys/internal/logger"
)

func Bootstrap() (context.Context, context.CancelFunc, error) {
	err := container.BuildContainer()
	if err != nil {
		return nil, nil, err //nolint:wrapcheck
	}

	globalContext, cancelFunc := context.WithCancel(context.Background())

	log := dic.GetService[logger.Logger]()
	conf := dic.GetService[config.Config]()

	log.WithField("pid", os.Getpid()).Debug("app starting")

	if conf.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              conf.SentryDSN,
			TracesSampleRate: 1.0,
			AttachStacktrace: true,
			EnableTracing:    true,
		})
		if err != nil {
			cancelFunc()
			return nil, nil, errors.Wrap(err, "unable to init sentry")
		}
		log.Info("sentry initialized, errors will be reported")
	}

	if err = dic.GetService[database.Database]().Connect(); err != nil {
		cancelFunc()
		return nil, nil, errors.Wrap(err, "unable to connect to postgres")
	}
	log.WithField("db", conf.DBURL.Redacted()).Info("Connected to postgres")

	if err := dic.GetService[cache.RedisClient]().Ping(globalContext); err != nil {
		cancelFunc()
		return nil, nil, errors.Wrap(err, "unable to connect to redis")
	}
	log.WithField("addr", conf.RedisAddress).Info("Connected to redis")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	go func() {
		s := <-sigCh
		log.WithField("signal", s.String()).Info("signal received, stopping")
		cancelFunc()
	}()

	return globalContext, cancelFunc, nil
}
