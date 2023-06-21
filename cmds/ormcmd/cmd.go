package ormcmd

import (
	"github.com/pubgo/dix"
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
	Generations migrates.Generation
}

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "orm",
		Usage: "orm manager",
		Commands: []*cli.Command{
			{
				Name:  "gen-model",
				Usage: "do gen query",
				Action: func(context *cli.Context) error {
					defer recovery.Exit()

					g := gen.NewGenerator(gen.Config{
						OutPath:           "./internal/db/query",
						ModelPkgPath:      "./internal/db/models",
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
