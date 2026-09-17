# gRPC-Web Example

HTTP/REST + gRPC-Web（Fiber）精简示例，前端通过可对外发布的 SDK **`@pubgo/lava-gateway-client`** 调用 gateway。

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
cd ../../internal/examples/grpcweb/frontend
npm install
npm run dev
```

打开 http://localhost:3000/（Vite 代理到 `:8080`）。

页面演示：

- `createHttpJsonClient` → `/v1/greeter/*`
- `createGrpcWebTransport` + protobuf-ts `GreeterServiceClient`（`acceptCompression: true`）

## SDK

见 [`sdk/js/gateway-client/README.md`](../../../sdk/js/gateway-client/README.md)。

```ts
import { createHttpJsonClient, createGrpcWebTransport } from "@pubgo/lava-gateway-client";
```
