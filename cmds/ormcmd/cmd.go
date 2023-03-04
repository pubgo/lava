package ormcmd

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/core/orm"
	"github.com/urfave/cli/v3"
	"gorm.io/gen"

	"github.com/pubgo/lava/core/migrates"
)

type params struct {
	Log         log.Logger
	Db          *orm.Client
	Generations migrates.Generation
}

func New() *cli.Command {
	return &cli.Command{
		Name:  "orm-gen",
		Usage: "orm gen",
		Commands: []*cli.Command{
			{
				Name:  "gen",
				Usage: "do gen query",
				Action: func(context *cli.Context) error {
					defer recovery.Exit()

					g := gen.NewGenerator(gen.Config{
						OutPath:           "./model/v2/query",
						ModelPkgPath:      "./model/v2/models",
						FieldWithTypeTag:  false,
						FieldWithIndexTag: false,
						FieldNullable:     true,
						FieldCoverable:    true,
						Mode:              gen.WithQueryInterface | gen.WithDefaultQuery | gen.WithoutContext,
					})

					p := di.Inject(new(params))
					g.UseDB(p.Db.DB)

					g.ApplyBasic(p.Generations(g)...)
					g.Execute()

					return nil
				},
			},
		},
	}
}
