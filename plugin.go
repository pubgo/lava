package lug

import (
	// 默认编码
	_ "github.com/pubgo/lug/encoding/json"
	_ "github.com/pubgo/lug/encoding/protobuf"

	// debug服务
	_ "github.com/pubgo/lug/internal/debug"

	// 默认metric加载
	_ "github.com/pubgo/lug/metric/prometheus"

	_ "github.com/pubgo/lug/plugins/automaxprocs"

	// grpc log插件加载
	_ "github.com/pubgo/lug/plugins/grpclog"

	// 默认注册中心加载
	_ "github.com/pubgo/lug/plugins/registry/mdns"

	// 请求ID
	_ "github.com/pubgo/lug/plugins/request_id"

	// 默认链路追中加载
	_ "github.com/pubgo/lug/tracing/jaeger"
)
