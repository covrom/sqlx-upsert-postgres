# sqlx-upsert-postgres
Insert or update (aka "upsert", "replace") structure into postgres table using sqlx prepared statement.

## Basic usage

```go
import (
    upsert "github.com/covrom/sqlx-upsert-postgres"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
)

type PgComment struct {
	ID        uuid.UUID    `db:"id" pk:"true"` // primary key for conflict resolving
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`

	Description     string  `db:"description"`
	ComputedColumn  float64 `store:"-"` // skipped
}

el := PgComment{
    ID:              uuid.New(),
    CreatedAt:       time.Now(),
    UpdatedAt:       time.Now(),
    Description:     "this is a comment",
    ComputedColumn:  5.54,   // skipped in tag
}

st, err := upsert.PrepareNamedQuery[PgComment](ctx, pgdb, "comments", el)
if err != nil {
    return err
}
defer st.Close()

_, err = st.ExecContext(ctx, el)
if err != nil {
    return err
}
```

## Full example

[Example](upsert_test.go)