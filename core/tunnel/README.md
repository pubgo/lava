# Tunnel - 服务注册监控网关

服务注册监控网关模块，采用**反向连接**架构（类似 ngrok/frp）。
服务节点上的 Agent **主动连接**到 Gateway，注册自己的服务，Gateway 被动接受连接并对外暴露这些服务。

## 核心特点

- **反向连接**：服务主动连接网关，无需开放服务端口
- **内网穿透**：服务可以在内网/防火墙后，只要能出站连接 Gateway
- **服务聚合**：多个服务通过同一个 Gateway 对外暴露
- **远程调试**：通过 Gateway 访问服务的 debug 接口进行远程监控

## 目录结构

```
core/tunnel/
├── doc.go              # 包文档（详细 API 说明）
├── types.go            # 核心类型和接口定义
├── config.go           # 配置结构定义
├── config.yaml         # 配置示例文件
├── errors.go           # 错误定义
├── transport.go        # 传输层注册表和工厂
├── README.md           # 本文档
├── tunnelagent/        # Agent 实现包
│   ├── agent.go        # Agent 封装和 Builder
│   ├── impl.go         # Agent 实现
│   └── doc.go          # 包文档
├── tunnelgateway/      # Gateway 实现包
│   ├── gateway.go      # Gateway 封装和 Builder
│   ├── impl.go         # Gateway 实现
│   └── doc.go          # 包文档
├── yamux/              # yamux 传输协议实现（基于 TCP）
├── quic/               # QUIC 传输协议实现（基于 UDP）
├── http/               # HTTP CONNECT 传输协议实现
├── kcp/                # KCP 传输协议实现（基于 UDP）
└── example/            # 使用示例
```

## 架构图

```
                        外部请求
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Gateway (公网/DMZ)                             │
│                                                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │ HTTP :8080  │ │ gRPC :9090  │ │ Debug :6060 │  <- 对外端口  │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘               │
│         │               │               │                       │
│  ┌──────┴───────────────┴───────────────┴──────┐               │
│  │         Service Router (按服务名路由)        │               │
│  └──────────────────────┬──────────────────────┘               │
│                         │                                       │
│  ┌──────────────────────┴──────────────────────┐               │
│  │         Session Manager (管理连接)           │               │
│  │    service-a ──> Session1                    │               │
│  │    service-b ──> Session2                    │               │
│  └──────────────────────────────────────────────┘               │
│                         │                                       │
│              Listener :7000  <- 接受 Agent 连接                 │
└─────────────────────────────────────────────────────────────────┘
                           ▲
            ┌──────────────┼──────────────┐
            │              │              │
      ┌─────┴─────┐  ┌─────┴─────┐  ┌─────┴─────┐
      │  Tunnel   │  │  Tunnel   │  │  Tunnel   │   <- 主动出站连接
      │ Session 1 │  │ Session 2 │  │ Session 3 │      (yamux 多路复用)
      └─────┬─────┘  └─────┬─────┘  └─────┬─────┘
            │              │              │
┌───────────┴──┐  ┌────────┴───┐  ┌───────┴────┐
│ Service A    │  │ Service B  │  │ Service C  │   <- 内网服务节点
│ (内网)       │  │ (内网)     │  │ (内网)     │
│              │  │            │  │            │
│ ┌──────────┐ │  │ ┌────────┐ │  │ ┌────────┐ │
│ │Agent     │─┼──┼─│Agent   │─┼──┼─│Agent   │ │   <- 主动连接 Gateway
│ └────┬─────┘ │  │ └───┬────┘ │  │ └───┬────┘ │
│      │       │  │     │      │  │     │      │
│ ┌────┴─────┐ │  │ ┌───┴────┐ │  │ ┌───┴────┐ │
│ │本地服务  │ │  │ │本地服务│ │  │ │本地服务│ │
│ │HTTP/gRPC │ │  │ │HTTP    │ │  │ │Debug   │ │
│ │Debug     │ │  │ └────────┘ │  │ └────────┘ │
│ └──────────┘ │  └────────────┘  └────────────┘
└──────────────┘
```

## 工作流程

```
1. Agent 启动，主动连接 Gateway
   Agent ─────────────────────────────────────> Gateway:7000
                    TCP + yamux

2. Agent 注册服务信息
   Agent ──── [Register] {name, endpoints} ───> Gateway
   Agent <─── [Ack] ──────────────────────────── Gateway

3. Agent 保持心跳
   Agent ──── [Heartbeat] ────────────────────> Gateway (每30秒)

4. 外部请求到达 Gateway
   Client ──── HTTP Request ──────────────────> Gateway:8080

5. Gateway 通过已建立的 tunnel 转发请求
   Gateway ──── [HTTPRequest] ────────────────> Agent
                  (通过 yamux stream)

6. Agent 转发到本地服务
   Agent ──────────────────────────────────────> localhost:8080

7. 响应原路返回
   localhost:8080 ──> Agent ──> Gateway ──> Client
```

