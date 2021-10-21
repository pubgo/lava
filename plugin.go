package lava

// 加载插件
import (
	// 加载debug插件
	_ "github.com/pubgo/lava/plugins/debug"

	// 加载metric插件
	_ "github.com/pubgo/lava/plugins/metric"

	_ "github.com/pubgo/lava/plugins/automaxprocs"

	// 加载registry插件
	_ "github.com/pubgo/lava/plugins/registry/mdns"

	// 编码加载
	_ "github.com/pubgo/lava/encoding/json"
	_ "github.com/pubgo/lava/encoding/protobuf"
)

// 加载拦截器, 注意加载顺序
import (
	// 加载log记录拦截器
	_ "github.com/pubgo/lava/middlewares/logRecord"

	// 加载trace记录拦截器
	_ "github.com/pubgo/lava/middlewares/traceRecord"

	// 加载timeout拦截器
	_ "github.com/pubgo/lava/middlewares/timeout"

	// 加载recovery拦截器, 最后一项, 最靠近业务handler
	_ "github.com/pubgo/lava/middlewares/recovery"
)
