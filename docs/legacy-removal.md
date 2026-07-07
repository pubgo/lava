# Legacy 代码移除计划

本文档描述 Lava v2 发版前已清理的 legacy 表面，以及仍保留的兼容项。

## 原则

v2 尚未正式发布，发版前优先删除历史兼容层，避免首次发布即背负废弃 API。

## 已移除（v2 发版前）

### 服务与包

| 项 | 替代方案 |
|----|----------|
| `servers/grpcs` 包 | `servers/gatewayserver` |
| legacy gRPC 双注册（`grpc_passthrough: false`） | 固定 Mux-only 注册 + 原生 gRPC 透传 |
| `pkg/wsproxy` 包级 `MethodOverrideParam` / `TokenCookieName` | `WithMethodParamOverride` / `WithTokenCookieName` |

### 配置键

| 已删除键 | 替代键 |
|-----------|--------|
| `grpc_server` | `gateway_server` |
| `grpc_server.yaml` 组件 | `gateway_server.yaml` |
| `grpc_passthrough` | 已内置，无需配置 |

### API

| 项 | 替代 |
|----|------|
| `HasLocalIPddr`（拼写错误） | `HasLocalIPAddr` |

## 仍保留的兼容项

| 项 | 说明 |
|----|------|
| 根目录 `lava/` 类型别名 shim | 指向 `pkg/lava`，发版后视情况移除 |
| `grpc-server-info` debug vars 名 | 兼容 `lava curl` 与旧部署 |
| metric 名 `grpc_server_rpc_*` | 观测指标名，非配置/API |

## 配置示例

```yaml
gateway_server:
  enable_print_router: true
  websocket_port: 8081
  http: {}
  grpc: {}
```

## 相关 Issue

- #87 / #120：grpc_passthrough 与双注册
- #123：本计划

## 维护者检查清单

- [x] 新示例不再引用 `servers/grpcs`
- [x] HTTP 中间件链统一到 `servers/serverhttp.HandlerMiddleware`（#107）
- [x] `internal/configs/components/grpc_server.yaml` 已移除，统一 `gateway_server.yaml`
- [x] `gatewayserver.LoadConfig` 仅加载 `gateway_server`
- [x] 架构文档仅描述 `gateway_server`
- [x] 删除 `grpcs` 包
- [x] 删除 `grpc_server` YAML 键与 `grpc_passthrough: false` 双注册模式
- [x] 删除 `pkg/wsproxy` 废弃包级变量
- [x] 删除 `HasLocalIPddr` 拼写错误 API
