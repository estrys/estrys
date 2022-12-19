package worker

import (
	"context"
	"encoding/json"

	"github.com/getsentry/sentry-go"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/observability"
	taskerrors "github.com/estrys/estrys/internal/worker/errors"
	"github.com/estrys/estrys/internal/worker/tasks"
)

type TraceableTask struct {
	TraceID string `json:"trace_id"`
}
type TaskHandler func(ctx context.Context, t *asynq.Task) error

func ErrorHandler(handler TaskHandler) TaskHandler {
	log := dic.GetService[logger.Logger]()
	return func(ctx context.Context, task *asynq.Task) error {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetContext("task", map[string]interface{}{
				"task_type": task.Type(),
				"payload":   string(task.Payload()),
			})
			scope.SetTag("task_type", task.Type())
		})
		err := handler(ctx, task)
		if err != nil {
			log.WithError(err).Errorf("Background task '%s' failed", task.Type())
			var taskErr taskerrors.TaskError
			if errors.As(err, &taskErr) {
				sentry.CaptureException(taskErr.Cause())
				if taskErr.SkipRetry {
					return asynq.SkipRetry
				}
			}
			sentry.CaptureException(err)
			return err
		}
		return nil
	}
}

func TracingHandler(handler TaskHandler) TaskHandler {
	return func(ctx context.Context, task *asynq.Task) error {
		traceTask := TraceableTask{}
		_ = json.Unmarshal(task.Payload(), &traceTask)
		opts := []sentry.SpanOption{
			sentry.OpName("worker"),
			sentry.ContinueFromTrace(traceTask.TraceID),
		}
		if task.Type() == tasks.TypeSendTweet {
			opts = append(opts, func(s *sentry.Span) {
				s.Sampled = sentry.SampledTrue
			})
		}
		tx := observability.StartTransaction(ctx, task.Type(), opts...)
		err := handler(tx.Context(), task) //nolint:contextcheck
		observability.FinishSpan(tx)
		return err
	}
}
