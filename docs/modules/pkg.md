# Pkg 模块文档

`pkg/*` 放置可复用公共组件，既用于框架内部，也可供业务项目直接复用。

## 模块清单

| 模块               | 说明                      | 关键文件                         |
| ------------------ | ------------------------- | -------------------------------- |
| `pkg/gateway`      | HTTP/JSON ↔ gRPC 协议转换 | `pkg/gateway/_doc.go`            |
| `pkg/httputil`     | Fiber/HTTP 适配与路径处理 | `pkg/httputil/fiber.go`          |
| `pkg/grpcutil`     | gRPC 元数据与调试辅助     | `pkg/grpcutil/util.go`           |
| `pkg/grpcbuilder`  | gRPC server 配置构建      | `pkg/grpcbuilder/config.go`      |
| `pkg/fiberbuilder` | Fiber 配置构建            | `pkg/fiberbuilder/config.go`     |
| `pkg/netutil`      | 网络与地址工具            | `pkg/netutil/*.go`               |
| `pkg/cliutil`      | CLI 说明文本与示例拼接    | `pkg/cliutil/cmd.go`             |
| `pkg/wsproxy`      | WebSocket/HTTP 转发代理   | `pkg/wsproxy/websocket_proxy.go` |
| `pkg/wsbuilder`    | WebSocket 构建辅助        | `pkg/wsbuilder/ws.go`            |
| `pkg/k8sutil`      | K8s 环境探测工具          | `pkg/k8sutil/util.go`            |
| `pkg/proto`        | protobuf 生成代码产物     | `pkg/proto/lavapbv1/*.pb.go`     |
| `pkg/zrpc`         | zrpc Go runtime           | `pkg/zrpc/*.go`                  |
| `pkg/zrpcbridge`   | 将 NATS/zrpc 桥接到 `gateway.Mux` | `pkg/zrpcbridge/register.go` |

## Gateway 位置说明

`pkg/gateway` 是协议转换核心：

- 读取 gRPC `ServiceDesc`
- 生成 HTTP 路由映射
- 负责请求编解码与错误映射

```mermaid
flowchart LR
    HTTP[HTTP/JSON] --> GW[pkg/gateway]
    GW --> GRPC[gRPC ServiceDesc]
    GRPC --> BIZ[Service Impl]
```

## zrpc 位置说明

`pkg/zrpc` 是 protobuf over NATS 的核心 runtime：

- 服务端：`Server` / `RegisterUnary` / `RegisterStream`
- 客户端：`Client` / `CallUnary` / `OpenStream`
- 错误：`Status` / `ReplyErrorWithHeader`
- 统一请求响应：`request` / `response`（适配 `lava.Middleware`）

日志不直接在 runtime 输出，而是由 `lava.Middleware`（accesslog/metric）记录。

它被以下模块直接消费：

- `servers/zrpcs`
- `clients/zrpcc`
- `protoc-gen-zrpc-go` 生成代码

## zrpcbridge 位置说明

`pkg/zrpcbridge` 把 **NATS/zrpc** 订阅桥接到 `gateway.Mux` 已注册的 gRPC handler。
zrpc 是消息总线传输，**不是** gateway 的前端协议。

```
zrpc Client → NATS → zrpc.Server → zrpcbridge → gateway.Mux → handler
```

与 `servers` 的分工：

| 组件 | 职责 |
| --- | --- |
| `pkg/gateway` | 协议转换（HTTP/gRPC-Web/WS/native gRPC） |
| `servers/gatewayserver` | 对外监听并装配 Mux |
| `servers/zrpcs` | 纯 NATS 微服务宿主 |
| `pkg/zrpcbridge` | 可选：把 Mux handler 额外暴露到 NATS |

用法见 `pkg/zrpcbridge/README.md`；此前 `Mux.RegisterZrpc` 与 `grpc_server.zrpc_url` 已移除。
