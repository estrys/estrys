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

type errorWithStack interface {
	Error() string
	StackTrace() errors.StackTrace
}

type handlerError struct {
	Cause             error
	UserMessage       string
	HTTPCode          int
	SkipErrorCapture  bool
	stack             errors.StackTrace
	additionalContext map[string]any
}

func (h handlerError) Error() string {
	message := h.UserMessage
	if h.Cause != nil {
		if message != "" {
			message += ": "
		}
		message += h.Cause.Error()
	}
	return message
}

func (h handlerError) StackTrace() errors.StackTrace {
	return h.stack
}

func (h handlerError) SkipCapture() handlerError {
	h.SkipErrorCapture = true
	return h
}

func (h handlerError) WithHTTPCode(code int) handlerError {
	h.HTTPCode = code
	return h
}

func (h handlerError) WithUserMessage(message string) handlerError {
	h.UserMessage = message
	return h
}

func (h handlerError) WithContext(key string, value any) handlerError {
	h.additionalContext[key] = value
	return h
}

func (h handlerError) GetContext() map[string]any {
	return h.additionalContext
}

func New(userMessage string, httpCode int) handlerError {
	return Wrap(errors.New(userMessage), httpCode)
}

func Wrap(err error, httpCode int) handlerError {
	newErr := handlerError{
		Cause:             err,
		HTTPCode:          httpCode,
		additionalContext: make(map[string]any, 0),
	}

	//nolint:errorlint
	if stackErr, ok := err.(errorWithStack); ok {
		newErr.stack = stackErr.StackTrace()
	}

	return newErr
}

type ErrorAwareHTTPHandler func(w http.ResponseWriter, req *http.Request) error

func HTTPErrorHandler(handler ErrorAwareHTTPHandler) func(w http.ResponseWriter, req *http.Request) {
	log := dic.GetService[logger.Logger]()
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		// Fetch the sentry hub from request context
		// It can be null in testing context
		hub := sentry.GetHubFromContext(request.Context())
		if hub != nil {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetRequest(request)
			})
		}
		if hub != nil {
			defer func() {
				if recovered := recover(); recovered != nil {
					if err, isErr := recovered.(error); isErr {
						hub.CaptureException(err)
						return
					} else {
						hub.Recover(recovered)
					}
					log.Errorf("panic: %s %s", recovered, debug.Stack())
				}
			}()
		}
		err := handler(responseWriter, request)
		var handlerError handlerError
		if errors.As(err, &handlerError) {
			log.WithError(err).Error("Error handling http request")
			if !handlerError.SkipErrorCapture && hub != nil {
				if additionalContext := handlerError.GetContext(); len(additionalContext) > 0 {
					hub.ConfigureScope(func(scope *sentry.Scope) {
						scope.SetContext("error", additionalContext)
					})
				}
				hub.CaptureException(handlerError)
			}
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
			scope.SetContext("task", map[string]interface{}{
				"task_type": task.Type(),
				"payload":   string(task.Payload()),
			})
			scope.SetTag("task_type", task.Type())
			sentry.CaptureException(err)
		})
	}
}
