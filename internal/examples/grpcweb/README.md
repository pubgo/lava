# gRPC-Web Example

HTTP/REST + gRPC 前端示例。前端通过可对外发布的 SDK **`@pubgo/lava-gateway-client`** 调用 gateway，用法对齐 agentrun `frontend/src/lib/rpc.ts`。

多协议（WebSocket / Native gRPC）请用 [`../grpcwebsocket`](../grpcwebsocket)。

## Run backend

```bash
go run ./internal/examples/grpcweb
# or multi-protocol:
go run ./internal/examples/grpcwebsocket
```

## Run frontend (SDK consumer)

```bash
cd sdk/js/gateway-client && npm install && npm run build
cd ../../../internal/examples/grpcweb/frontend
npm install
npm run dev
```

打开 http://localhost:3000/

页面演示：

- **HTTP REST**：`createHttpJsonClient` → `/v1/greeter/*`
- **RPC facade**：`createGatewayTransport({ format })` + `rpc.greeter.*`
  - `json`（默认，Network 明文）
  - `binary` / `text`（gRPC-Web）

控制台可访问 `window.rpc`（`rpc.setFormat('binary')`）。

## SDK

见 [`sdk/js/gateway-client/README.md`](../../../sdk/js/gateway-client/README.md)。
