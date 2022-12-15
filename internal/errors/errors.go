package errors

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
)

type HandlerError struct {
	Cause       error
	UserMessage string
	HTTPCode    int
}

func (h HandlerError) Error() string {
	return h.UserMessage
}

type ErrorAwareHTTPHandler func(w http.ResponseWriter, req *http.Request) error

func HTTPErrorHandler(handler ErrorAwareHTTPHandler) func(w http.ResponseWriter, req *http.Request) {
	log := dic.GetService[logger.Logger]()
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				sentry.WithScope(func(scope *sentry.Scope) {
					scope.SetRequest(request)
					if err, isErr := recovered.(error); isErr {
						sentry.CaptureException(err)
						return
					} else {
						sentry.CurrentHub().Recover(recovered)
					}
				})
				log.Errorf("panic: %s %s", recovered, debug.Stack())
			}
		}()
		err := handler(responseWriter, request)
		var handlerError HandlerError
		if errors.As(err, &handlerError) {
			sentry.WithScope(func(scope *sentry.Scope) {
				scope.SetRequest(request)
				if handlerError.Cause != nil {
					sentry.CaptureException(handlerError.Cause)
				} else {
					sentry.CaptureException(handlerError)
				}
			})
			log.WithError(err).Error("Error handling http request")
			var respBody []byte
			if handlerError.UserMessage != "" {
				responseWriter.Header().Add("content-type", "application/json")
				respBody, _ = json.Marshal(map[string]string{
					"error": handlerError.UserMessage,
				})
			}
			responseWriter.WriteHeader(handlerError.HTTPCode)
			_, _ = responseWriter.Write(respBody)
		}
	}
}

func AsynqErrorHandler() asynq.ErrorHandlerFunc {
	log := dic.GetService[logger.Logger]()
	return func(ctx context.Context, task *asynq.Task, err error) {
		log.WithError(err).Error("Background task failed")
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("type", task.Type())
			scope.SetRequestBody(task.Payload())
			sentry.CaptureException(err)
		})
	}
}