## 快速开始

### 1. 部署 Gateway（公网服务器）

Gateway 部署在公网可访问的服务器上，被动等待服务连接。

```go
import (
    "context"
    "github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
    _ "github.com/pubgo/lava/v2/core/tunnel/yamux" // 注册 yamux 传输
)

func main() {
    ctx := context.Background()
    
    // 方式一：直接创建
    gw := tunnelgateway.New(&tunnelgateway.Config{
        ListenAddr: ":7000",      // Agent 连接端口
        Transport:  "yamux",
        HTTPPort:   8080,          // 对外暴露的 HTTP 端口
        GRPCPort:   9090,          // 对外暴露的 gRPC 端口
        DebugPort:  6060,          // 对外暴露的 Debug 端口
    })
    
    // 方式二：使用 Builder 模式
    gw, err := tunnelgateway.NewBuilder().
        WithListenAddr(":7000").
        WithTransport("yamux").
        WithHTTPPort(8080).
        WithDebugPort(6060).
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    if err := gw.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer gw.Stop(ctx)
    
    // Gateway 现在等待 Agent 连接...
    // 当 Agent 连接并注册服务后，可以通过 Services() 查看
    for _, svc := range gw.Services() {
        fmt.Printf("已注册服务: %s v%s\n", svc.Name, svc.Version)
    }
}
```

### 2. 部署 Agent（内网服务节点）

Agent 部署在服务所在的机器上（可以是内网），主动连接到 Gateway。

```go
import (
    "context"
    "github.com/pubgo/lava/v2/core/tunnel"
    "github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
    _ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

func main() {
    ctx := context.Background()
    
    // 方式一：直接创建
    agent := tunnelagent.New(&tunnelagent.Config{
        GatewayAddr: "gateway.example.com:7000",
        Transport:   "yamux",
        ServiceName: "my-service",
        ServiceVersion: "1.0.0",
        Endpoints: []tunnel.EndpointConfig{
            {Type: "http", LocalAddr: "localhost:8080"},
            {Type: "debug", LocalAddr: "localhost:6060"},
        },
    })
    
    // 方式二：使用 Builder 模式
    agent, err := tunnelagent.NewBuilder().
        WithGatewayAddr("gateway.example.com:7000").
        WithServiceName("my-service").
        WithServiceVersion("1.0.0").
        AddEndpoint("http", "localhost:8080", "/api").
        AddEndpoint("debug", "localhost:6060", "/debug").
        WithReconnectInterval(5).
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    // 启动 Agent，会自动：
    // 1. 连接到 Gateway
    // 2. 注册服务
    // 3. 保持心跳
    // 4. 断线自动重连
    if err := agent.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer agent.Stop(ctx)
    
    // 之后外部可以通过 Gateway 访问本地服务：
    // http://gateway.example.com:8080/my-service/api  -> localhost:8080
    // gateway.example.com:9090 (gRPC)                 -> localhost:9090  
    // http://gateway.example.com:6060/my-service/debug -> localhost:6060
    
    // 检查状态
    if agent.Status() == tunnelagent.StatusConnected {
        fmt.Println("已连接到 Gateway")
    }
}
```

### 3. 运行示例

```bash
# 运行完整示例（Gateway + Agent + Backend）
go run ./core/tunnel/example/main.go

# 使用不同传输协议
go run ./core/tunnel/example/main.go -transport=quic
go run ./core/tunnel/example/main.go -transport=http
go run ./core/tunnel/example/main.go -transport=kcp
```

### 4. 集成调试接口

调试接口由 `core/tunnel/tunneldebug` 包提供，可以集成到服务的调试端点：

```go
import "github.com/pubgo/lava/v2/core/tunnel/tunneldebug"

// 设置 Gateway 实例
tunneldebug.SetGateway(gw.Inner())

// 设置 Agent 实例  
tunneldebug.SetAgent(agent.Inner())
```

## 使用场景

