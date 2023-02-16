package migratecmd

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/clients/orm"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/core/migrates"
	"github.com/urfave/cli/v3"
)

type params struct {
	Log        log.Logger
	Db         *orm.Client
	Migrations []migrates.Migrate
}

func migrate(m []migrates.Migrate) []*gormigrate.Migration {
	var migrations []*gormigrate.Migration
	for i := range m {
		migrations = append(migrations, m[i]())
	}
	return migrations
}

func New() *cli.Command {
	var id string
	var ids []string
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
			p := di.Inject(new(params))
			assert.Must(p.Db.Table(gormigrate.DefaultOptions.TableName).Select("id").Find(&ids).Error)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "migrate",
				Usage:   "do migrate",
				Aliases: []string{"m"},
				Action: func(context *cli.Context) error {
					defer recovery.Exit()

					p := di.Inject(new(params))
					m := gormigrate.New(p.Db.DB, gormigrate.DefaultOptions, migrate(p.Migrations))
					if id == "" {
						assert.Must(m.Migrate())
					} else {
						assert.Must(m.MigrateTo(id))
					}
					p.Log.Info().Msg("migration ok")
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
					for _, m := range migrate(p.Migrations) {
						p.Log.Info().Msgf("migration-id=%s %s", m.ID, generic.Ternary(generic.Contains(ids, m.ID), "done", ""))
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
					m := gormigrate.New(p.Db.DB, gormigrate.DefaultOptions, migrate(p.Migrations))
					if id == "" {
						assert.Must(m.RollbackLast())
					} else {
						assert.Must(m.RollbackTo(id))
					}
					p.Log.Info().Msg("rollback last ok")
					return nil
				},
			},
		},
	}
}
