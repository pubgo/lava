package running

import (
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"os"
	"sort"

	"github.com/pubgo/dix"
	"github.com/pubgo/dix/di"
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
	"github.com/pubgo/lava/cmds/ormcmd"
	"github.com/pubgo/lava/cmds/versioncmd"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/modules/gcnotifier"
	"github.com/pubgo/lava/pkg/cmdutil"

	// debug
	_ "github.com/pubgo/lava/core/debug/pprof"
	_ "github.com/pubgo/lava/core/debug/process"
	_ "github.com/pubgo/lava/core/debug/stats"
	_ "github.com/pubgo/lava/core/debug/trace"
	_ "github.com/pubgo/lava/core/debug/vars"
	_ "github.com/pubgo/lava/core/debug/version"

	// metric
	_ "github.com/pubgo/lava/core/metric/drivers/prometheus"

	// sqlite
	_ "github.com/pubgo/lava/core/orm/drivers/sqlite"

	// encoding
	_ "github.com/pubgo/lava/core/encoding/protobuf"
	_ "github.com/pubgo/lava/core/encoding/protojson"

	// logging
	_ "github.com/pubgo/lava/core/logging/logext/grpclog"
	_ "github.com/pubgo/lava/core/logging/logext/stdlog"

	_ "go.uber.org/automaxprocs"
)

func defaultProviders() []any {
	return []any{
		lifecycle.New,
		gcnotifier.New,

		middleware_accesslog.New,
		middleware_metric.New,
	}
}

func InitDefaultProvider(dix *dix.Dix) {
	for _, p := range defaultProviders() {
		dix.Provide(p)
	}
}

func init() {
	for _, p := range defaultProviders() {
		di.Provide(p)
	}
}

func Main(cmdL ...*cli.Command) {
	defer recovery.Exit()

	runmode.Check()

	cmdL = append(cmdL,
		versioncmd.New(),
		migratecmd.New(),
		healthcmd.New(),
		depcmd.New(),
		grpcservercmd.New(),
		httpservercmd.New(),
		ormcmd.New(),
	)

	app := &cli.App{
		Name:                   version.Project(),
		Suggest:                true,
		UseShortOptionHandling: true,
		Usage:                  cmdutil.UsageDesc("%s service", version.Project()),
		Version:                version.Version(),
		Flags:                  flags.GetFlags(),
		Commands:               cmdL,
		ExtraInfo:              runmode.GetSysInfo,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	assert.Must(app.Run(os.Args))
}
