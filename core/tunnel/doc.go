// Package tunnel 提供服务代理网关功能
//
// # 概述
//
// tunnel 包实现了一个基于反向连接的服务代理网关系统（类似 ngrok/frp）。
// 服务节点上的 Agent **主动连接**到 Gateway，注册自己的服务端点，
// 然后 Gateway 对外暴露这些服务的 HTTP API、gRPC 和 debug 调试接口。
//
// 核心特点：
//   - 反向连接：服务主动连接网关，网关被动接受连接
//   - 内网穿透：服务可以在内网/防火墙后，只要能出站连接 Gateway
//   - 服务聚合：多个服务通过同一个 Gateway 对外暴露
//   - 远程调试：通过 Gateway 访问服务的 debug 接口进行远程监控
//
// # 目录结构
//
//	core/tunnel/
//	├── doc.go          - 包文档
//	├── types.go        - 核心类型和接口定义
//	├── config.go       - 配置结构定义
//	├── config.yaml     - 配置示例文件
//	├── errors.go       - 错误定义
//	├── transport.go    - 传输层注册表和工厂
//	├── agent.go        - Agent 实现 (运行在服务节点，主动连接 Gateway)
//	├── gateway.go      - Gateway 实现 (运行在公网，被动接受 Agent 连接)
//	├── builder.go      - 构建器模式 API
//	├── debug.go        - 调试接口
//	└── yamux/
//	    └── yamux.go    - yamux 传输协议实现
//
// # 架构设计
//
// 与传统网关（如 Nginx）不同，本系统采用反向连接架构：
//
//   - 传统架构：客户端 -> 网关 -> 后端服务（网关主动连接后端）
//
//   - 本系统：  后端服务(Agent) -> 网关(Gateway)（服务主动连接网关）
//
//     外部请求
//     │
//     ▼
//     ┌─────────────────────────────────────────────────────────────────┐
//     │                   Gateway (公网/DMZ)                             │
//     │                                                                  │
//     │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
//     │  │ HTTP :8080  │ │ gRPC :9090  │ │ Debug :6060 │  <- 对外端口  │
//     │  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘               │
//     │         │               │               │                       │
//     │  ┌──────┴───────────────┴───────────────┴──────┐               │
//     │  │         Service Router (按服务名路由)        │               │
//     │  └──────────────────────┬──────────────────────┘               │
//     │                         │                                       │
//     │  ┌──────────────────────┴──────────────────────┐               │
//     │  │         Session Manager (管理连接)           │               │
//     │  │    service-a ──> Session1                    │               │
//     │  │    service-b ──> Session2                    │               │
//     │  └──────────────────────────────────────────────┘               │
//     │                         │                                       │
//     │              Listener :7000  <- 接受 Agent 连接 (被动)         │
//     └─────────────────────────────────────────────────────────────────┘
//     ▲
//     ┌──────────────┼──────────────┐
//     │              │              │
//     ┌─────┴─────┐  ┌─────┴─────┐  ┌─────┴─────┐
//     │  Tunnel   │  │  Tunnel   │  │  Tunnel   │   <- 主动出站连接
//     │ Session 1 │  │ Session 2 │  │ Session 3 │      (yamux 多路复用)
//     └─────┬─────┘  └─────┬─────┘  └─────┬─────┘
//     │              │              │
//     ┌───────────┴──┐  ┌────────┴───┐  ┌───────┴────┐
//     │ Service A    │  │ Service B  │  │ Service C  │   <- 内网服务节点
//     │ (内网)       │  │ (内网)     │  │ (内网)     │
//     │              │  │            │  │            │
//     │ ┌──────────┐ │  │ ┌────────┐ │  │ ┌────────┐ │
//     │ │Agent     │─┼──┼─│Agent   │─┼──┼─│Agent   │ │   <- 主动连接 Gateway
//     │ └────┬─────┘ │  │ └───┬────┘ │  │ └───┬────┘ │
//     │      │       │  │     │      │  │     │      │
//     │ ┌────┴─────┐ │  │ ┌───┴────┐ │  │ ┌───┴────┐ │
//     │ │本地服务  │ │  │ │本地服务│ │  │ │本地服务│ │
//     │ │HTTP/gRPC │ │  │ │HTTP    │ │  │ │Debug   │ │
//     │ │Debug     │ │  │ └────────┘ │  │ └────────┘ │
//     │ └──────────┘ │  └────────────┘  └────────────┘
//     └──────────────┘
//
// # 工作流程
//
//  1. Gateway 启动，监听 :7000 等待 Agent 连接
//  2. Agent 启动，主动连接到 Gateway:7000 (TCP + yamux)
//  3. Agent 发送 Register 消息，注册自己的服务信息和端点
//  4. Agent 定期发送 Heartbeat 保持连接
//  5. 外部请求到达 Gateway:8080
//  6. Gateway 根据服务名找到对应的 Agent Session
//  7. Gateway 通过 yamux stream 将请求转发给 Agent
//  8. Agent 将请求转发到本地服务 (localhost:8080)
//  9. 响应原路返回：本地服务 -> Agent -> Gateway -> 外部客户端
//
// ## 接口定义 (types.go)
//
//	Transport   - 传输层接口，负责建立连接
//	  ├── Name() string                                    - 协议名称
//	  ├── Dial(ctx, addr) (Session, error)                 - 客户端连接
//	  └── Listen(ctx, addr) (Listener, error)              - 服务端监听
//
//	Session     - 会话接口，支持多路复用
//	  ├── Open(ctx) (Stream, error)                        - 打开新流
//	  ├── Accept() (Stream, error)                         - 接受新流
//	  ├── Close() error                                    - 关闭会话
//	  ├── IsClosed() bool                                  - 是否已关闭
//	  ├── NumStreams() int                                 - 活跃流数量
//	  ├── LocalAddr() net.Addr                             - 本地地址
//	  └── RemoteAddr() net.Addr                            - 远程地址
//
//	Stream      - 流接口，单个请求/响应通道
//	  ├── io.ReadWriteCloser                               - 读写关闭
//	  ├── LocalAddr() / RemoteAddr()                       - 地址信息
//	  └── SetDeadline() / SetReadDeadline() / SetWriteDeadline()
//
//	Listener    - 监听器接口
//	  ├── Accept() (Session, error)                        - 接受新会话
//	  ├── Close() error                                    - 关闭监听
//	  └── Addr() net.Addr                                  - 监听地址
//
//	Agent       - 代理客户端接口
//	  ├── Start(ctx) error                                 - 启动
//	  ├── Stop(ctx) error                                  - 停止
//	  ├── Register(ctx, *ServiceInfo) error                - 注册服务
//	  ├── Deregister(ctx, serviceName) error               - 注销服务
//	  └── Status() AgentStatus                             - 获取状态
//
//	Gateway     - 代理网关接口
//	  ├── Start(ctx) error                                 - 启动
//	  ├── Stop(ctx) error                                  - 停止
//	  ├── Services() []*ServiceInfo                        - 获取所有服务
//	  ├── GetService(name) (*ServiceInfo, error)           - 获取指定服务
//	  ├── Status() GatewayStatus                           - 获取状态
//	  └── Forward(ctx, name, type, conn) error             - 转发请求
//
// ## 数据结构
//
//	ServiceInfo - 服务信息
//	  ├── ID            string              - 服务唯一标识
//	  ├── Name          string              - 服务名称
//	  ├── Version       string              - 服务版本
//	  ├── Metadata      map[string]string   - 元数据
//	  ├── Endpoints     []Endpoint          - 端点列表
//	  ├── RegisterTime  time.Time           - 注册时间
//	  ├── LastHeartbeat time.Time           - 最后心跳
//	  └── Status        ServiceStatus       - 服务状态
//
//	Endpoint    - 服务端点
//	  ├── Type     EndpointType            - 类型: http/grpc/debug
//	  ├── Path     string                  - 路径
//	  ├── Port     int                     - 端口
//	  ├── Address  string                  - 完整地址
//	  └── Metadata map[string]string       - 端点元数据
//
//	Message     - 通信消息
//	  ├── Type    MessageType              - 消息类型
//	  ├── ID      string                   - 消息ID
//	  ├── Service *ServiceInfo             - 服务信息
//	  ├── Payload []byte                   - 负载数据
//	  └── Error   string                   - 错误信息
//
// ## 状态枚举
//
//	AgentStatus:
//	  - StatusDisconnected  (0) - 未连接
//	  - StatusConnecting    (1) - 连接中
//	  - StatusConnected     (2) - 已连接
//	  - StatusReconnecting  (3) - 重连中
//
//	GatewayStatus:
//	  - GatewayStatusStopped   (0) - 已停止
//	  - GatewayStatusStarting  (1) - 启动中
//	  - GatewayStatusRunning   (2) - 运行中
//	  - GatewayStatusStopping  (3) - 停止中
//
//	ServiceStatus:
//	  - ServiceStatusOnline    - 在线
//	  - ServiceStatusOffline   - 离线
//	  - ServiceStatusUnhealthy - 不健康
//
//	EndpointType:
//	  - EndpointTypeHTTP  ("http")  - HTTP 端点
//	  - EndpointTypeGRPC  ("grpc")  - gRPC 端点
//	  - EndpointTypeDebug ("debug") - Debug 端点
//
//	MessageType:
//	  - MessageTypeRegister     (1)  - 服务注册
//	  - MessageTypeDeregister   (2)  - 服务注销
//	  - MessageTypeHeartbeat    (3)  - 心跳
//	  - MessageTypeRequest      (4)  - 请求
//	  - MessageTypeResponse     (5)  - 响应
//	  - MessageTypeStream       (6)  - 流数据
//	  - MessageTypeAck          (7)  - 确认
//	  - MessageTypeError        (8)  - 错误
//	  - MessageTypeHTTPRequest  (9)  - HTTP 请求转发
//	  - MessageTypeGRPCRequest  (10) - gRPC 请求转发
//	  - MessageTypeDebugRequest (11) - Debug 请求转发
//
// # 传输协议
//
// 支持的传输协议常量 (transport.go):
//
//	TransportYamux = "yamux"  - 基于 TCP 的多路复用 (已实现)
//	TransportQUIC  = "quic"   - 基于 UDP 的多路复用 (待实现)
//	TransportHTTP  = "http"   - HTTP CONNECT 隧道 (待实现)
//	TransportKCP   = "kcp"    - 基于 UDP 的可靠传输 (待实现)
//
// 注册自定义传输协议:
//
//	tunnel.RegisterTransport("custom", func(opts *TransportOptions) (Transport, error) {
//	    return &customTransport{opts: opts}, nil
//	})
//
// # 配置说明 (config.go)
//
// ## GatewayConfig - 网关配置
//
//	Enabled             bool               - 是否启用
//	ListenAddr          string             - 监听地址 (默认 :7000)
//	Transport           string             - 传输协议 (默认 yamux)
//	TransportOptions    *TransportOptions  - 传输层选项
//	HTTPPort            int                - HTTP 代理端口 (默认 8080)
//	GRPCPort            int                - gRPC 代理端口 (默认 9090)
//	DebugPort           int                - Debug 代理端口 (默认 6060)
//	HeartbeatInterval   int                - 心跳间隔秒数 (默认 30)
//	HeartbeatTimeout    int                - 心跳超时秒数 (默认 90)
//	HealthCheckInterval int                - 健康检查间隔 (默认 30)
//	TLS                 TLSConfig          - TLS 配置
//
// ## AgentConfig - 代理客户端配置
//
//	Enabled              bool               - 是否启用
//	GatewayAddr          string             - 网关地址
//	Transport            string             - 传输协议 (默认 yamux)
//	TransportOptions     *TransportOptions  - 传输层选项
//	ServiceID            string             - 服务ID (可选)
//	ServiceName          string             - 服务名称
//	ServiceVersion       string             - 服务版本
//	Metadata             map[string]string  - 元数据
//	Endpoints            []EndpointConfig   - 端点配置
//	HeartbeatInterval    int                - 心跳间隔秒数 (默认 30)
//	ReconnectInterval    int                - 重连间隔秒数 (默认 5)
//	MaxReconnectAttempts int                - 最大重连次数 (0=无限)
//	TLS                  TLSConfig          - TLS 配置
//
// ## TransportOptions - 传输层选项
//
//	EnableTLS              bool   - 启用 TLS
//	CertFile               string - 证书文件
//	KeyFile                string - 私钥文件
//	CAFile                 string - CA 证书
//	Insecure               bool   - 跳过证书验证
//	MaxStreams             int    - 最大流数量 (默认 256)
//	KeepAliveInterval      int    - 保活间隔秒数 (默认 30)
//	ConnectionWriteTimeout int    - 写超时秒数 (默认 10)
//	StreamOpenTimeout      int    - 流打开超时 (默认 30)
//
// # 使用示例
//
// ## 1. 部署 Gateway (公网服务器)
//
// Gateway 部署在公网可访问的服务器上，被动等待 Agent 连接。
//
//	import (
//	    "context"
//	    "github.com/pubgo/lava/v2/core/tunnel"
//	    _ "github.com/pubgo/lava/v2/core/tunnel/yamux" // 注册 yamux 传输
//	)
//
//	// 启动 Gateway，监听 :7000 接受 Agent 连接
//	gw, err := tunnel.NewGatewayBuilder().
//	    WithListenAddr(":7000").      // Agent 连接端口
//	    WithTransport("yamux").
//	    WithHTTPPort(8080).           // 对外暴露的 HTTP 端口
//	    WithGRPCPort(9090).           // 对外暴露的 gRPC 端口
//	    WithDebugPort(6060).          // 对外暴露的 Debug 端口
//	    WithHeartbeatInterval(30).
//	    WithHealthCheckInterval(30).
//	    Build()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := gw.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	// Gateway 现在等待 Agent 连接...
//
// ## 2. 部署 Agent (内网服务节点)
//
// Agent 部署在服务所在的机器上（可以是内网），主动连接到 Gateway。
//
//	// Agent 主动连接到 Gateway，注册本地服务
//	agent, err := tunnel.NewAgentBuilder().
//	    WithGatewayAddr("gateway.example.com:7000").  // Gateway 地址
//	    WithServiceName("my-service").
//	    WithServiceVersion("1.0.0").
//	    WithMetadata(map[string]string{"env": "prod"}).
//	    // 声明本地服务端点，Gateway 会代理这些端点
//	    AddEndpoint("http", "localhost:8080", "/api").    // 本地 HTTP 服务
//	    AddEndpoint("grpc", "localhost:9090", "").        // 本地 gRPC 服务
//	    AddEndpoint("debug", "localhost:6060", "/debug"). // 本地 Debug 端口
//	    WithHeartbeatInterval(30).
//	    WithReconnectInterval(5).  // 断线后 5 秒重连
//	    Build()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 启动 Agent，会自动：
//	// 1. 连接到 Gateway
//	// 2. 注册服务
//	// 3. 保持心跳
//	// 4. 断线自动重连
//	if err := agent.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// 之后外部可以通过 Gateway 访问本地服务：
//	// http://gateway.example.com:8080/my-service/api  -> localhost:8080
//	// gateway.example.com:9090 (gRPC)                 -> localhost:9090
//	// http://gateway.example.com:6060/my-service/debug -> localhost:6060
//
//	// 动态注册额外服务
//	agent.Register(ctx, &tunnel.ServiceInfo{
//	    Name: "another-service",
//	    Endpoints: []tunnel.Endpoint{
//	        {Type: tunnel.EndpointTypeHTTP, Address: "localhost:8081"},
//	    },
//	})
//
// ## 3. 集成调试接口
//
//	// 创建调试处理器
//	debugHandler := tunnel.NewDebugHandler()
//	debugHandler.SetGateway(gw)
//	debugHandler.SetAgent(agent)
//
//	// Fiber 路由集成
//	app := fiber.New()
//	debugHandler.FiberRoutes(app.Group("/debug"))
//
//	// 标准 HTTP 路由集成
//	mux := http.NewServeMux()
//	debugHandler.HTTPRoutes(mux)
//
// ## 4. TLS 配置
//
//	agent, _ := tunnel.NewAgentBuilder().
//	    WithGatewayAddr("gateway.example.com:7000").
//	    WithTLS(tunnel.TLSConfig{
//	        Enabled:  true,
//	        CertFile: "/path/to/client.crt",
//	        KeyFile:  "/path/to/client.key",
//	        CAFile:   "/path/to/ca.crt",
//	    }).
//	    Build()
//
// # 通信协议
//
// Agent 和 Gateway 之间使用长度前缀的 JSON 消息进行通信：
//
//	┌────────────┬─────────────────────────────┐
//	│ Length (4B)│     JSON Message            │
//	│  uint32 BE │  { "type": 1, "service":... │
//	└────────────┴─────────────────────────────┘
//
// 消息流程：
//
//	Agent                                    Gateway
//	  │                                         │
//	  │──── [Register] ServiceInfo ────────────>│  服务注册
//	  │                                         │
//	  │<─── [Ack] ─────────────────────────────│  确认
//	  │                                         │
//	  │──── [Heartbeat] ───────────────────────>│  心跳
//	  │                                         │
//	  │<─── [HTTPRequest] ServiceInfo ─────────│  请求转发
//	  │                                         │
//	  │──── [Stream] Response Data ────────────>│  响应数据
//	  │                                         │
//	  │──── [Deregister] ServiceName ──────────>│  服务注销
//	  │                                         │
//
// # 调试接口
//
// 可用的调试端点：
//
//	GET /tunnel/              - 概览 HTML 页面
//	GET /tunnel/gateway       - 网关状态 (JSON)
//	GET /tunnel/gateway/services       - 所有注册服务 (JSON)
//	GET /tunnel/gateway/services/:name - 指定服务详情 (JSON)
//	GET /tunnel/agent         - 代理客户端状态 (JSON)
//
// # 错误处理 (errors.go)
//
//	ErrSessionClosed         - 会话已关闭
//	ErrStreamClosed          - 流已关闭
//	ErrConnectionFailed      - 连接失败
//	ErrServiceNotFound       - 服务未找到
//	ErrServiceAlreadyExists  - 服务已存在
//	ErrInvalidMessage        - 无效消息
//	ErrTimeout               - 超时
//	ErrTransportNotSupported - 不支持的传输协议
//	ErrGatewayNotConnected   - 未连接到网关
//	ErrAgentNotRunning       - 代理客户端未运行
//	ErrAgentAlreadyRunning   - 代理客户端已运行
//	ErrGatewayAlreadyRunning - 网关已运行
//
// # 注意事项
//
//  1. 必须导入传输协议实现包以注册协议:
//     import _ "github.com/pubgo/lava/v2/core/tunnel/yamux"
//
//  2. Agent 会自动重连，但需要确保网关地址可达
//
//  3. 服务端点的 Address 字段应为完整的 host:port 格式
//
//  4. 心跳超时后服务会被标记为离线并从网关移除
//
//  5. 支持同一 Agent 注册多个服务
//
//  6. Gateway 的 Forward 方法用于将外部请求转发到对应服务
package tunnel
