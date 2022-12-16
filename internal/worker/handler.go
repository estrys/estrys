package worker

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"github.com/getsentry/sentry-go"
	"github.com/hibiken/asynq"

	"github.com/estrys/estrys/internal/observability"
)

type TraceableTask struct {
	TraceID string `json:"trace_id,omitempty"`
}
type TaskHandler func(ctx context.Context, t *asynq.Task) error

func traceIDFromHex(s string) sentry.TraceID {
	var id sentry.TraceID
	_, err := hex.Decode(id[:], []byte(s))
	if err != nil {
		panic(err)
	}
	return id
}

func TracingHandler(handler TaskHandler) TaskHandler {
	return func(ctx context.Context, task *asynq.Task) error {
		traceTask := TraceableTask{}
		var traceID *sentry.TraceID
		if err := json.Unmarshal(task.Payload(), &traceTask); err == nil {
			if traceTask.TraceID != "" {
				fromHex := traceIDFromHex(traceTask.TraceID)
				traceID = &fromHex
			}
		}
		tx := observability.StartTransaction(ctx, task.Type(), func(s *sentry.Span) {
			s.Op = "worker"
			if traceID != nil {
				s.TraceID = *traceID
			}
		})
		err := handler(tx.Context(), task) //nolint:contextcheck
		observability.FinishSpan(tx)
		return err
	}
}
