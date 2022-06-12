package cmds

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/example/pkg/menuservice"
)

func Menu() *cli.Command {
	return &cli.Command{
		Name:  "menu",
		Usage: "Load local menu config to database",
		Action: func(c *cli.Context) error {
			xerror.RecoverAndExit()
			dix.Register(func(m *menuservice.Menu, log *logging.Logger) {
				xerror.Panic(m.SaveLocalMenusToDb())
				log.Info("menu saving success")
			})
			dix.Invoke()
			return nil
		},
	}
}
