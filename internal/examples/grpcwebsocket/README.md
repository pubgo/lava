# Gateway Multi-Protocol Example

同一套 gRPC handler，同时暴露：

| 端口 | 前端 |
| --- | --- |
| `:8080` | HTTP/REST + gRPC-Web（Fiber） |
| `:8081` | WebSocket（net/http） |
| `:50051` | Native gRPC |

演示本轮契约：

- `gatewayserver.NewGatewaySurface` 装配入口
- `Mux.UseRPCMiddleware`（响应头 `X-Example-Mw`）
- gRPC-Web 成功 trailer / gzip 压缩协商
- HTTP 拒绝 client/bidi（`/v1/greeter/chat` → Unimplemented）

## Run

```bash
go run ./internal/examples/grpcwebsocket
```

浏览器：http://localhost:8080/

## Verify

```bash
# 另开终端
go run ./internal/examples/grpcwebsocket/verify/
```

覆盖：HTTP/JSON、HTTP client-stream 拒绝、gRPC-Web unary/gzip/server-stream、WS 四流、native gRPC。

## Related

- JS SDK（HTTP/JSON + gRPC-Web）：`sdk/js/gateway-client`（`@pubgo/lava-gateway-client`）
- HTTP/gRPC-Web 精简示例：`internal/examples/grpcweb`（前端消费 SDK）
- 设计文档：`pkg/gateway/docs/design-evolution.md`
