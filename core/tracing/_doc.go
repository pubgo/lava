// Package tracing 提供 OpenTelemetry 链路追踪的配置与工具函数。
//
// 通过 tracingbuilder 在 DI 启动时初始化 TracerProvider，
// 支持 baggage 传播与 HTTP/gRPC 中间件集成。
package tracing

// https://github.com/opentracing-contrib/go-stdlib
// https://github.com/thundra-io/thundra-lambda-agent-go/tree/master/wrappers
// https://github.com/DataDog/dd-trace-go
// https://github.com/traefik/traefik/tree/master/pkg/tracing
// https://github.com/go-kratos/kratos/tree/main/middleware/tracing
