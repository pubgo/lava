# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.3.0] - 2026-09-29

本版本统一 gateway 分发层与 RPC middleware 契约,新增浏览器/Node 侧 JS SDK,**包含两项 breaking changes**(RPC middleware 边界、RPC 指标名)。

### Added

- **pkg/gateway**: 统一分发层(dispatcher)与前端流契约,新增 Backend 拦截器层与 gzip 压缩支持(带解压膨胀边界防护)
- **sdk/js/gateway-client**: `@pubgo/lava-gateway-client`,浏览器/Node 侧 HTTP/JSON、JSON-RPC 与 gRPC-Web 客户端(`createHttpJsonClient` / `createJsonRpcTransport` / `createGrpcWebTransport`)
- **internal/examples/grpcwebsocket/verify**: WebSocket 端到端验证工具;grpcweb / grpcwebsocket 示例重写并补 README
- **测试**: `pkg/gateway`(HTTP/gRPC-Web/WS/流式)、`core/tunnel`、`pkg/middleware`、`pkg/zrpc`、`pkg/netutil` 等覆盖加深

### Changed

- **BREAKING — RPC middleware 边界**:`UseRPCMiddleware` 只包装 `Mux.Dispatch` / `DispatchFrontend`(HTTP/WS/gRPC-Web 等前端路径),不再覆盖直接把 Mux 当 `grpc.ClientConnInterface` 用的调用(`pb.NewXxxClient(mux)`)。这类调用改挂 `UseBackendUnaryInterceptor` / `UseBackendStreamInterceptor`,或走前端 / `DispatchFrontend`。原因:`Invoke`/`NewStream` 在流开始时就返回,包住它们会提前释放请求 `timeout` 的 `CancelFunc`。详见 `pkg/gateway/docs/usage.md`
- **BREAKING — metrics**: RPC 指标名统一为 `lava_rpc_total` / `lava_rpc_failed_total` / `lava_rpc_handling_seconds`,客户端与服务端共用一组,用 `side`/`kind`/`service`/`method`/`stream`/`proto`(失败另有 `code`)区分;`gateway_server_*` 已删除,抓取端与看板需同步改
- **gatewayserver**: 中间件链重构为 `rpc_middleware.go`(单链覆盖 unary + 全部流式模式,本地与代理后端共用)
- **deps**: Fiber v3.4.0、OpenTelemetry v1.44.0、gRPC v1.81.1、prometheus/client_golang v1.24.1、golang.org/x/net v0.57.0
- **CI**: lint-test 扩展;Pages 构建包含 `pkg/gateway/docs` 嵌套文档

### Fixed

- **servers**: 优雅关闭拿到的是已取消的 context,HTTP/WS 排空窗口实际为 0;现改为独立 5s drain context,gRPC `GracefulStop` 与 HTTP/WS 排空并行并带超时兜底(`servers/gatewayserver`、`servers/https`)
- **pkg/gateway**: WebSocket close reason 按字节截断可能产生非法 UTF-8 并破坏 `{"grpcStatus":...}` JSON;改为 rune-safe 收缩 `grpcMessage` 后重新序列化
- **pkg/zrpc**: handler context 未附加 gRPC incoming metadata,NATS 调用方的 header(auth/trace/req-id)在 gateway 桥接链路被静默丢弃;现转换为 incoming metadata(过滤传输内部键)
- **pkg/gateway**: HTTP 流式帧长度上限、gRPC-Web trailer 键名与默认状态等帧协议加固;JS SDK inflate 边界同步收紧

### Documentation

- **pkg/gateway/docs**: 新增/重写 usage、architecture、design-evolution、grpcnative、grpcweb、internals、websocket 文档;README 重写,metrics 章节与代码对齐

## [2.2.0] - 2026-07-07

本版本是 v2 分支的大规模架构收敛与发版前清理，**包含多项 breaking changes**。升级前请阅读下方「Removed」与 `docs/legacy-removal.md`。

### Added

