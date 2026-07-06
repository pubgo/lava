# Legacy 代码移除计划

本文档描述 Lava v2 中仍并行维护的 legacy 表面，以及建议的废弃与移除节奏。

## 原则

1. **v2.x**：保留兼容，启动时打印 `deprecation` 日志（如已支持）。
2. **v2 末版**：文档与示例全部迁移到新 API，CI 对 legacy 路径仅做冒烟。
3. **v3**：删除 legacy 包与配置键，不提供静默兼容。

## 时间表（建议）

| 阶段 | 版本/时间 | 动作 |
|------|-----------|------|
| 文档化 | 当前 | 本文件 + 各包 `Deprecated` 注释 |
| 警告期 | v2.1+ | 使用 legacy 配置/包时 `log.Warn` |
| 冻结 | v2.2+ | legacy 仅修安全 bug，不加功能 |
| 移除 | v3.0 | 删除下表所列项 |

## 待移除项

### 服务与包

| 项 | 位置 | 替代方案 |
|----|------|----------|
| `pkg/wsproxy` 废弃构造函数 | `pkg/wsproxy` | `pkg/gateway` WebSocket 前端 |
| legacy gRPC 双注册 | `gatewayserver` `grpc_passthrough: false` | 默认 `true`，仅在 Mux 注册 |

### 已移除（v2）

| 项 | 替代方案 |
|----|----------|
| `servers/grpcs` 包 | `servers/gatewayserver` + `grpc_passthrough: true` |

### 配置键（已移除）

| 已删除键 | 替代键 |
|-----------|--------|
| `grpc_server` | `gateway_server` |
| `grpc_server.yaml` 组件 | `gateway_server.yaml` |

迁移：YAML 顶层键统一为 `gateway_server`；`gatewayserver.LoadConfig` 不再解析 `grpc_server`。

### API

| 项 | 位置 | 替代 |
|----|------|------|
| `HasLocalIPddr`（拼写错误） | `pkg/netutil/ip.go` | `HasLocalIPAddr` |

## 配置迁移示例

```yaml
# 旧 (removed)
grpc_server:
  enable_print_router: true

# 新
gateway_server:
  enable_print_router: true
  grpc_passthrough: true
```

## 相关 Issue

- #87 / #120：grpc_passthrough 与双注册（已在 v2 默认 passthrough）
- #123：本计划

## 维护者检查清单

- [x] 新示例不再引用 `servers/grpcs`
- [x] HTTP 中间件链统一到 `servers/serverhttp.HandlerMiddleware`（#107）
- [x] `internal/configs/components/grpc_server.yaml` 已移除，统一 `gateway_server.yaml`
- [x] `gatewayserver.LoadConfig` 仅加载 `gateway_server`（不再解析 `grpc_server`）
- [x] 架构文档仅描述 `gateway_server`
- [ ] `task test` 不依赖 legacy 路径（或单独 `task test:legacy`）
- [x] 删除 `grpcs` 包（v2，不再等待 v3）
