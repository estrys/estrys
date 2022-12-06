package mocks

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

type nullLogger struct {
	logrus.FieldLogger
}

func NewNullLogger() *nullLogger {
	logrusNullLogger, _ := test.NewNullLogger()
	return &nullLogger{logrusNullLogger}
}

func (l *nullLogger) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return nil
}

func (l *nullLogger) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {}
