package lava

// 加载插件
import (
	// set GOMAXPROCS
	_ "github.com/pubgo/lava/modules/automaxprocs"

	// gc plugin
	_ "github.com/pubgo/lava/modules/gcnotifier"

	// metric
	//_ "github.com/pubgo/lava/core/metric/metric_builder"

	// 用于系统诊断
	_ "github.com/pubgo/lava/modules/gops"
)

// 加载middleware, 注意加载顺序
import (
	_ "github.com/pubgo/lava/logging/logmiddleware"

	_ "github.com/pubgo/lava/core/requestid"
)

// 加载插件
import (
	// 默认metric
	_ "github.com/pubgo/lava/core/metric/drivers/prometheus"

	// 加载protobuf编码
	_ "github.com/pubgo/lava/encoding/protobuf"

	// 默认driver
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"

	// 默认注册中心
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"

	_ "github.com/pubgo/lava/logging/logext/grpclog"
)
