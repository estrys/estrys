package logger

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"

	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/observability"
)

type contextKey int

const (
	tracelogQueryCtxKey contextKey = iota
)

type traceQueryData struct {
	startTime time.Time
	sql       string
	args      []any
	span      *sentry.Span
}

type Logger interface {
	logrus.FieldLogger
	pgx.QueryTracer
}

type logger struct {
	logrus.FieldLogger
}

func CreateLogger(config *config.Config) *logger {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		DisableQuote: true,
	}
	log.SetLevel(config.LogLevel)
	return &logger{
		FieldLogger: log,
	}
}

func (l *logger) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	args := data.Args
	if len(args) >= 2 {
		args = args[1:]
	}
	return context.WithValue(ctx, tracelogQueryCtxKey, &traceQueryData{
		startTime: time.Now(),
		sql:       data.SQL,
		args:      args,
		span:      observability.StartSpan(ctx, "db", map[string]any{"db.system": "postgres", "db.query": data.SQL}),
	})
}

func (l *logger) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	queryData := ctx.Value(tracelogQueryCtxKey).(*traceQueryData) //nolint:forcetypeassert
	observability.FinishSpan(queryData.span)
	endTime := time.Now()
	interval := endTime.Sub(queryData.startTime)
	l.WithFields(logrus.Fields{
		"duration": interval.String(),
		"query":    queryData.sql,
		"args":     queryData.args,
	}).Trace("database query")
}
