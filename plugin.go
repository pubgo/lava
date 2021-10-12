package lava

// 插件加载, 注意加载顺序
import (
	// 加载recovery插件
	_ "github.com/pubgo/lava/internal/plugins/recovery"

	// 加载request_id插件
	_ "github.com/pubgo/lava/plugins/request_id"

	// 默认链路追中加载
	_ "github.com/pubgo/lava/tracing/jaeger"

	// 加载logger插件
	_ "github.com/pubgo/lava/internal/plugins/logger"

	// 加载debug插件
	_ "github.com/pubgo/lava/internal/plugins/debug"

	// 默认metric加载
	_ "github.com/pubgo/lava/metric/prometheus"

	_ "github.com/pubgo/lava/plugins/automaxprocs"

	// grpc log插件加载
	_ "github.com/pubgo/lava/plugins/grpclog"

	// 默认注册中心加载
	_ "github.com/pubgo/lava/plugins/registry/mdns"

	// 默认编码
	_ "github.com/pubgo/lava/encoding/json"
	_ "github.com/pubgo/lava/encoding/protobuf"
)
