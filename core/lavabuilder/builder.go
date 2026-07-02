// Package lavabuilder 是 lava 应用的 DI 装配与 CLI 启动入口。
//
// 它负责：
//   - 创建 dix 容器并注册默认 Provider（日志、指标、生命周期、服务发现等）
//   - 通过 blank import 加载编解码器、debug 端点、日志扩展等 side-effect 模块
//   - 注册所有子命令（grpc/http/scheduler/tunnel 等）并挂载全局 CLI flag
//   - 绑定 signals.Context() 作为根 context，实现优雅关停
//
// 典型用法：
//
//	di := lavabuilder.New()
//	lavabuilder.Run(di)
package lavabuilder

import (
	"context"

	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/features/featureflags"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/redant"
	_ "go.uber.org/automaxprocs"

	"github.com/pubgo/lava/v2/clients/grpcc/grpccresolver"
	"github.com/pubgo/lava/v2/cmds/depcmd"
	"github.com/pubgo/lava/v2/cmds/grpcservercmd"
	"github.com/pubgo/lava/v2/cmds/healthcmd"
	"github.com/pubgo/lava/v2/cmds/httpservercmd"
	"github.com/pubgo/lava/v2/cmds/schedulercmd"
	"github.com/pubgo/lava/v2/cmds/tunnelcmd"
	"github.com/pubgo/lava/v2/cmds/versioncmd"
	_ "github.com/pubgo/lava/v2/core/debug/configview"
	_ "github.com/pubgo/lava/v2/core/debug/debug"
	"github.com/pubgo/lava/v2/core/debug/dixdebug"
	_ "github.com/pubgo/lava/v2/core/debug/featurehttp"
	_ "github.com/pubgo/lava/v2/core/debug/goroutine"
	_ "github.com/pubgo/lava/v2/core/debug/healthy"
	_ "github.com/pubgo/lava/v2/core/debug/loglevel"
	_ "github.com/pubgo/lava/v2/core/debug/pprof"
	_ "github.com/pubgo/lava/v2/core/debug/process"
	_ "github.com/pubgo/lava/v2/core/debug/ratelimit"
	_ "github.com/pubgo/lava/v2/core/debug/runtime"
	_ "github.com/pubgo/lava/v2/core/debug/statsviz"
	_ "github.com/pubgo/lava/v2/core/debug/sysinfo"
	_ "github.com/pubgo/lava/v2/core/debug/trace"
	_ "github.com/pubgo/lava/v2/core/debug/vars"
	_ "github.com/pubgo/lava/v2/core/debug/version"
	"github.com/pubgo/lava/v2/core/discovery"
	_ "github.com/pubgo/lava/v2/core/encoding/protobuf"
	_ "github.com/pubgo/lava/v2/core/encoding/protojson"
	"github.com/pubgo/lava/v2/core/flags"
	"github.com/pubgo/lava/v2/core/lifecycle/lifecyclebuilder"
	"github.com/pubgo/lava/v2/core/logging/logbuilder"
	_ "github.com/pubgo/lava/v2/core/logging/logext/grpclog"
	_ "github.com/pubgo/lava/v2/core/logging/logext/slog"
	_ "github.com/pubgo/lava/v2/core/logging/logext/stdlog"
	_ "github.com/pubgo/lava/v2/core/logging/loggerdebug"
	_ "github.com/pubgo/lava/v2/core/metrics/drivers/prometheus"
	"github.com/pubgo/lava/v2/core/metrics/metricbuilder"
	"github.com/pubgo/lava/v2/core/signals"
	"github.com/pubgo/lava/v2/pkg/cliutil"
)

// defaultProviders 是各命令共享的基础 DI Provider。
var defaultProviders = []any{
	grpccresolver.NewDirectBuilder,
	grpccresolver.NewDiscoveryBuilder,
	discovery.NewNoopDiscovery,

	logbuilder.New,
	metricbuilder.New,

	lifecyclebuilder.New,
}

// New 创建并初始化 dix 容器，注册 defaultProviders 并挂载 dix debug 端点。
func New(opts ...dix.Option) *dix.Dix {
	di := dix.New(append(opts, dix.WithValuesNull())...)
	for _, p := range defaultProviders {
		dix.Provide(di, p)
	}

	dixdebug.Init(di)
	return di
}

// Run 注册所有 CLI 子命令并以 signals.Context() 作为根 context 启动应用。
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
	dix.Provide(di, tunnelcmd.New)
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
