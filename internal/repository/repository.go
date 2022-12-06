package repository

import (
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/estrys/estrys/internal/database"
)

func getExecutor(ctx context.Context, db *sql.DB) boil.ContextExecutor {
	tx, ok := ctx.Value(database.ContextTransaction).(*sql.Tx)
	if ok {
		return tx
	}
	return db
}
