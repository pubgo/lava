# gRPC-Web Example

HTTP/REST + gRPC-Web 前端示例。通过可发布的 SDK **`@pubgo/lava-gateway-client`** 调用 gateway。

多协议（WebSocket / Native gRPC）请用 [`../grpcwebsocket`](../grpcwebsocket)。

## Run backend

```bash
go run ./internal/examples/grpcweb
```

## Run frontend

```bash
cd sdk/js/gateway-client && npm install && npm run build
cd ../../../internal/examples/grpcweb/frontend
npm install
npm run generate   # 重新生成 protobuf-ts 客户端（含 WatchHello）
npm run dev
```

打开 http://localhost:3000/（Vite 默认会把 API 代理到 :8080，或与 backend 同源部署）。

## 页面能力

| 区块 | 说明 |
|------|------|
| **Unary** SayHello / SayGoodbye | REST + RPC；展示 request meta、response headers、trailers、status |
| **Server stream** WatchHello | **订阅推送**：`count=0` 持续推送直到「取消订阅」；客户端边收边显示 |
| **Bidi Chat** | HTTP/gRPC-Web 上预期 `UNIMPLEMENTED`（完整双向流见 grpcwebsocket） |
| **Transport** | `json` / `binary` / `text` 下拉切换 |
| **Headers** | 可填 `x-demo-token`、`x-request-id`、`x-demo-interval-ms`；服务端回显 `x-demo-echo` |

> 说明：这是「每条连接上的服务端长推送」（server-stream），适合通知/进度/行情类场景。
> 若要真正的多订阅者 fan-out（一个事件推给 N 个客户端），需要在服务端加 topic/broker；传输层仍可用同一套 server-stream 或 WebSocket。

控制台可访问 `window.rpc`（`rpc.setFormat('binary')`）。

## SDK

见 [`sdk/js/gateway-client/README.md`](../../../sdk/js/gateway-client/README.md)。