### 典型场景：内网服务暴露

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    开发者电脑    │     │   公网 Gateway   │     │   内网服务器    │
│                 │     │                 │     │                 │
│  浏览器/curl    │────>│  :8080 (HTTP)   │<────│  Agent          │
│                 │     │  :9090 (gRPC)   │     │  └─ 主动连接    │
│                 │     │  :6060 (Debug)  │     │                 │
│                 │     │                 │     │  本地服务       │
│                 │     │  :7000 (Agent)  │     │  └─ :8080       │
│                 │     │       ↑ 被动    │     │  └─ :9090       │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

### 应用场景

1. **远程调试**：通过 Gateway 访问内网服务的 pprof、metrics 等调试接口
2. **服务聚合**：多个内网服务通过同一个 Gateway 对外暴露
3. **安全访问**：内网服务无需开放端口，只需能出站连接 Gateway
4. **临时暴露**：开发测试时临时将本地服务暴露到公网

## 配置说明

### 网关配置

```yaml
tunnel:
  gateway:
    enabled: true
    listen_addr: ":7000"          # Agent 连接地址
    transport: yamux              # 传输协议
    http_port: 8080               # HTTP 代理端口
    grpc_port: 9090               # gRPC 代理端口
    debug_port: 6060              # Debug 代理端口
    heartbeat_interval: 30        # 心跳间隔(秒)
    heartbeat_timeout: 90         # 心跳超时(秒)
    health_check_interval: 30     # 健康检查间隔(秒)
    tls:
      enabled: false
      cert_file: ""
      key_file: ""
```

### 代理客户端配置

```yaml
tunnel:
  agent:
    enabled: true
    gateway_addr: "gateway.example.com:7000"
    transport: yamux
    service_name: my-service
    service_version: "1.0.0"
    metadata:
      env: production
    endpoints:
      - type: http
        local_addr: "localhost:8080"
        path: /api
      - type: grpc
        local_addr: "localhost:9090"
      - type: debug
        local_addr: "localhost:6060"
        path: /debug
    heartbeat_interval: 30        # 心跳间隔(秒)
    reconnect_interval: 5         # 重连间隔(秒)
    max_reconnect_attempts: 0     # 0=无限重试
```

## 核心接口

### Transport - 传输层接口

```go
type Transport interface {
    Name() string
    Dial(ctx context.Context, addr string) (Session, error)
    Listen(ctx context.Context, addr string) (Listener, error)
}
```

### Session - 会话接口

```go
type Session interface {
    io.Closer
    Open(ctx context.Context) (Stream, error)
    Accept() (Stream, error)
    IsClosed() bool
    NumStreams() int
    LocalAddr() net.Addr
    RemoteAddr() net.Addr
}
```

### Agent - 代理客户端接口

```go
type Agent interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Register(ctx context.Context, service *ServiceInfo) error
    Deregister(ctx context.Context, serviceName string) error
    Status() AgentStatus
}
```

### Gateway - 网关接口

```go
type Gateway interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Services() []*ServiceInfo
    GetService(name string) (*ServiceInfo, error)
    Status() GatewayStatus
    Forward(ctx context.Context, serviceName string, endpointType EndpointType, conn net.Conn) error
}
```

## 调试端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/tunnel/` | GET | 概览 HTML 页面 |
| `/tunnel/gateway` | GET | 网关状态 (JSON) |
| `/tunnel/gateway/services` | GET | 所有注册服务列表 |
| `/tunnel/gateway/services/:name` | GET | 指定服务详情 |
| `/tunnel/agent` | GET | 代理客户端状态 |

## 状态说明

### AgentStatus

| 值 | 说明 |
|----|------|
| `StatusDisconnected` | 未连接 |
| `StatusConnecting` | 连接中 |
| `StatusConnected` | 已连接 |
| `StatusReconnecting` | 重连中 |

### GatewayStatus

| 值 | 说明 |
|----|------|
| `GatewayStatusStopped` | 已停止 |
| `GatewayStatusStarting` | 启动中 |
| `GatewayStatusRunning` | 运行中 |
| `GatewayStatusStopping` | 停止中 |

## 传输协议

| 协议 | 常量 | 状态 | 说明 |
|------|------|------|------|
| yamux | `TransportYamux` | ✅ 已实现 | 基于 TCP 的多路复用 |
| QUIC | `TransportQUIC` | ✅ 已实现 | 基于 UDP 的多路复用，低延迟、0-RTT |
| HTTP | `TransportHTTP` | ✅ 已实现 | HTTP CONNECT 隧道，适用于代理穿透 |
| KCP | `TransportKCP` | ✅ 已实现 | 基于 UDP 的可靠传输，弱网优化 |

### 协议选择指南

