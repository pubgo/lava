# NATS/zrpc 支持

Gateway 提供 **NATS/zrpc 前端**，让 zrpc 客户端通过 NATS 调用与 HTTP、WebSocket、Native gRPC 相同的后端 handler，实现「`RegisterService` 一次，多协议复用」。

## 原理

```
zrpc Client → NATS (subject + queue) → zrpc.Server → streamZrpc → Dispatcher → Mux → handler
```

- 服务**只注册在** `gateway.Mux`（`RegisterService` / `RegisterProxy`）
- `Mux.RegisterZrpc` 为每个已注册 RPC 方法绑定 NATS queue subscription
- 内部使用 `streamZrpc` 将 zrpc 流归一化为 `grpc.ServerStream`，再交给统一 `Dispatcher`

## Subject 约定

NATS subject 由 gRPC 全方法名推导：

| gRPC Full Method | NATS Subject（默认 prefix `svc.`） |
| --- | --- |
| `/pkg.v1.Service/Method` | `svc.pkg.v1.Service/Method` |

可通过 `ZrpcConfig.SubjectPrefix` 或配置项 `zrpc_subject_prefix` 自定义前缀。

## 用法

### 代码

```go
package main

import (
    "github.com/nats-io/nats.go"
    "github.com/pubgo/lava/v2/pkg/gateway"
    "github.com/pubgo/lava/v2/pkg/zrpc"
    pb "your/proto/package"
)

func main() {
    mux := gateway.NewMux()
    mux.RegisterService(&pb.GreeterService_ServiceDesc, &greeterService{})

    nc, err := nats.Connect("nats://127.0.0.1:4222")
    if err != nil {
        panic(err)
    }
    defer nc.Close()

    srv := zrpc.NewServer(nc)
    if err = mux.RegisterZrpc(srv, gateway.ZrpcConfig{
        Queue: "greeter",
    }); err != nil {
        panic(err)
    }
    if err = nc.Flush(); err != nil {
        panic(err)
    }

    select {} // 保持 NATS 订阅
}
```

### grpc-server 配置

在 `grpc_server` 里同时配置 `zrpc_url` 与 `zrpc_queue` 即可启用（与 HTTP、WebSocket、Native gRPC 共用同一个 `Mux`）：

```yaml
grpc_server:
  zrpc_url: nats://127.0.0.1:4222
  zrpc_queue: my-service
  # 可选，默认 svc.
  zrpc_subject_prefix: svc.
  websocket_port: 8081
  grpc_passthrough: true
  http:
    # ...
```

启用后，`servers/grpcs` 会在启动时连接 NATS、注册所有 Mux 方法，并使用与 HTTP/gRPC 相同的全局 middleware 链。

## 流模式

zrpc 前端复用统一 `Dispatcher`，支持：

| 模式 | zrpc 行为 |
| --- | --- |
| Unary | 单条 request/response protobuf 帧 |
| Server-Stream | 首帧 request，后续多帧 response |
| Client-Stream | 多帧 request，单帧 response |
| Bidi | 双向流式 protobuf 帧 |

## 与其他前端的关系

| 前端 | 入口 | 注册方式 |
| --- | --- | --- |
| HTTP/REST + gRPC-Web | `mux.Handler` (Fiber) | `mux.RegisterService` |
| WebSocket | `mux.WebSocketHandler()` (net/http) | 同上 |
| Native gRPC | `mux.GRPCServerOptions()` | 同上 |
| NATS/zrpc | `mux.RegisterZrpc()` | 同上 |

## 客户端调用

使用 zrpc 生成客户端或 `zrpcc` 工具，subject 需与 `ZrpcSubject` 规则一致。例如 Greeter 的 `SayHello`：

```
svc.grpcweb.example.v1.GreeterService/SayHello
```

独立 zrpc 服务（`servers/zrpcs`）与 Gateway zrpc 前端可并存：前者适合纯 NATS 微服务，后者适合「已有 Gateway handler，额外暴露 NATS 入口」。

## 当前限制

- Subject 由 gRPC full method 推导，与 zrpc 代码生成器里的 subject 常量需手动对齐（或统一使用相同命名规则）
- 未内置 NATS 连接重连；连接失败时 grpc-server 启动报错
- REST body 字段映射（`body:"field"`）在 zrpc 前端未启用（整条 protobuf 消息即请求体）
