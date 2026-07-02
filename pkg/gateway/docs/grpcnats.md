# NATS/zrpc 桥接（已迁移）

NATS/zrpc 不再属于 `pkg/gateway` 前端。请参见：

- [pkg/zrpcbridge/README.md](../../zrpcbridge/README.md) — 将 NATS 订阅桥接到 `gateway.Mux`
- [servers/zrpcs](../../../servers/zrpcs/doc.go) — 纯 zrpc 微服务宿主

## 简要用法

```go
import "github.com/pubgo/lava/v2/pkg/zrpcbridge"

_ = zrpcbridge.RegisterMux(zrpcSrv, mux, zrpcbridge.Config{Queue: "my-service"})
```

原先 `grpc_server.zrpc_url` 配置项已移除；请在应用 DI 中显式连接 NATS 并调用 `RegisterMux`。
