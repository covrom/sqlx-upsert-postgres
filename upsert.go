package upsert


import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func PrepareNamedQuery[T any](ctx context.Context, db *sqlx.DB, tableName string, val T, skipCols ...string) (*sqlx.NamedStmt, error) {
	dcols, err := StructColumns(val)
	if err != nil {
		return nil, err
	}
	cols := columnsWithout(dcols.DBs(), skipCols...)
	pkcols := columnsWithout(dcols.PKs(), skipCols...)
	idcs := strings.Join(pkcols, ",")

	colsnoids := columnsWithout(cols, append(pkcols, skipCols...)...)

	var suffix string
	suffixVals := []any{
		strings.Join(colsnoids, ","),
		strings.Join(excluded(colsnoids), ","),
	}
	switch {
	case len(colsnoids) == 0:
		suffix = "NOTHING"
		suffixVals = nil
	case len(colsnoids) == 1:
		suffix = "UPDATE SET %s=%s"
	case len(colsnoids) > 1:
		suffix = "UPDATE SET (%s)=(%s)"
	}

	q := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`,
		tableName,
		strings.Join(cols, ","),
		replacers(cols),
	)

	if len(idcs) > 0 {
		q = fmt.Sprintf(`%s ON CONFLICT(%s) DO %s`,
			q,
			idcs,
			fmt.Sprintf(suffix, suffixVals...),
		)
	}

	// slog.Debug("PrepareUpsertNamedQuery", "q", q)

	return db.Unsafe().PrepareNamedContext(ctx, q)
}

func replacers(cols []string) string {
	sb := &strings.Builder{}
	sb.Grow(len(cols) * 10)
	for i, col := range cols {
		fmt.Fprintf(sb, ":%s", col)
		if i < len(cols)-1 {
			sb.WriteByte(',')
		}
	}
	return sb.String()
}

func columnsWithout(cols []string, skip ...string) []string {
	if len(skip) == 0 {
		return cols
	}
	ret := make([]string, 0, len(cols))

	for _, c := range cols {
		fnd := false
		for _, v := range skip {
			if strings.EqualFold(c, v) {
				fnd = true
				break
			}
		}
		if !fnd {
			ret = append(ret, c)
		}
	}
	return ret
}

func excluded(cols []string) []string {
	ret := make([]string, len(cols))
	for i, v := range cols {
		ret[i] = fmt.Sprintf("excluded.%s", v)
	}
	return ret
}
