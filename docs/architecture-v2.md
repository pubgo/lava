# Lava 架构文档（v2）

> 本文档描述当前仓库的实际分层与关键运行流程，重点面向开发与维护。

## 1. 分层视图

```mermaid
flowchart TD
    subgraph Entry[入口层]
        MAIN[main.go]
        BUILDER[core/lavabuilder]
    end

    subgraph Cmd[命令层]
        CMDS[cmds/*]
    end

    subgraph Service[服务层]
        HTTPS[servers/https]
        GRPCS[servers/grpcs]
        TUNNEL[core/tunnel]
    end

    subgraph Core[核心能力层]
        SUP[core/supervisor]
        DBG[core/debug]
        LOG[core/logging]
        MET[core/metrics]
        TRC[core/tracing]
        SCHED[core/scheduler]
        DISC[core/discovery]
    end

    subgraph Infra[基础设施层]
        GW[pkg/gateway]
        HTTPU[pkg/httputil]
        GRPCU[pkg/grpcutil]
        NETU[pkg/netutil]
    end

    MAIN --> CMDS
    BUILDER --> CMDS
    CMDS --> HTTPS
    CMDS --> GRPCS
    CMDS --> TUNNEL
    HTTPS --> SUP
    GRPCS --> SUP
    TUNNEL --> SUP
    HTTPS --> DBG
    GRPCS --> DBG
    HTTPS --> LOG
    GRPCS --> LOG
    HTTPS --> MET
    GRPCS --> MET
    HTTPS --> TRC
    GRPCS --> TRC
    GRPCS --> GW
    HTTPS --> HTTPU
    GRPCS --> GRPCU
    TUNNEL --> NETU
    SCHED --> SUP
    DISC --> SUP
```

## 2. 启动流程

Lava 当前存在两套常见入口：

1. `main.go`：偏工具化 CLI（已接入 `watch/curl/tunnel/fileserver/devproxy`）
2. `core/lavabuilder.Run`：偏 DI 装配入口（注入 `version/health/dep/grpc/http/cron/tunnel` 等命令）

```mermaid
sequenceDiagram
    participant User as 用户
    participant Main as main.go
    participant Cmd as cmds/*
    participant Sup as supervisor.Manager
    participant Svc as 服务(https/grpcs/tunnel)

    User->>Main: 执行 lava <command>
    Main->>Cmd: 解析并进入命令处理
    Cmd->>Sup: 构建/获取 Manager
    Sup->>Svc: Add + Run
    Svc-->>Sup: 生命周期与状态上报
    Sup-->>User: 运行日志/错误输出
```

## 3. HTTP / gRPC 请求路径

### 3.1 HTTP 服务器（`servers/https`）

```mermaid
sequenceDiagram
    participant Client as HTTP Client
    participant Fiber as Fiber App
    participant Mid as Middlewares
    participant Router as HttpRouter
    participant Handler as 业务处理器

    Client->>Fiber: HTTP Request
    Fiber->>Mid: serviceinfo/metric/accesslog/recovery
    Mid->>Router: 路由匹配（Prefix）
    Router->>Handler: 处理业务
    Handler-->>Client: HTTP Response
```

### 3.2 gRPC 服务器（`servers/grpcs`）

```mermaid
sequenceDiagram
    participant Client as gRPC Client
    participant GrpcSrv as grpc.Server
    participant Intc as Unary/Stream Interceptor
    participant Service as GrpcRouter
    participant Gateway as pkg/gateway
    participant HttpClient as HTTP Client

    Client->>GrpcSrv: gRPC 调用
    GrpcSrv->>Intc: 中间件链
    Intc->>Service: 业务处理
    Service-->>Client: gRPC 响应

    HttpClient->>Gateway: HTTP/JSON
    Gateway->>Intc: 转换并复用拦截链
    Intc->>Service: 调用 gRPC 服务
    Service-->>HttpClient: JSON 响应
```

## 4. Tunnel 反向连接流程

`core/tunnel` 使用 Agent 主动连接 Gateway 的模式，适合内网服务暴露与远程调试。

```mermaid
sequenceDiagram
    participant Agent as Tunnel Agent
    participant Gw as Tunnel Gateway
    participant Ext as External Client
    participant Local as Local Service

    Agent->>Gw: Connect (:7007)
    Agent->>Gw: Register(service/endpoints)
    Ext->>Gw: HTTP/gRPC/Debug 请求
    Gw->>Agent: 转发请求
    Agent->>Local: 本地调用
    Local-->>Agent: 响应
    Agent-->>Gw: 返回
    Gw-->>Ext: 输出响应
```

## 5. 目录与模块锚点

- 入口：`main.go`、`core/lavabuilder/builder.go`
- 服务：`servers/https`、`servers/grpcs`
- 核心：`core/supervisor`、`core/debug`、`core/tunnel`
- 组件：`clients/*`、`pkg/*`、`lava/*`

详细模块信息见：`docs/modules/README.md`。
