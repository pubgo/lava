# zrpcbridge

`pkg/zrpcbridge` 将 **NATS/zrpc** 订阅桥接到 `gateway.Mux` 已注册的 gRPC handler。

zrpc 是消息总线上的 RPC 传输，不是 HTTP/TCP 网关前端；本包负责在应用层把两者接起来，使 handler 在 Gateway 上注册一次后，也可通过 NATS 调用。

## 原理

```
zrpc Client → NATS (subject + queue) → zrpc.Server → zrpcbridge → gateway.Mux → handler
```

## Subject 约定

| gRPC Full Method | NATS Subject（默认 prefix `svc.`） |
| --- | --- |
| `/pkg.v1.Service/Method` | `svc.pkg.v1.Service/Method` |

## 用法

```go
mux := gateway.NewMux()
mux.RegisterService(&pb.GreeterService_ServiceDesc, &greeterService{})

nc, err := nats.Connect("nats://127.0.0.1:4222")
// ...

srv := zrpc.NewServer(nc)
if err = zrpcbridge.RegisterMux(srv, mux, zrpcbridge.Config{
    Queue: "greeter",
}); err != nil {
    panic(err)
}
```

## 与 servers 的关系

| 组件 | 职责 |
| --- | --- |
| `servers/gatewayserver` | 对外 HTTP/REST、gRPC-Web、WebSocket、原生 gRPC |
| `servers/zrpcs` | 纯 NATS zrpc 微服务宿主（代码生成 `Register...ZrpcRoutes`） |
| `pkg/zrpcbridge` | 可选：把已有 `gateway.Mux` handler 暴露到 NATS |

## 流模式

桥接复用 `gateway.Dispatcher`，支持 Unary / Server-Stream / Client-Stream / Bidi。

## 迁移说明

此前 `grpc_server.zrpc_url` / `Mux.RegisterZrpc` 已移除。若需 Gateway + NATS，在 DI 中显式调用 `zrpcbridge.RegisterMux`（或单独运行 `servers/zrpcs`）。
