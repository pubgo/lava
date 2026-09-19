# `@pubgo/lava-gateway-client`

Browser / Node client for [lava](https://github.com/pubgo/lava) gateway.

Designed to match how real apps (e.g. agentrun `frontend/src/lib/rpc.ts`) talk to lava:

| API | Use |
| --- | --- |
| `createGatewayTransport({ format })` | Unified protobuf-ts `RpcTransport`: `json` / `binary` / `text` |
| `createHttpJsonClient` | REST routes from `google.api.http` (`/v1/...`) |
| `createJsonRpcTransport` | ProtoJSON over `POST /{package.Service}/{Method}` (+ NDJSON streams) |
| `createGrpcWebTransport` | gRPC-Web frames (trailers + optional gzip) |

## Install

```bash
cd sdk/js/gateway-client && npm install && npm run build

# from an app / example
npm install file:../../../../sdk/js/gateway-client
npm install @protobuf-ts/runtime @protobuf-ts/runtime-rpc
```

## App pattern (agentrun-style)

```ts
import { createGatewayTransport } from "@pubgo/lava-gateway-client";
import { GreeterServiceClient } from "./generated/greeter.client";

const transport = createGatewayTransport({
  baseUrl: "/api",          // or https://api.example.com
  format: "json",           // json | binary | text
  acceptCompression: true,  // binary/text only
});

export const rpc = {
  greeter: new GreeterServiceClient(transport),
};

const { response } = await rpc.greeter.sayHello({ name: "World" });
```

- **`json`** (dev): readable in Network panel; errors as lava/errorpb JSON
- **`binary`** (prod): `application/grpc-web+proto`
- **`text`**: `application/grpc-web-text+proto`

## HTTP REST

```ts
import { createHttpJsonClient, GatewayError } from "@pubgo/lava-gateway-client";

const http = createHttpJsonClient({ baseUrl: "http://localhost:8080" });
const { data } = await http.post("/v1/greeter/hello", { name: "World" });
```

## Example

`internal/examples/grpcweb/frontend` consumes this package with an `rpc` facade and a transport-format dropdown.

## Build

```bash
cd sdk/js/gateway-client
npm install
npm run build
```
