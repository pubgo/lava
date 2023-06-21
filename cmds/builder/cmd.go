package builder

import (
	"os"
	"sort"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
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
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/modules/gcnotifier"
	"github.com/pubgo/lava/pkg/cmdutil"

	// debug
	_ "github.com/pubgo/lava/core/debug/pprof"
	_ "github.com/pubgo/lava/core/debug/process"
	_ "github.com/pubgo/lava/core/debug/trace"
	_ "github.com/pubgo/lava/core/debug/vars"
	_ "github.com/pubgo/lava/core/debug/version"

	// metric
	_ "github.com/pubgo/lava/core/metrics/drivers/prometheus"

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

var defaultProviders = []any{
	lifecycle.New,
	gcnotifier.New,

	middleware_accesslog.New,
	middleware_metric.New,

	logging.New,
	metrics.New,

	scheduler.New,
}

func New(opts ...dix.Option) *dix.Dix {
	var di = dix.New(opts...)
	for _, p := range defaultProviders {
		di.Provide(p)
	}
	return di
}

func Run(di *dix.Dix, cmdL ...*cli.Command) {
	defer recovery.Exit()

	running.CheckVersion()

	cmdL = append(cmdL,
		versioncmd.New(),
		migratecmd.New(di),
		healthcmd.New(),
		depcmd.New(di),
		grpcservercmd.New(di),
		httpservercmd.New(di),
		ormcmd.New(di),
	)

	app := &cli.App{
		Name:                   version.Project(),
		Suggest:                true,
		UseShortOptionHandling: true,
		Usage:                  cmdutil.UsageDesc("%s service", version.Project()),
		Version:                version.Version(),
		Flags:                  flags.GetFlags(),
		Commands:               cmdL,
		ExtraInfo:              running.GetSysInfo,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	assert.Must(app.Run(os.Args))
}
