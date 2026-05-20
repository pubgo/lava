// Package gateway 提供 gRPC Gateway 功能，实现 HTTP/JSON 到 gRPC 的协议转换。
//
// Gateway 基于 Google API HTTP Annotation 规范，支持：
//   - HTTP Rule 解析：支持 google.api.http 注解，自动解析 RESTful 路径模板
//   - 协议转换：HTTP/JSON ↔ gRPC/Protobuf 双向自动转换
//   - 路由匹配：支持精确匹配、* 单段通配符、** 多段通配符、动词 (:verb)
//   - 服务注册：支持本地服务 (RegisterService) 和代理服务 (RegisterProxy)
//   - 中间件：Unary 和 Stream 拦截器
//   - 多种流类型：HTTP、WebSocket、进程内、代理
//   - gRPC Web 支持：支持 application/grpc-web 和 application/grpc-web-text 内容类型
//
// 基本用法:
//
//	mux := gateway.NewMux()
//	mux.RegisterService(&pb.MyService_ServiceDesc, &myServiceImpl{})
//	app.Use("/api", mux.Handler)
//
// gRPC Web 用法:
//
//	mux := gateway.NewMux()
//	mux.RegisterService(&pb.MyService_ServiceDesc, &myServiceImpl{})
//	http.ListenAndServe(":8080", mux)
//
// 浏览器可以通过 gRPC Web 协议调用 gRPC 服务。
//
// 参考资料:
//   - https://cloud.google.com/service-infrastructure/docs/service-management/reference/rpc/google.api
//   - https://google.aip.dev/123
//   - https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md
//
// 参考资料:
//   - https://cloud.google.com/service-infrastructure/docs/service-management/reference/rpc/google.api
//   - https://google.aip.dev/123
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
