package migrate

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/core/migrates"
	"github.com/pubgo/lava/logging"
)

type params struct {
	Log *logging.Logger `inject:""`
	Db  *orm.Client     `inject:""`
}

func Cmd() *cli.Command {
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
			defer xerror.RecoverAndExit()
			gormigrate.DefaultOptions.TableName = "orm_migrations"
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:    "migrate",
				Usage:   "do migrate",
				Aliases: []string{"m"},
				Action: func(context *cli.Context) error {
					defer xerror.RecoverAndExit()

					p := dix.Inject(new(params)).(*params)
					m := gormigrate.New(p.Db.DB, gormigrate.DefaultOptions, migrates.Migrations())
					if id == "" {
						xerror.Panic(m.Migrate())
					} else {
						xerror.Panic(m.MigrateTo(id))
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
					defer xerror.RecoverAndExit()

					p := dix.Inject(new(params)).(*params)
					for _, m := range migrates.Migrations() {
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
					defer xerror.RecoverAndExit()

					p := dix.Inject(new(params)).(*params)
					m := gormigrate.New(p.Db.DB, gormigrate.DefaultOptions, migrates.Migrations())
					if id == "" {
						xerror.Panic(m.RollbackLast())
					} else {
						xerror.Panic(m.RollbackTo(id))
					}
					p.Log.Info("RollbackLast run ok")
					return nil
				},
			},
		},
	}
}