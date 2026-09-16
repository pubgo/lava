# gRPC-Web Example

HTTP/REST + gRPC-Web（Fiber）精简示例。多协议（WebSocket / Native gRPC）请用 [`../grpcwebsocket`](../grpcwebsocket)。

## Run

```bash
go run ./internal/examples/grpcweb
```

浏览器：http://localhost:8080/

```bash
curl -X POST http://localhost:8080/v1/greeter/hello \
  -H 'Content-Type: application/json' \
  -d '{"name":"World"}'
```

可选前端（protobuf-ts）：

```bash
cd frontend && npm install && npm run dev
```

## Notes

- 已挂 `UseRPCMiddleware`（`X-Example-Mw`）
- CORS 暴露 `Grpc-Status` / `Grpc-Encoding` 等 gRPC-Web 头
