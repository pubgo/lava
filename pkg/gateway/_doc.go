// Package gateway 提供多协议 gRPC Gateway 功能，实现 HTTP/JSON、gRPC-Web、
// WebSocket、原生 gRPC 等上层协议到底层 gRPC handler 的统一调度。
//
// # 分层架构
//
// Gateway 借鉴 connectrpc/vanguard-go 的思想，将请求处理分为三层：
//
//   - 前端协议层 Frontend：httpFrontend（HTTP/REST + gRPC-Web，基于 Fiber）、
//     wsFrontend（WebSocket，基于 coder/websocket + net/http）、
//     GRPCPassthroughStreamHandler（原生 gRPC 透传）。
//     每个前端把请求归一化为 grpc.ServerStream。
//   - 核心调度层 Dispatcher：统一处理 unary / server-stream / client-stream /
//     bidi 四种流模式，对接前端流与后端连接。各前端经统一入口
//     Dispatcher.DispatchFrontend 接入（按流模式自动决定是否预读请求）。
//   - 后端 gRPC 层 Backend：Mux 实现 grpc.ClientConnInterface，通过 inprocgrpc
//     进程内通道调用本地 handler，或经 remoteProxyCli 转发到远程服务。
//
// 底层 gRPC handler 注册一次（RegisterService / RegisterProxy），即可被多种
// 上层协议前端复用。NATS/zrpc 桥接见 pkg/zrpcbridge。
//
// # 对外服务宿主
//
// servers/gatewayserver 负责监听端口并装配 HTTP/WS/gRPC 前端。
//
// # 基本用法
//
//	mux := gateway.NewMux()
//	mux.RegisterService(&pb.MyService_ServiceDesc, &myServiceImpl{})
//	app.Use("/api", mux.Handler) // Fiber：HTTP/REST + gRPC-Web
//
// # WebSocket 用法
//
// WebSocket 前端基于 coder/websocket，必须运行在标准 net/http 栈上：
//
//	wsHandler := mux.WebSocketHandler(gateway.WithWSOriginPatterns("example.com"))
//	http.ListenAndServe(":8081", wsHandler)
//
// # 原生 gRPC 透传
//
// 让原生 gRPC 客户端复用同一套 handler（外层 Server 不重复注册服务）：
//
//	grpcServer := grpc.NewServer(mux.GRPCServerOptions()...)
//
// # NATS/zrpc
//
// 见 pkg/zrpcbridge 与 servers/zrpcs。
//
//   - https://github.com/connectrpc/vanguard-go
//   - https://cloud.google.com/service-infrastructure/docs/service-management/reference/rpc/google.api
//   - https://google.aip.dev/123
//   - https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md
package gateway

// 相关项目参考:
// https://github.com/dgrr/http2
// https://github.com/r3labs/sse
// https://github.com/googollee/go-socket.io
// https://github.com/connectrpc/vanguard-go
// https://github.com/twitchtv/twirp
// https://github.com/stackrox/go-grpc-http1
// github.com/improbable-eng/grpc-web
// github.com/grpc-ecosystem/grpc-gateway/v2
// https://github.com/nhooyr/websocket
// https://github.com/connectrpc/vanguard-go
// https://github.com/emcfarlane/larking

// https://github.com/GoogleCloudPlatform/esp-v2/blob/master/src/go/configgenerator/routegen/helpers/backend_route.go
// https://github.com/vine-io/vine/blob/e435e77c14082e84c92028ef035d1d7597e8616b/lib/api/router/httprule/runtime.go
// https://github.com/grpc-ecosystem/grpc-gateway/blob/1bf77dd97e2f74c7511a7405ad7c950d36e45894/runtime/mux.go#L313
// https://github.com/solo-io/gloo
// https://github.com/altipla-consulting/protoc-gen-grpc_browser/tree/master
// https://github.com/emcfarlane/larking/blob/91250e03da0d4670288dbc663aab7481e07c81dc/larking/rules.go#L53
// https://github.com/connectrpc/vanguard-go
// https://github.com/stackrox/go-grpc-http1
// https://github.com/flakrimjusufi/grpc-with-rest
// https://github.com/tidwall/match/blob/master/match.go
