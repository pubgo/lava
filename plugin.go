package lava

// 加载插件
import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/lava/core/metric/drivers/prometheus"
	"github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/core/tracing/tracing_middleware"
	"github.com/pubgo/lava/logging/logmiddleware"

	// set GOMAXPROCS
	_ "github.com/pubgo/lava/modules/automaxprocs"

	// gc plugin
	_ "github.com/pubgo/lava/modules/gcnotifier"

	// 用于系统诊断
	_ "github.com/pubgo/lava/modules/gops"
)

// 加载插件
import (
	// 加载protobuf编码
	_ "github.com/pubgo/lava/encoding/protobuf"

	// 默认driver
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"

	// 默认注册中心
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"

	_ "github.com/pubgo/lava/logging/logext/grpclog"
)

func init() {
	// 加载middleware, 注意加载顺序
	di.Provide(logmiddleware.Middleware)
	di.Provide(requestid.Middleware)
	di.Provide(prometheus.New)
	di.Provide(tracing_middleware.Middleware)
	di.Provide(tracing.New)
}
