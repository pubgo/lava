package migrates

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
)

type Migrate func() *gormigrate.Migration
type Migration = gormigrate.Migration
