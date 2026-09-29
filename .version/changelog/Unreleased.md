## [Unreleased]

<!-- 在此记录尚未发版的变更；发版时移至 CHANGELOG.md 并标注版本号。 -->

### Changed

- **BREAKING — RPC middleware 边界**：`UseRPCMiddleware` 只包装 `Mux.Dispatch` / `DispatchFrontend`（HTTP/WS/gRPC-Web 等前端路径），不再覆盖直接把 Mux 当 `grpc.ClientConnInterface` 用的调用（`pb.NewXxxClient(mux)`）。这类调用改挂 `UseBackendUnaryInterceptor` / `UseBackendStreamInterceptor`，或走前端 / `DispatchFrontend`。原因：`Invoke`/`NewStream` 在流开始时就返回，包住它们会提前释放请求 `timeout` 的 `CancelFunc`。详见 `pkg/gateway/docs/usage.md`。
- **metrics**: RPC 指标名统一为 `lava_rpc_total` / `lava_rpc_failed_total` / `lava_rpc_handling_seconds`，客户端与服务端共用一组，用 `side`/`kind`/`service`/`method`/`stream`/`proto`（失败另有 `code`）区分；`gateway_server_*` 已删除，抓取端与看板需同步改。

### Added

- **sdk/js/gateway-client**：`@pubgo/lava-gateway-client`，浏览器/Node 侧 HTTP/JSON、JSON-RPC 与 gRPC-Web 客户端（`createHttpJsonClient` / `createJsonRpcTransport` / `createGrpcWebTransport`）。
