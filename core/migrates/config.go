package migrates

// Config define options for all migrations.
type Config struct {
	// TableName is the migration table.
	TableName string `yaml:"table_name"`

	// IDColumnName is the name of column where the migration id will be stored.
	IDColumnName string `yaml:"id_column_name"`

	// IDColumnSize is the length of the migration id column
	IDColumnSize int `yaml:"id_column_size"`

	// UseTransaction makes GoMigrate execute migrations inside a single transaction.
	// Keep in mind that not all databases support DDL commands inside transactions.
	UseTransaction bool `yaml:"use_transaction"`

	// ValidateUnknownMigrations will cause migrate to fail if there's unknown migration
	// IDs in the database
	ValidateUnknownMigrations bool `yaml:"validate_unknown_migrations"`
}

// DefaultConfig can be used if you don't want to think about options.
var DefaultConfig = Config{
	TableName:                 "migrations",
	IDColumnName:              "id",
	IDColumnSize:              255,
	UseTransaction:            false,
	ValidateUnknownMigrations: false,
}
