package migrates

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
)

type Migration = gormigrate.Migration
type Migrations = []func() *gormigrate.Migration