| 场景 | 推荐协议 | 原因 |
|------|----------|------|
| 通用场景 | yamux | 稳定可靠，兼容性好 |
| 高延迟网络 | QUIC/KCP | 0-RTT 连接，快速恢复 |
| 弱网环境 | KCP | 激进重传策略，抗丢包 |
| 企业代理穿透 | HTTP | 兼容 HTTP 代理服务器 |
| 需要 TLS 1.3 | QUIC | 内置加密，更安全 |

### 协议特性对比

| 特性 | yamux | QUIC | HTTP | KCP |
|------|-------|------|------|-----|
| 传输层 | TCP | UDP | TCP | UDP |
| 多路复用 | ✅ | ✅ | ❌ | ✅ (smux) |
| 连接迁移 | ❌ | ✅ | ❌ | ❌ |
| 0-RTT | ❌ | ✅ | ❌ | ❌ |
| 内置加密 | ❌ | ✅ | ❌ | ❌ |
| 代理穿透 | ❌ | ❌ | ✅ | ❌ |
| 抗丢包 | 一般 | 好 | 一般 | 优秀 |
| CPU 占用 | 低 | 中 | 低 | 中 |

### 自定义传输协议

```go
func init() {
    tunnel.RegisterTransport("custom", func(opts *tunnel.TransportOptions) (tunnel.Transport, error) {
        return &customTransport{opts: opts}, nil
    })
}

type customTransport struct {
    opts *tunnel.TransportOptions
}

func (t *customTransport) Name() string { return "custom" }
func (t *customTransport) Dial(ctx context.Context, addr string) (tunnel.Session, error) { ... }
func (t *customTransport) Listen(ctx context.Context, addr string) (tunnel.Listener, error) { ... }
```

## 通信协议

Agent 和 Gateway 使用长度前缀的 JSON 消息通信：

```
┌────────────┬─────────────────────────────┐
│ Length (4B)│     JSON Message            │
│  uint32 BE │  { "type": 1, "service":... │
└────────────┴─────────────────────────────┘
```

### 消息类型

| 类型 | 值 | 说明 |
|------|-----|------|
| `MessageTypeRegister` | 1 | 服务注册 |
| `MessageTypeDeregister` | 2 | 服务注销 |
| `MessageTypeHeartbeat` | 3 | 心跳 |
| `MessageTypeHTTPRequest` | 9 | HTTP 请求转发 |
| `MessageTypeGRPCRequest` | 10 | gRPC 请求转发 |
| `MessageTypeDebugRequest` | 11 | Debug 请求转发 |

## 错误类型

```go
ErrSessionClosed         // 会话已关闭
ErrStreamClosed          // 流已关闭
ErrConnectionFailed      // 连接失败
ErrServiceNotFound       // 服务未找到
ErrServiceAlreadyExists  // 服务已存在
ErrInvalidMessage        // 无效消息
ErrTimeout               // 超时
ErrTransportNotSupported // 不支持的传输协议
ErrGatewayNotConnected   // 未连接到网关
ErrAgentNotRunning       // 代理客户端未运行
ErrAgentAlreadyRunning   // 代理客户端已运行
ErrGatewayAlreadyRunning // 网关已运行
```

## 注意事项

1. **必须导入传输协议**：使用前需要导入对应的传输协议实现包
   ```go
   // 根据需要导入一个或多个传输协议
   import _ "github.com/pubgo/lava/v2/core/tunnel/yamux" // TCP + 多路复用（推荐）
   import _ "github.com/pubgo/lava/v2/core/tunnel/quic"  // UDP + 低延迟
   import _ "github.com/pubgo/lava/v2/core/tunnel/http"  // HTTP CONNECT 穿透
   import _ "github.com/pubgo/lava/v2/core/tunnel/kcp"   // UDP + 弱网优化
   ```

2. **自动重连**：Agent 断线后会自动重连，可通过 `ReconnectInterval` 配置重连间隔

3. **端点地址格式**：`Endpoint.Address` 应为完整的 `host:port` 格式

4. **心跳机制**：超过 `HeartbeatTimeout` 未收到心跳，服务会被标记为离线

5. **多服务支持**：单个 Agent 可以注册多个服务

## 依赖

### 核心依赖
- `github.com/gofiber/fiber/v3` - Fiber Web 框架（调试接口）
- `github.com/pubgo/funk/v2/log` - 日志库

### 传输层依赖
- `github.com/libp2p/go-yamux/v5` - yamux 多路复用实现
- `github.com/quic-go/quic-go` - QUIC 协议实现
- `github.com/xtaci/kcp-go/v5` - KCP 协议实现
- `github.com/xtaci/smux` - smux 多路复用（KCP 使用）
