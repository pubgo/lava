package migrates

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
)

type Migration = gormigrate.Migration

var migrations []func() *gormigrate.Migration

func Register(m ...func() *gormigrate.Migration) {
	migrations = append(migrations, m...)
}

func Migrations() []*gormigrate.Migration {
	var mm []*gormigrate.Migration
	for i := range migrations {
		mm = append(mm, migrations[i]())
	}
	return mm
}
