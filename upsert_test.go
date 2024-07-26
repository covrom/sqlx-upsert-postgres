package upsert_test

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"time"

	upsert "github.com/covrom/sqlx-upsert-postgres"
	"github.com/covrom/sqlx-upsert-postgres/db"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var _ upsert.Upserter = PgComment{}

type PgComment struct {
	ID        uuid.UUID    `db:"id" pk:"true"` // primary key for conflict resolving
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`

	Description     string  `db:"description"`
	ComputedColumn  float64 // needs to skip
	ComputedColumn2 float64 `store:"-"` // needs to skip
	ComputedColumn3 float64 // needs to skip
}

// returns database table name
func (v PgComment) Table() string {
	return "comment"
}

// returns database name of primary key columnts
func (v PgComment) UpsertPrimaryKeyColumns() []string {
	return nil // use "pk" tag
}

// returns database name of skipped columns
func (v PgComment) UpsertSkipColumns() []string {
	return []string{sqlx.NameMapper("ComputedColumn3")}
}

func ExampleUpsert() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(l)

	pgdb, err := db.NewPostgres(ctx, os.Getenv("PG_DATABASE_DSN"))
	if err != nil {
		slog.Error("db.NewPostgres error", "err", err)
		return
	}
	defer pgdb.Close()

	el := PgComment{
		ID:              uuid.New(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Description:     "this is a comment",
		ComputedColumn:  5.54,   // skipped in prepare
		ComputedColumn2: 6.234,  // skipped in tag
		ComputedColumn3: 83.234, // skipped in UpsertSkipColumns
	}

	st, err := upsert.PrepareNamedQuery[PgComment](ctx, pgdb,
		el.Table(), el,
		// additionally skipping columns
		sqlx.NameMapper("ComputedColumn"))
	if err != nil {
		slog.Error("upsert.PrepareNamedQuery error", "err", err)
		return
	}
	defer st.Close()

	_, err = st.ExecContext(ctx, el)
	if err != nil {
		slog.Error("st.ExecContext error", "err", err)
		return
	}
}
