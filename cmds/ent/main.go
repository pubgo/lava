package ent

import (
	"entgo.io/ent/dialect/sql"
	"fmt"
	"log"
	"os"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v3"
	// https://github.com/ent/ent/blob/master/cmd/internal/base/base.go
)

type params struct {
	Drv            dialect.Driver
	Log            log.Logger
	Tables         []*schema.Table
	MigrateOptions []schema.MigrateOption
}

func init() {
}

func New() *cli.Command {
	return &cli.Command{
		Name:  "ent",
		Usage: "ent manager",
		Commands: []*cli.Command{
			{
				Name:  "gen",
				Usage: "do gen query",
				Action: func(context *cli.Context) error {
					defer recovery.Exit()
					return entc.Generate("./ent/schema", &gen.Config{
						Features: []gen.Feature{
							gen.FeatureVersionedMigration,
							gen.FeatureUpsert,
							gen.FeatureSchemaConfig,
							gen.FeatureModifier,
							gen.FeatureExecQuery,
						},
					})
				},
			},

			{
				Name:  "generate migration",
				Usage: "automatically generate migration files for your Ent schema:",
				Action: func(context *cli.Context) error {
					// atlas migrate lint \
					//  --dev-url="docker://mysql/8/test" \
					//  --dir="file://ent/migrate/migrations" \
					//  --latest=1

					// atlas migrate diff migration_name \
					//  --dev-url "docker://mysql/8/ent"
					//  --dir "file://ent/migrate/migrations" \
					//  --to "ent://ent/schema" \

					defer recovery.Exit()
					return entc.Generate("./ent/schema", &gen.Config{
						Features: []gen.Feature{
							gen.FeatureVersionedMigration,
							gen.FeatureUpsert,
							gen.FeatureSchemaConfig,
							gen.FeatureModifier,
							gen.FeatureExecQuery,
						},
					})
				},
			},

			{
				Name:  "apply migration",
				Usage: "apply the pending migration files onto the database",
				Action: func(context *cli.Context) error {
					// atlas migrate status \
					//  --dir "file://ent/migrate/migrations" \
					//  --url "mysql://root:pass@localhost:3306/example"

					//atlas migrate apply \
					//--dir "file://ent/migrate/migrations" \
					//--url "mysql://root:pass@localhost:3306/example"

					// atlas migrate status \
					//  --dir "file://ent/migrate/migrations" \
					//  --url "mysql://root:pass@localhost:3306/example"

				},
			},

			// atlas migrate hash \
			//  --dir "file://my/project/migrations"

			// atlas migrate hash --dir file://<path-to-your-migration-directory>
			// atlas migrate status \
			//  --dir "file://ent/migrate/migrations" \
			//  --url "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable&search_path=public"
			// atlas migrate lint
			// atlas migrate lint \
			//  --dev-url="docker://mysql/8/test" \
			//  --dir="file://ent/migrate/migrations" \
			//  --latest=1
			// atlas migrate apply \
			//  --dir "file://ent/migrate/migrations" \
			//  --url "postgres://postgres:pass@localhost:5432/database?search_path=public&sslmode=disable"
			// atlas migrate validate --dir file://<path-to-your-migration-directory>
			// atlas migrate hash --dir file://<path-to-your-migration-directory>
			// atlas migrate new add_user

			// entproto
			// protoc -I=.. --go_out=.. --go-grpc_out=.. --go_opt=paths=source_relative --entgrpc_out=.. --entgrpc_opt=paths=source_relative,schema_path=../../schema --go-grpc_opt=paths=source_relative entpb/entpb.proto
			// https://github.com/ent/contrib/blob/master/entproto/cmd/entproto/main.go
			// ent new Todo
			//  ent describe ./ent/schema
			//  ent gen ./ent/schema
			// go run ariga.io/entimport/cmd/entimport -dsn "mysql://root:pass@tcp(localhost:3308)/test" -tables "users"
			// atlas migrate apply --dir file://ent/migrate/migrations --url mysql://root:pass@localhost:3306/db
			// atlas migrate apply --dir file://ent/migrate/migrations --url mysql://root:pass@localhost:3306/db --baseline 20221114165732
			// atlas migrate status --dir file://ent/migrate/migrations --url postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable&search_path=public
			//  atlas schema inspect -u "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable&search_path=public" > schema.hcl

			{
				Name:  "create-migration",
				Usage: "create migration",
				Action: func(context *cli.Context) error {
					defer recovery.Exit()
					ctx := context.Context
					dir, err := atlas.NewLocalDir("./ent/migrate/migrations")
					assert.Must(err)

					hash := assert.Must1(dir.Checksum())
					assert.Must(atlas.WriteSumFile(dir, hash))

					assert.Must(atlas.Validate(dir))

					// Migrate diff options.
					opts := []schema.MigrateOption{
						schema.WithDir(dir),                         // provide migration directory
						schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
						schema.WithDialect(dialect.MySQL),           // Ent dialect to use
						schema.WithFormatter(atlas.DefaultFormatter),
						schema.WithDropIndex(true),
						schema.WithDropColumn(true),
					}

					if len(os.Args) != 2 {
						log.Fatalln("migration name is required. Use: 'go run -mod=mod ent/migrate/main.go <name>'")
					}

					var drv dialect.Driver
					migrate, err := schema.NewMigrate(drv, opts...)
					if err != nil {
						return fmt.Errorf("ent/migrate: %w", err)
					}

					var Tables []*schema.Table
					if err := migrate.VerifyTableRange(ctx, Tables); err != nil {
						log.Fatalf("failed verifyint range allocations: %v", err)
					}

					return migrate.NamedDiff(ctx, "change name", Tables...)
				},
			},
		},
	}
}

func Open(driverName, dataSourceName string) (*sql.Driver, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return drv, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}
