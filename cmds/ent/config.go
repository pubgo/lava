package ent

import (
	_ "embed"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/pubgo/funk/assert"
	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var sqlConfigFile string
var sqlTmpl sqlConfig

func init() {
	assert.Exit(yaml.Unmarshal([]byte(sqlConfigFile), &sqlTmpl), sqlConfigFile)
}

type sqlConfig struct {
	CreateTable     string `yaml:"create_table"`
	CreateMigration string `yaml:"create_migration"`
	DropMigration   string `yaml:"drop_migration"`
	ListMigration   string `yaml:"list_migration"`
}

// Config define config for all migrations.
type Config struct {
	GenTx func(db *entsql.Driver) Tx `yaml:"-" json:"-"`

	// TableName is the migration table.
	TableName string `yaml:"table_name"`

	// ColumnName is the name of column where the migration id will be stored.
	ColumnName string `yaml:"column_name"`

	// ColumnSize is the length of the migration id column
	ColumnSize int `yaml:"column_size"`

	MigrationPath string `yaml:"migration_path"`
}

// DefaultConfig can be used if you don't want to think about config.
var DefaultConfig = Config{
	TableName:     "migrations",
	ColumnName:    "id",
	ColumnSize:    255,
	MigrationPath: "./internal/schema/migrations",
}
