# Native gRPC 透传

Gateway 提供 **Native gRPC 透传前端**，让标准 gRPC 客户端通过 `grpc.Server` 调用与 HTTP/WebSocket 相同的后端 handler，实现「`RegisterService` 一次，多协议复用」。

## 原理

```
gRPC Client → grpc.Server (UnknownServiceHandler) → Mux (Backend) → inprocgrpc / proxy
```

- 服务**只注册在** `gateway.Mux`（`RegisterService` / `RegisterProxy`）
- 外层 `grpc.Server` 通过 `grpc.UnknownServiceHandler` 将所有 RPC 转发到 `Mux`
- 透传 handler 复用统一 `Dispatcher.DispatchFrontend`：原生 `grpc.ServerStream` 直接作为前端流，unary 走 `Invoke`、流式走 `NewStream`，四种流模式均支持

## 用法

### 代码

```go
mux := gateway.NewMux()
mux.RegisterService(&pb.MyService_ServiceDesc, impl)

// 方式 1：追加到已有 ServerOption
grpcServer := grpc.NewServer(mux.GRPCServerOptions(
    grpc.ChainUnaryInterceptor(myUnary),
)...)

// 方式 2：仅取透传 handler
handler := mux.GRPCPassthroughStreamHandler()
grpcServer := grpc.NewServer(grpc.UnknownServiceHandler(handler))
```

**注意**：启用透传后，不要再对同一个 `grpc.Server` 调用 `RegisterService`，否则会与 `UnknownServiceHandler` 冲突。

### gateway_server 配置

```yaml
gateway_server:
  grpc_passthrough: true
  websocket_port: 8081
```

开启后，`servers/gatewayserver` 仅在 `Mux` 上注册服务，gRPC 端口上的原生客户端与 HTTP/WS 前端共享同一套 handler。

> **默认值**：`grpc_passthrough` 默认为 `true`（handler 仅在 Mux 注册）。若需 legacy 双注册（同时挂在外层 `grpc.Server`），显式设为 `false`。

## 与其他前端的关系

| 前端 | 入口 | 注册方式 |
| --- | --- | --- |
| HTTP/REST + gRPC-Web | `mux.Handler` (Fiber) | `mux.RegisterService` |
| WebSocket | `mux.WebSocketHandler()` (net/http) | 同上 |
| Native gRPC | `mux.GRPCServerOptions()` | 同上，外层 Server 不重复注册 |

## 当前限制

- 外层 `grpc.Server` 的全局拦截器与 `UnknownServiceHandler` 的协作依赖 grpc-go 行为；业务拦截器建议配置在 `Mux.SetUnaryInterceptor` / `SetStreamInterceptor`（作用于 inproc 后端）
- 反射（reflection）、health 等服务仍需在外层 `grpc.Server` 单独注册（若需要）
