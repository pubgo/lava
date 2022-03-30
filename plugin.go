package lava

// 加载插件
import (
	// 加载version插件
	_ "github.com/pubgo/lava/version"

	// 加载metric插件
	_ "github.com/pubgo/lava/core/metric"

	// set GOMAXPROCS
	_ "github.com/pubgo/lava/internal/plugins/automaxprocs"

	// 加载registry插件
	_ "github.com/pubgo/lava/core/registry/registry_driver/mdns"

	// 编码加载
	_ "github.com/pubgo/lava/encoding/json"

	// 加载protobuf编码
	_ "github.com/pubgo/lava/encoding/protobuf"

	// 用于系统诊断
	_ "github.com/pubgo/lava/internal/plugins/gops"

	// gc plugin
	_ "github.com/pubgo/lava/internal/plugins/gcnotifier"
)

// 加载middleware, 注意加载顺序
import (
	// 加载log记录拦截器
	_ "github.com/pubgo/lava/core/logging/log_plugin"

	// tracing插件, 依赖加载

	// 加载timeout拦截器
	_ "github.com/pubgo/lava/plugins/timeout"
)
