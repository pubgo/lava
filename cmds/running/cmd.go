package running

import (
	"fmt"
	"github.com/pubgo/lava/cmds/ormcmd"
	"os"
	"sort"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/runmode"
	"github.com/pubgo/funk/version"
	cli "github.com/urfave/cli/v3"

	"github.com/pubgo/lava/cmds/depcmd"
	"github.com/pubgo/lava/cmds/grpcservercmd"
	"github.com/pubgo/lava/cmds/healthcmd"
	"github.com/pubgo/lava/cmds/httpservercmd"
	"github.com/pubgo/lava/cmds/migratecmd"
	"github.com/pubgo/lava/cmds/versioncmd"
	"github.com/pubgo/lava/core/flags"

	_ "github.com/pubgo/lava/core/debug/pprof"
	_ "github.com/pubgo/lava/core/debug/process"
	_ "github.com/pubgo/lava/core/debug/stats"
	_ "github.com/pubgo/lava/core/debug/trace"
	_ "github.com/pubgo/lava/core/debug/vars"
	_ "github.com/pubgo/lava/core/debug/version"
	_ "github.com/pubgo/lava/core/metric/drivers/prometheus"
	_ "github.com/pubgo/lava/core/orm/drivers/sqlite"

	// 加载插件
	_ "github.com/pubgo/lava/encoding/protobuf"
	_ "github.com/pubgo/lava/encoding/protojson"
)

func Main(cmdL ...*cli.Command) {
	defer recovery.Exit()
	cmdL = append(cmdL,
		versioncmd.New(),
		migratecmd.New(),
		healthcmd.New(),
		depcmd.New(),
		grpcservercmd.New(),
		httpservercmd.New(),
		ormcmd.New(),
	)

	var app = &cli.App{
		Name:                   version.Project(),
		Suggest:                true,
		UseShortOptionHandling: true,
		Usage:                  fmt.Sprintf("%s service", version.Project()),
		Version:                version.Version(),
		Flags:                  flags.GetFlags(),
		Commands:               cmdL,
		ExtraInfo:              runmode.GetSysInfo,
		Before: func(context *cli.Context) error {
			version.Check()
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	assert.Must(app.Run(os.Args))
}
