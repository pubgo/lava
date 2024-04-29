package entmigrates

import (
	"context"

	"gorm.io/gorm"
)

// The version of postgres against which the tests are run
// if the POSTGRES_VERSION environment variable is not set.
const defaultPostgresVersion = "15"

// MigrateFunc is the func signature for migrating.
type MigrateFunc func(ctx context.Context, tx *gorm.DB) error

// RollbackFunc is the func signature for rollbacking.
type RollbackFunc func(ctx context.Context, tx *gorm.DB) error
