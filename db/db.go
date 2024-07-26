package db

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Postgresql driver
	"github.com/jmoiron/sqlx"
)

const (
	DriverPostgreSQL = "pgx"
)

func NewPostgres(ctx context.Context, dsn string) (*sqlx.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("DB not initialized (driver is not specified)")
	}

	db, err := sqlx.ConnectContext(ctx, DriverPostgreSQL, dsn)
	if err != nil {
		return nil, fmt.Errorf("can't open %s: %w", DriverPostgreSQL, err)
	}

	go regularPing(ctx, "postgres", db) // pgbounce needed

	return db, db.PingContext(ctx)
}

func regularPing(ctx context.Context, dbtype string, db *sqlx.DB) {
	tck := time.NewTicker(time.Minute)
	defer tck.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tck.C:
			if err := db.PingContext(ctx); err != nil {
				slog.Error("can't ping database", "type", dbtype, "err", err)
			}
		}
	}
}
