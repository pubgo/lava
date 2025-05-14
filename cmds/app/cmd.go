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

	"github.com/pubgo/lava/clients/grpcc/grpccresolver"
	"github.com/pubgo/lava/cmds/depcmd"
	"github.com/pubgo/lava/cmds/grpcservercmd"
	"github.com/pubgo/lava/cmds/healthcmd"
	"github.com/pubgo/lava/cmds/httpservercmd"
	"github.com/pubgo/lava/cmds/schedulercmd"
	"github.com/pubgo/lava/cmds/versioncmd"
	"github.com/pubgo/lava/core/discovery"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/lifecycle/lifecyclebuilder"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metrics/metricbuilder"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/pkg/cmdutil"

	_ "github.com/pubgo/lava/core/debug/debug"
	// debug
	_ "github.com/pubgo/lava/core/debug/pprof"
	_ "github.com/pubgo/lava/core/debug/process"
	_ "github.com/pubgo/lava/core/debug/statsviz"
	_ "github.com/pubgo/lava/core/debug/trace"
	_ "github.com/pubgo/lava/core/debug/vars"
	_ "github.com/pubgo/lava/core/debug/version"

	// metric
	_ "github.com/pubgo/lava/core/metrics/drivers/prometheus"

	// encoding
	_ "github.com/pubgo/lava/core/encoding/protobuf"
	_ "github.com/pubgo/lava/core/encoding/protojson"

	// logging
	_ "github.com/pubgo/lava/core/logging/logext/grpclog"
	_ "github.com/pubgo/lava/core/logging/logext/stdlog"

	_ "go.uber.org/automaxprocs"
)

var defaultProviders = []any{
	grpccresolver.NewDirectBuilder,
	grpccresolver.NewDiscoveryBuilder,
	discovery.NewNoopDiscovery,

	middleware_accesslog.New,
	middleware_metric.New,
	logging.New,
	metricbuilder.New,

	lifecyclebuilder.New,
	scheduler.New,
}

func NewBuilder(opts ...dix.Option) *dix.Dix {
	di := dix.New(append(opts, dix.WithValuesNull())...)
	for _, p := range defaultProviders {
		di.Provide(p)
	}
	return di
}

func Run(di *dix.Dix) {
	defer recovery.Exit()

	di.Provide(versioncmd.New)
	di.Provide(healthcmd.New)
	di.Provide(depcmd.New)
	di.Provide(grpcservercmd.New)
	di.Provide(httpservercmd.New)
	di.Provide(schedulercmd.New)

	di.Inject(func(cmd []*cli.Command) {
		app := &cli.Command{
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
		assert.Must(app.Run(signal.Context(), os.Args))
	})
}
