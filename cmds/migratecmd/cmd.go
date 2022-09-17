package migratecmd

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/core/migrates"
	"github.com/pubgo/lava/logging"
)

type params struct {
	Log *logging.Logger
	Db  *orm.Client
}

func migrate(m []migrates.Migrate) []*gormigrate.Migration {
	var migrations []*gormigrate.Migration
	for i := range m {
		migrations = append(migrations, m[i]())
	}
	return migrations
}

func New(migrations []migrates.Migrate) *cli.Command {
	var id string

	return &cli.Command{
		Name:  "migrate",
		Usage: "db migrate",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "mid",
				Usage:       "migration id",
				Destination: &id,
			},
		},
		Before: func(context *cli.Context) error {
			gormigrate.DefaultOptions.TableName = "orm_migrations"
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:    "migrate",
				Usage:   "do migrate",
				Aliases: []string{"m"},
				Action: func(context *cli.Context) error {
					defer recovery.Exit()

					p := di.Inject(new(params))
					m := gormigrate.New(p.Db.DB, gormigrate.DefaultOptions, migrate(migrations))
					if id == "" {
						assert.Must(m.Migrate())
					} else {
						assert.Must(m.MigrateTo(id))
					}
					p.Log.Info("Migration run ok")
					return nil
				},
			},
			{
				Name:    "list",
				Usage:   "list migrate",
				Aliases: []string{"l"},
				Action: func(context *cli.Context) error {
					defer recovery.Exit()

					p := di.Inject(new(params))
					for _, m := range migrate(migrations) {
						p.Log.Info(fmt.Sprintf("migration-id=%s", m.ID))
					}
					time.Sleep(time.Millisecond * 10)
					return nil
				},
			},
			{
				Name:    "rollback",
				Usage:   "do rollback",
				Aliases: []string{"r"},
				Action: func(context *cli.Context) error {
					defer recovery.Exit()

					p := di.Inject(new(params))
					m := gormigrate.New(p.Db.DB, gormigrate.DefaultOptions, migrate(migrations))
					if id == "" {
						assert.Must(m.RollbackLast())
					} else {
						assert.Must(m.RollbackTo(id))
					}
					p.Log.Info("RollbackLast run ok")
					return nil
				},
			},
		},
	}
}
