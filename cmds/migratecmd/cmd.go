package migratecmd

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/core/migrates"
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
	var options = gormigrate.DefaultOptions
	return &cli.Command{
		Name:  "migrate",
		Usage: "db migrate",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "id",
				Usage:       "migration id",
				Destination: &id,
			},
		},
		Before: func(context *cli.Context) error {
			p := di.Inject(new(params))
			options.TableName = p.Db.TablePrefix + gormigrate.DefaultOptions.TableName
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
					m := gormigrate.New(p.Db.DB, options, migrate(p.Migrations))
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

					var ids []string
					assert.Must(p.Db.Table(options.TableName).Select("id").Find(&ids).Error)

					for _, m := range migrate(p.Migrations) {
						p.Log.Info().Msgf("migration-id=%s %s", m.ID, generic.Ternary(generic.Contains(ids, m.ID), "done", "missing"))
						ids = generic.Delete(ids, m.ID)
					}

					for i := range ids {
						p.Log.Info().Msgf("migration-id=%s %s", ids[i], "undo")
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
					m := gormigrate.New(p.Db.DB, options, migrate(p.Migrations))
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
