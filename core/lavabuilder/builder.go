package lavabuilder

import (
	"os"
	"sort"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	cli "github.com/urfave/cli/v3"

	"github.com/pubgo/lava/v2/clients/grpcc/grpccresolver"
	"github.com/pubgo/lava/v2/cmds/depcmd"
	"github.com/pubgo/lava/v2/cmds/grpcservercmd"
	"github.com/pubgo/lava/v2/cmds/healthcmd"
	"github.com/pubgo/lava/v2/cmds/httpservercmd"
	"github.com/pubgo/lava/v2/cmds/schedulercmd"
	"github.com/pubgo/lava/v2/cmds/versioncmd"
	"github.com/pubgo/lava/v2/core/discovery"
	"github.com/pubgo/lava/v2/core/flags"
	"github.com/pubgo/lava/v2/core/lifecycle/lifecyclebuilder"
	"github.com/pubgo/lava/v2/core/logging/logbuilder"
	"github.com/pubgo/lava/v2/core/metrics/metricbuilder"
	"github.com/pubgo/lava/v2/core/scheduler"
	"github.com/pubgo/lava/v2/core/signal"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/v2/pkg/cmdutil"

	_ "github.com/pubgo/lava/v2/core/debug/debug"
	//_ "github.com/pubgo/lava/v2/core/debug/gops"
	_ "github.com/pubgo/lava/v2/core/debug/pprof"
	_ "github.com/pubgo/lava/v2/core/debug/process"
	_ "github.com/pubgo/lava/v2/core/debug/statsviz"
	_ "github.com/pubgo/lava/v2/core/debug/trace"
	_ "github.com/pubgo/lava/v2/core/debug/vars"
	_ "github.com/pubgo/lava/v2/core/debug/version"

	// metric
	_ "github.com/pubgo/lava/v2/core/metrics/drivers/prometheus"

	// encoding
	_ "github.com/pubgo/lava/v2/core/encoding/protobuf"
	_ "github.com/pubgo/lava/v2/core/encoding/protojson"

	// logging
	_ "github.com/pubgo/lava/v2/core/logging/logext/grpclog"
	_ "github.com/pubgo/lava/v2/core/logging/logext/stdlog"

	_ "go.uber.org/automaxprocs"
)

var defaultProviders = []any{
	grpccresolver.NewDirectBuilder,
	grpccresolver.NewDiscoveryBuilder,
	discovery.NewNoopDiscovery,

	middleware_accesslog.New,
	middleware_metric.New,
	logbuilder.New,
	metricbuilder.New,

	lifecyclebuilder.New,
	scheduler.New,
}

func New(opts ...dix.Option) *dix.Dix {
	di := dix.New(append(opts, dix.WithValuesNull())...)
	for _, p := range defaultProviders {
		dix.Provide(di, p)
	}
	return di
}

func Run(di *dix.Dix) {
	defer recovery.Exit()

	dix.Provide(di, versioncmd.New)
	dix.Provide(di, healthcmd.New)
	dix.Provide(di, depcmd.New)
	dix.Provide(di, grpcservercmd.New)
	dix.Provide(di, httpservercmd.New)
	dix.Provide(di, schedulercmd.New)
	dix.Inject(di, func(cmd []*cli.Command) {
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
