package app

import (
	"os"
	"sort"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	cli "github.com/urfave/cli/v3"

	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/cmds/depcmd"
	"github.com/pubgo/lava/cmds/grpcservercmd"
	"github.com/pubgo/lava/cmds/healthcmd"
	"github.com/pubgo/lava/cmds/httpservercmd"
	"github.com/pubgo/lava/cmds/versioncmd"
	"github.com/pubgo/lava/core/discovery"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/services/metadata"

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
	grpcc_resolver.NewDirectBuilder,
	grpcc_resolver.NewDiscoveryBuilder,
	discovery.NewNoopDiscovery,

	middleware_accesslog.New,
	middleware_metric.New,
	logging.New,
	metrics.New,

	lifecycle.New,
	scheduler.New,

	metadata.New,
}

func NewBuilder(opts ...dix.Option) *dix.Dix {
	var di = dix.New(append(opts, dix.WithValuesNull())...)
	for _, p := range defaultProviders {
		di.Provide(p)
	}
	return di
}

func Run(di *dix.Dix) {
	defer recovery.Exit()

	running.CheckVersion()

	di.Provide(versioncmd.New)
	di.Provide(healthcmd.New)
	di.Provide(depcmd.New)
	di.Provide(grpcservercmd.New)
	di.Provide(httpservercmd.New)

	di.Inject(func(cmd []*cli.Command) {
		app := &cli.App{
			Name:                   version.Project(),
			Suggest:                true,
			UseShortOptionHandling: true,
			Usage:                  cmdutil.UsageDesc("%s service", version.Project()),
			Version:                version.Version(),
			Flags:                  flags.GetFlags(),
			Commands:               cmd,
			ExtraInfo:              running.GetSysInfo,
		}

		sort.Sort(cli.FlagsByName(app.Flags))
		sort.Sort(cli.CommandsByName(app.Commands))
		assert.Must(app.Run(os.Args))
	})
}
