package plugins

// 加载插件
import (
	// 默认metric
	_ "github.com/pubgo/lava/core/metric/drivers/prometheus"

	// 编码加载
	_ "github.com/pubgo/lava/encoding/json"

	// 加载protobuf编码
	_ "github.com/pubgo/lava/encoding/protobuf"

	// 默认driver
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"

	// 默认注册中心
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"

	_ "github.com/pubgo/lava/logging/log_ext/grpclog"
)
