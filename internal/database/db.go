package database

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/logger"
)

type Database interface {
	Connect() error
	DB() *sql.DB
	StartTransaction(context.Context) (*sql.Tx, context.Context, error)
}

type contextKey int

const (
	ContextTransaction contextKey = iota
)

type postgres struct {
	url *url.URL
	log logger.Logger
	db  *sql.DB
}

func NewPostgres(url *url.URL, log logger.Logger) *postgres {
	return &postgres{
		url: url,
		log: log,
	}
}

func (p *postgres) Connect() error {
	connConfig, _ := pgx.ParseConfig(p.url.String())
	connConfig.Tracer = p.log
	connStr := stdlib.RegisterConnConfig(connConfig)
	database, err := sql.Open("pgx", connStr)
	if err != nil {
		return errors.Wrap(err, "unable to open a connection to the database")
	}
	err = database.Ping()
	if err != nil {
		return errors.Wrap(err, "unable to ping database")
	}

	p.db = database
	return nil
}

func (p *postgres) DB() *sql.DB {
	return p.db
}

func (p *postgres) StartTransaction(ctx context.Context) (*sql.Tx, context.Context, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err //nolint:wrapcheck
	}
	return tx, context.WithValue(ctx, ContextTransaction, tx), nil
}
