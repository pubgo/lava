# Tunnel Gateway Example

这是一个独立运行的 Tunnel Gateway 服务示例。

## 架构说明

```
                     外部请求
                        │
                        ▼
┌───────────────────────────────────────────────────┐
│              Tunnel Gateway (本示例)               │
│                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────┐ │
│  │ HTTP :8888  │  │ gRPC :9999  │  │Debug :6066│ │
│  └──────┬──────┘  └──────┬──────┘  └─────┬─────┘ │
│         └────────────────┼───────────────┘       │
│                          │                       │
│            Tunnel Listener :7000                 │
└──────────────────────────┬───────────────────────┘
                           │
            ┌──────────────┼──────────────┐
            │              │              │
     ┌──────┴──────┐ ┌─────┴─────┐ ┌──────┴──────┐
     │ Scheduler   │ │ Service B │ │ Service C   │
     │ (Agent)     │ │ (Agent)   │ │ (Agent)     │
     │ :8080/:6060 │ │           │ │             │
     └─────────────┘ └───────────┘ └─────────────┘
            内网服务节点（主动连接 Gateway）
```

## 运行

### 1. 启动 Gateway

```bash
cd internal/examples/tunnel
go run . scheduler -c ../../configs/tunnel.yaml
```

Gateway 将监听以下端口：
- `:7007` - 接受 Agent 连接
- `:8888` - HTTP 代理（对外暴露服务）
- `:9999` - gRPC 代理
- `:6066` - Debug 代理

### 2. 启动带 Agent 的 Scheduler 服务

```bash
# 在另一个终端
cd internal/examples/scheduler
TUNNEL_GATEWAY_ADDR=localhost:7007 go run . scheduler -c ../../configs/scheduler.yaml
```

Scheduler 服务会：
1. 启动本地 HTTP 服务（`:8080`）
2. 启动本地 Debug 服务（`:6060`）
3. 通过 Agent 连接到 Gateway，注册自己

### 3. 通过 Gateway 访问服务

```bash
# 查看已注册的服务列表
curl http://localhost:8888/

# 访问 scheduler 服务的接口
curl http://localhost:8888/scheduler/api/v1/jobs

# 访问 scheduler 服务的 debug 接口
curl http://localhost:6066/scheduler/debug/pprof/
```

## 配置

配置文件位于 `internal/configs/` 目录下，复用项目统一的配置结构：

```
internal/configs/
├── tunnel.yaml              # Tunnel 主配置
├── scheduler.yaml           # Scheduler 主配置
├── components/
│   ├── tunnel.yaml          # Tunnel Gateway 组件配置
│   ├── http_server.yaml     # HTTP 服务配置
│   ├── logger.yaml          # 日志配置
│   └── metric.yaml          # 指标配置
└── envs/
    └── .env                 # 环境变量
```

### Gateway 配置 (components/tunnel.yaml)

```yaml
tunnel:
  listen_addr: ":7007"    # Agent 连接地址
  http_port: 8888         # HTTP 代理端口
  grpc_port: 9999         # gRPC 代理端口
  debug_port: 6066        # Debug 代理端口
```

### Agent 配置（环境变量）

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `TUNNEL_GATEWAY_ADDR` | Gateway 地址 | `localhost:7000` |
| `HTTP_ADDR` | 本地 HTTP 服务地址 | `localhost:8080` |
| `DEBUG_ADDR` | 本地 Debug 服务地址 | `localhost:6060` |

## 使用场景

1. **内网服务暴露**：服务在内网/防火墙后，通过 Agent 主动连接 Gateway 暴露到公网
2. **服务聚合**：多个微服务通过同一个 Gateway 统一入口
3. **远程调试**：通过 Gateway 访问内网服务的 pprof/debug 接口
4. **零配置部署**：服务只需知道 Gateway 地址，无需开放端口
