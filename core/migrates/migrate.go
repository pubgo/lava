package migrates

import (
	m "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gen"
)

type Migration = m.Migration
type Migrate func() *Migration
type Generation func(g *gen.Generator) []interface{}
