package ent

import (
	"context"
	"errors"

	"database/sql/driver"
	entsql "entgo.io/ent/dialect/sql"
)

// The version of postgres against which the tests are run
// if the POSTGRES_VERSION environment variable is not set.
const defaultPostgresVersion = "15"

// MigrateFunc is the func signature for migrating.
type MigrateFunc func(ctx context.Context, tx Tx) error

// RollbackFunc is the func signature for rollbacking.
type RollbackFunc func(ctx context.Context, tx Tx) error

var (
	// ErrRollbackImpossible is returned when trying to rollback a migration
	// that has no rollback function.
	ErrRollbackImpossible = errors.New("ormigrate: It's impossible to rollback this migration")

	// ErrNoMigrationDefined is returned when no migration is defined.
	ErrNoMigrationDefined = errors.New("ormigrate: No migration defined")

	// ErrMissingID is returned when the ID od migration is equal to ""
	ErrMissingID = errors.New("ormigrate: Missing ID in migration")

	// ErrNoRunMigration is returned when any run migration was found while
	// running RollbackLast
	ErrNoRunMigration = errors.New("ormigrate: Could not find last run migration")

	// ErrMigrationIDDoesNotExist is returned when migrating or rolling back to a migration ID that
	// does not exist in the list of migrations
	ErrMigrationIDDoesNotExist = errors.New("ormigrate: Tried to migrate to an ID that doesn't exist")

	// ErrUnknownPastMigration is returned if a migration exists in the DB that doesn't exist in the code
	ErrUnknownPastMigration = errors.New("ormigrate: Found migration in DB that does not exist in code")
)

type Tx interface {
	driver.Tx
	entsql.ExecQuerier
}
