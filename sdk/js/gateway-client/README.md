# `@pubgo/lava-gateway-client`

Browser / Node client for [lava](https://github.com/pubgo/lava) gateway:

- **HTTP/JSON** — REST routes from `google.api.http`
- **gRPC-Web** — binary frames, success trailers (`grpc-status`), optional gzip

WebSocket is intentionally out of scope for v0.1 (use native WS or a later SDK release).

## Install

```bash
# from this monorepo (example)
npm install file:../../../../sdk/js/gateway-client

# or after publish
npm install @pubgo/lava-gateway-client
```

For protobuf-ts generated clients, also install:

```bash
npm install @protobuf-ts/runtime @protobuf-ts/runtime-rpc
```

## HTTP/JSON

```ts
import { createHttpJsonClient, GatewayError } from "@pubgo/lava-gateway-client";

const http = createHttpJsonClient({ baseUrl: "http://localhost:8080" });

try {
  const { data, headers } = await http.post<{ name: string }, { message: string }>(
    "/v1/greeter/hello",
    { name: "World" },
  );
  console.log(data.message, headers.get("x-example-mw"));
} catch (e) {
  if (e instanceof GatewayError) {
    console.error(e.code, e.grpcMessage, e.httpStatus);
  }
}
```

## gRPC-Web + protobuf-ts

```ts
import { createGrpcWebTransport } from "@pubgo/lava-gateway-client";
import { GreeterServiceClient } from "./generated/greeter.client";

const transport = createGrpcWebTransport({
  baseUrl: "http://localhost:8080",
  acceptCompression: true, // grpc-accept-encoding: gzip
});

const client = new GreeterServiceClient(transport);
const call = client.sayHello({ name: "World" });
const response = await call.response;
const status = await call.status;
console.log(response.message, status.code);
```

Server-stream:

```ts
const call = client.watchHello({ name: "World", count: 3 });
for await (const msg of call.responses) {
  console.log(msg.message);
}
```

Client-stream / bidi over HTTP/gRPC-Web throw `UNIMPLEMENTED` (use WebSocket or native gRPC).

## Low-level gRPC-Web bytes API

```ts
import { createGrpcWebClient } from "@pubgo/lava-gateway-client";

const grpc = createGrpcWebClient({ baseUrl: "http://localhost:8080" });
const { message, trailers } = await grpc.unary(
  "/pkg.Service/Method",
  /* protobuf bytes */ new Uint8Array(...),
);
```

## Build

```bash
cd sdk/js/gateway-client
npm install
npm run build
```

## Example

See `internal/examples/grpcweb/frontend` — it consumes this package for both HTTP/JSON and gRPC-Web.
