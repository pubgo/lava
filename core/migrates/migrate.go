package migrates

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
)

type Migration func() *gormigrate.Migration
type Migrations []Migration