- **gatewayserver**: 统一对外多协议网关宿主（HTTP/REST、gRPC-Web、WebSocket、原生 gRPC 透传）
- **pkg/zrpcbridge**: 将 NATS/zrpc 订阅桥接到 `gateway.Mux`（替代 `Mux.RegisterZrpc`）
- **pkg/lava**: 公共契约迁至 `pkg/lava`（Middleware、Router、Request/Response）
- **pkg/lavacontexts**: 请求级 context 工具迁至 `pkg/`
- **servers/serverhttp**: 共享 HTTP 中间件链 `HandlerMiddleware`
- **core/p2p**: ICE 信令、TURN 凭证、QUIC 传输与 tunnel 集成
- **tunnel**: 重连、指标、TLS 合并、代理鉴权与 gateway 限流
- **deploy/traefik**: HTTP/3 + TLS 边缘部署示例
- **测试**: `curlcmd`、`grpcc`、`wsproxy`、`pkg/middleware`、`gatewayserver`、`https` 等单元测试加深

### Changed

- **配置**: Gateway YAML 键统一为 `gateway_server`（见 `internal/configs/components/gateway_server.yaml`）
- **gRPC**: 默认 Mux-only 注册 + 原生 gRPC 透传（不再需要 `grpc_passthrough` 配置）
- **CLI**: `lava grpc` 经 `gatewayserver.LoadConfig` 加载配置
- **debug**: `/debug/vars` 路由信息键名为 `gateway-server-info`
- **metrics**: RPC 指标名改为 `gateway_server_rpc_*`
- **tracing**: `tracing` 与 `tracingbuilder` 合并为单包
- **supervisor**: Manager 拆分为聚焦文件；服务启动路径统一
- **deps**: Fiber v3.3.0、quic-go v0.59.1、golang.org/x/net v0.55.0、OpenTelemetry SDK 等升级

### Removed

- **servers/grpcs** 废弃别名包 → 使用 `servers/gatewayserver`
- **grpc_server** YAML 配置键 → 使用 `gateway_server`
- **grpc_passthrough: false** 双注册模式 → 固定透传行为
- **lava/** 根目录类型别名 shim → 使用 `pkg/lava`
- **grpc-server-info** debug vars 别名 → 使用 `gateway-server-info`
- **HasLocalIPddr** 拼写错误 API → 使用 `HasLocalIPAddr`
- **wsproxy** 包级 `MethodOverrideParam` / `TokenCookieName` → 使用 `With*` Option
- **internal/configs/components/grpc_server.yaml** 组件文件

### Fixed

- Gateway、gRPC resolver、tunnel 认证与 body limit 等 P0–P2 可靠性问题
- debug 鉴权、CORS、配置默认值等安全与健壮性修复
- supervisor 生命周期与服务命令集成
- zrpc stream ACK / CloseSend 错误回传

### Documentation

- 新增/更新 `docs/architecture-v2.md`、`docs/modules/servers.md`、`docs/legacy-removal.md`
- Gateway 部署文档统一 `gateway_server` 表述（`pkg/gateway/docs/`、`deploy/traefik/`）

## [2.1.0] - 2026-06-17

### Added

- **zrpc**: protobuf RPC over NATS，支持 unary、server/client/bidi streaming
- **protoc-gen-zrpc-go**: 本地 protoc 插件与 `internal/zrpcgen` 代码生成器
- **clients/zrpcc** / **servers/zrpcs**: Lava 风格 zrpc 客户端与服务宿主
- **internal/examples/zrpcdemo**: 端到端 demo 与集成测试
- **release**: GoReleaser + git-cliff 发布 `protoc-gen-zrpc-go` 二进制（`v*.*.*` tag 触发）
- **ci**: `lint-test` 覆盖 `v2` 分支，测试 job 内置 NATS service

### Changed

- **deps**: `github.com/pubgo/dix/v2` → v2.0.1，`github.com/pubgo/funk/v2` → v2.0.4
- **lava**: `RequestKindZrpc` 请求类型标识

### Fixed

- **zrpc**: `OpenStream` handshake 与 stream 生命周期 context 分离
- **zrpc**: stream ACK / CloseSend 失败时向客户端返回 error frame
- **zrpcc**: 并发安全与 `Close()` 后禁止重连

### Documentation

- 新增 `docs/zrpc.md`，更新 modules 文档与可观测性/排障说明
