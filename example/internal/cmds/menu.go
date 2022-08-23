package cmds

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/example/internal/services/menuservice"
	"github.com/pubgo/lava/logging"
	"github.com/urfave/cli/v2"
)

type param struct {
	M   *menuservice.Menu
	Log *logging.Logger
}

func Menu() *cli.Command {
	return &cli.Command{
		Name:  "menu",
		Usage: "Load local menu config to database",
		Action: func(c *cli.Context) error {
			defer recovery.Exit()
			var p = dix.Inject(new(param))
			p.M.SaveLocalMenusToDb()
			p.Log.Info("menu saving success")
			return nil
		},
	}
}
