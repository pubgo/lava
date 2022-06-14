package cmds

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/example/pkg/menuservice"
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
			xerror.RecoverAndExit()
			var p = dix.Inject(new(param)).(*param)
			p.M.SaveLocalMenusToDb()
			p.Log.Info("menu saving success")
			return nil
		},
	}
}
