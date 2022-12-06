package main

import (
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/cmd"
	"github.com/estrys/estrys/internal"
	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/worker"
)

func main() {
	appContext, _, err := cmd.Bootstrap()
	log := dic.GetService[logger.Logger]()
	if err != nil {
		if log != nil {
			log.WithError(err).Error("unable to start application")
			os.Exit(1)
		}
		panic(err)
	}

	conf := dic.GetService[config.Config]()

	if !conf.DisableEmbedWorker {
		go func() {
			err := worker.StartBroker(appContext)
			if err != nil {
				log.WithError(err).Error("worker failed")
			}
		}()
	}

	err = internal.StartServer(appContext, internal.Config{Address: conf.Address})
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.WithError(err).Error("server failed")
		os.Exit(1)
	}
	log.Info("http server stopped")
	os.Exit(0)
}
