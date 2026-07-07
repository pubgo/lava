# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
