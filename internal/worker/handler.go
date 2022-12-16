package worker

import (
	"context"
	"encoding/json"

	"github.com/getsentry/sentry-go"
	"github.com/hibiken/asynq"

	"github.com/estrys/estrys/internal/observability"
)

type TraceableTask struct {
	TraceID string `json:"trace_id"`
}
type TaskHandler func(ctx context.Context, t *asynq.Task) error

func TracingHandler(handler TaskHandler) TaskHandler {
	return func(ctx context.Context, task *asynq.Task) error {
		traceTask := TraceableTask{}
		_ = json.Unmarshal(task.Payload(), &traceTask)
		opts := []sentry.SpanOption{
			sentry.OpName("worker"),
			sentry.ContinueFromTrace(traceTask.TraceID),
		}
		tx := observability.StartTransaction(ctx, task.Type(), opts...)
		err := handler(tx.Context(), task) //nolint:contextcheck
		observability.FinishSpan(tx)
		return err
	}
}
