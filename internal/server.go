package internal

import (
	"context"
	"net/http"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
)

type Config struct {
	Address string
}

func StartServer(ctx context.Context, cfg Config) error {
	muxRouter := dic.GetService[*mux.Router]()
	log := dic.GetService[logger.Logger]()
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	handler := sentryHandler.HandleFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		fields := logrus.Fields{}
		for k, v := range vars {
			fields[k] = v
		}
		fields["querystring"] = request.URL.RawQuery
		log.WithFields(fields).Debugf("%s %s", request.Method, request.URL.Path)
		muxRouter.ServeHTTP(responseWriter, request)
	})

	srv := &http.Server{
		Handler:      handler,
		Addr:         cfg.Address,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	log.WithField("address", cfg.Address).Info("Starting http server")

	go func() {
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Error("server failed")
			return
		}
		log.Info("server stopped")
	}()

	<-ctx.Done()
	err := srv.Shutdown(ctx)
	if err != nil {
		return errors.Wrap(err, "error while sending shutdown signal to http server")
	}

	return nil
}
