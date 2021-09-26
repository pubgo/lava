package lug

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "github.com/pubgo/lug/encoding/json"
	_ "github.com/pubgo/lug/encoding/protobuf"
	_ "github.com/pubgo/lug/internal/debug"
	_ "github.com/pubgo/lug/metric"
	_ "github.com/pubgo/lug/metric/prometheus"
	_ "github.com/pubgo/lug/plugins/automaxprocs"
	_ "github.com/pubgo/lug/plugins/grpclog"
	_ "github.com/pubgo/lug/registry/mdns"
	_ "github.com/pubgo/lug/tracing/jaeger"
)
