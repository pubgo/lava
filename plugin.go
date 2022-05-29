package lava

// 加载插件
import (
	// set GOMAXPROCS
	_ "github.com/pubgo/lava/internal/plugins/automaxprocs"

	// 加载registry插件
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"

	// 编码加载
	_ "github.com/pubgo/lava/encoding/json"

	// 加载protobuf编码
	_ "github.com/pubgo/lava/encoding/protobuf"

	// gc plugin
	_ "github.com/pubgo/lava/internal/plugins/gcnotifier"

	// metric
	//_ "github.com/pubgo/lava/core/metric/metric_builder"

	// 用于系统诊断
	_ "github.com/pubgo/lava/imports/import_gops"
	_ "github.com/pubgo/lava/imports/import_grpc_log"
)

// 加载middleware, 注意加载顺序
import (
	// 加载log-record拦截器
	_ "github.com/pubgo/lava/imports/import_log_record_middleware"

	// 加载timeout拦截器
	_ "github.com/pubgo/lava/imports/import_timeout_middleware"
)
