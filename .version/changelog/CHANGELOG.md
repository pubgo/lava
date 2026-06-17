# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
