package lavabuilder

import (
	"context"

	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/features/featureflags"
	"github.com/pubgo/funk/v2/recovery"
	// metric
	"github.com/pubgo/redant"
	_ "go.uber.org/automaxprocs"

	"github.com/pubgo/lava/v2/clients/grpcc/grpccresolver"
	"github.com/pubgo/lava/v2/cmds/depcmd"
	"github.com/pubgo/lava/v2/cmds/grpcservercmd"
	"github.com/pubgo/lava/v2/cmds/healthcmd"
	"github.com/pubgo/lava/v2/cmds/httpservercmd"
	"github.com/pubgo/lava/v2/cmds/schedulercmd"
	"github.com/pubgo/lava/v2/cmds/versioncmd"
	_ "github.com/pubgo/lava/v2/core/debug/debug"
	"github.com/pubgo/lava/v2/core/debug/dixdebug"
	//_ "github.com/pubgo/lava/v2/core/debug/gops"
	_ "github.com/pubgo/lava/v2/core/debug/pprof"
	_ "github.com/pubgo/lava/v2/core/debug/process"
	_ "github.com/pubgo/lava/v2/core/debug/statsviz"
	_ "github.com/pubgo/lava/v2/core/debug/trace"
	_ "github.com/pubgo/lava/v2/core/debug/vars"
	_ "github.com/pubgo/lava/v2/core/debug/version"
	"github.com/pubgo/lava/v2/core/discovery"
	// encoding
	_ "github.com/pubgo/lava/v2/core/encoding/protobuf"
	_ "github.com/pubgo/lava/v2/core/encoding/protojson"
	"github.com/pubgo/lava/v2/core/flags"
	"github.com/pubgo/lava/v2/core/lifecycle/lifecyclebuilder"
	"github.com/pubgo/lava/v2/core/logging/logbuilder"
	// logging
	_ "github.com/pubgo/lava/v2/core/logging/logext/grpclog"
	_ "github.com/pubgo/lava/v2/core/logging/logext/slog"
	_ "github.com/pubgo/lava/v2/core/logging/logext/stdlog"
	_ "github.com/pubgo/lava/v2/core/metrics/drivers/prometheus"
	"github.com/pubgo/lava/v2/core/metrics/metricbuilder"
	"github.com/pubgo/lava/v2/core/signals"
	"github.com/pubgo/lava/v2/pkg/cliutil"
)

var defaultProviders = []any{
	grpccresolver.NewDirectBuilder,
	grpccresolver.NewDiscoveryBuilder,
	discovery.NewNoopDiscovery,

	logbuilder.New,
	metricbuilder.New,

	lifecyclebuilder.New,
}

func New(opts ...dix.Option) *dix.Dix {
	di := dix.New(append(opts, dix.WithValuesNull())...)
	for _, p := range defaultProviders {
		dix.Provide(di, p)
	}

	dixdebug.Init(di)
	return di
}

func Run(di *dix.Dix) {
	defer recovery.Exit(func(err error) error {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	})

	dix.Provide(di, versioncmd.New)
	dix.Provide(di, healthcmd.New)
	dix.Provide(di, depcmd.New)
	dix.Provide(di, grpcservercmd.New)
	dix.Provide(di, httpservercmd.New)
	dix.Provide(di, schedulercmd.New)
	dix.Inject(di, func(commands []*redant.Command) {
		app := &redant.Command{
			Use:      version.Project(),
			Short:    cliutil.UsageDesc("%s service", version.Project()),
			Options:  append(flags.GetFlags(), featureflags.GetFlags()...),
			Children: commands,
		}

		assert.Must(app.Run(signals.Context()))
	})
}
