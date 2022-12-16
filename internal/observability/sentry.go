package observability

import (
	"context"

	"github.com/getsentry/sentry-go"
)

func StartTransaction(ctx context.Context, name string, options ...sentry.SpanOption) *sentry.Span {
	tx := sentry.StartTransaction(ctx, name, options...)
	tx.Op = name
	return tx
}

func GetTraceIDFromContext(ctx context.Context) string {
	tx := sentry.TransactionFromContext(ctx)
	if tx == nil {
		return ""
	}
	return tx.ToSentryTrace()
}

func StartSpan(ctx context.Context, name string, data map[string]any) *sentry.Span {
	transaction := sentry.TransactionFromContext(ctx)
	if transaction == nil {
		return nil
	}
	return transaction.StartChild(name, func(s *sentry.Span) {
		s.Data = data
	})
}

func FinishSpan(span *sentry.Span) {
	if span != nil {
		span.Finish()
	}
}
