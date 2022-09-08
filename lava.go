package lava

import (
	"fmt"
	"os"
	"sort"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/cmds/depcmd"
	"github.com/pubgo/lava/cmds/grpcservercmd"
	"github.com/pubgo/lava/cmds/healthcmd"
	"github.com/pubgo/lava/cmds/versioncmd"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/version"
)

func Run(cmdL ...*cli.Command) {
	defer recovery.Exit()

	var app = &cli.App{
		Name:                   version.Project(),
		Suggest:                true,
		UseShortOptionHandling: true,
		Usage:                  fmt.Sprintf("%s service", version.Project()),
		Version:                version.Version(),
		Flags:                  flags.GetFlags(),
		Commands:               append(cmdL, versioncmd.New(), healthcmd.New(), depcmd.New(), grpcservercmd.New()),
		ExtraInfo:              runmode.GetVersion,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	assert.Must(app.Run(os.Args))
}
