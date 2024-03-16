package entmigrates

import (
	"context"

	entsql "entgo.io/ent/dialect/sql"
	sq "github.com/Masterminds/squirrel"
)

// The version of postgres against which the tests are run
// if the POSTGRES_VERSION environment variable is not set.
const defaultPostgresVersion = "15"

// MigrateFunc is the func signature for migrating.
type MigrateFunc func(ctx context.Context, builder sq.StatementBuilderType, tx *Tx) error

// RollbackFunc is the func signature for rollbacking.
type RollbackFunc func(ctx context.Context, builder sq.StatementBuilderType, tx *Tx) error

type Tx struct {
	driver *entsql.Driver
	sqlTx  *entsql.Tx
}

func (tx *Tx) GetDriver() *entsql.Driver {
	return tx.driver
}

func (tx *Tx) QueryList(ctx context.Context, query string, args []any, v any) error {
	rows, err := tx.sqlTx.QueryContext(ctx, query, args)
	if err != nil {
		return err
	}
	return entsql.ScanSlice(rows, v)
}

func (tx *Tx) QueryOne(ctx context.Context, query string, args []any, v any) error {
	rows, err := tx.sqlTx.QueryContext(ctx, query, args)
	if err != nil {
		return err
	}
	return entsql.ScanOne(rows, v)
}

func (tx *Tx) Exec(ctx context.Context, query string, args []any) error {
	_, err := tx.sqlTx.ExecContext(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}
