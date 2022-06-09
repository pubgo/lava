package lava

// 加载插件
import (
	_ "github.com/pubgo/lava/core/metric/drivers/prometheus"

	// set GOMAXPROCS
	_ "github.com/pubgo/lava/internal/modules/automaxprocs"

	// 编码加载
	_ "github.com/pubgo/lava/encoding/json"

	// 加载protobuf编码
	_ "github.com/pubgo/lava/encoding/protobuf"

	// gc plugin
	_ "github.com/pubgo/lava/internal/modules/gcnotifier"

	// metric
	//_ "github.com/pubgo/lava/core/metric/metric_builder"

	// 用于系统诊断
	_ "github.com/pubgo/lava/internal/modules/gops"
	_ "github.com/pubgo/lava/logging/log_ext/grpclog"
)

// 加载middleware, 注意加载顺序
import (
	_ "github.com/pubgo/lava/logging/middleware"

	_ "github.com/pubgo/lava/core/requestid"
)
