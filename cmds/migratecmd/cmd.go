package migratecmd

import (
	"path/filepath"
	"time"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v3"
	"gorm.io/gen"

	"github.com/pubgo/lava/core/migrates"
	"github.com/pubgo/lava/core/orm"
)

type params struct {
	Log         log.Logger
	Db          *orm.Client
	Migrations  []migrates.Migrate
	Generations migrates.Generation
}

func migrate(m []migrates.Migrate) []*migrates.Migration {
	var migrations []*migrates.Migration
	for i := range m {
		migrations = append(migrations, m[i]())
	}
	return migrations
}

func New(di *dix.Dix) *cli.Command {
	var id string

	options := migrates.DefaultConfig
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
			p := dix.Inject(di, new(params))
			options.TableName = p.Db.TablePrefix + migrates.DefaultConfig.TableName
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "migrate",
				Usage:   "do migrate",
				Aliases: []string{"m"},
				Action: func(context *cli.Context) error {
					defer recovery.Exit()

					p := dix.Inject(di, new(params))
					m := migrates.New(p.Db.DB, &options, migrate(p.Migrations))
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

					p := dix.Inject(di, new(params))

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
					defer recovery.Recovery(func(err error) {
						if errors.Is(err, migrates.ErrNoRunMigration) {
							return
						}

						assert.Exit(err)
					})

					p := dix.Inject(di, new(params))
					m := migrates.New(p.Db.DB, &options, migrate(p.Migrations))
					if id == "" {
						assert.Must(m.RollbackLast())
					} else {
						assert.Must(m.RollbackTo(id))
					}
					p.Log.Info().Msg("rollback last ok")
					return nil
				},
			},
			{
				Name:      "gen",
				Usage:     "do gen orm model and query code",
				Aliases:   []string{"g"},
				UsageText: "migrate gen [./internal/db]",
				Action: func(context *cli.Context) error {
					defer recovery.Exit()

					var genPath = "./internal/db"
					if context.NArg() > 0 {
						genPath = context.Args().First()
					}

					g := gen.NewGenerator(gen.Config{
						OutPath:           filepath.Join(genPath, "query"),
						ModelPkgPath:      filepath.Join(genPath, "models"),
						FieldWithTypeTag:  false,
						FieldWithIndexTag: false,
						FieldNullable:     true,
						FieldCoverable:    true,
						Mode:              gen.WithQueryInterface | gen.WithDefaultQuery | gen.WithoutContext,
					})

					p := dix.Inject(di, new(params))
					g.UseDB(p.Db.DB)

					g.ApplyBasic(p.Generations(g)...)
					g.Execute()

					return nil
				},
			},
		},
	}
}
