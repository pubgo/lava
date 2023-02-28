package migrates

type Config struct {
	TableName                 string `yaml:"table_name"`
	ValidateUnknownMigrations bool   `yaml:"validate_unknown_migrations"`
}
