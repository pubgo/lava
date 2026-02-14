# Lava 功能模块说明文档

## 1. 客户端模块 (clients)

### 1.1 gRPC 客户端 (grpcc)

#### 模块概述
提供高性能、可靠的 gRPC 客户端，支持服务发现、负载均衡和中间件。

#### 核心功能
- **服务发现**：支持直连和基于注册中心的服务发现
- **负载均衡**：内置 P2C 负载均衡算法
- **中间件支持**：支持 gRPC 中间件
- **连接管理**：自动管理连接生命周期
- **健康检查**：内置服务健康检查

#### 主要 API

```go
// New 创建 gRPC 客户端
func New(cfg *grpccconfig.Cfg, p Params, middlewares ...lava.Middleware) Client

// Client gRPC 客户端接口
type Client interface {
    Get() result.Result[grpc.ClientConnInterface]
    Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error
    NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
    Healthy(ctx context.Context) error
}
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/clients/grpcc"
    "github.com/pubgo/lava/v2/clients/grpcc/grpccconfig"
)

// 创建 gRPC 客户端
client := grpcc.New(
    &grpccconfig.Cfg{
        Service: &grpccconfig.ServiceCfg{
            Name:   "user-service",
            Addr:   "localhost:9090",
            Scheme: "direct", // 直连模式
        },
    },
    grpcc.Params{
        Log:    logger,
        Metric: metric,
    },
    // 中间件...
)

// 获取 gRPC 连接
conn := client.Get().Expect("failed to get grpc client")

// 创建服务客户端
userClient := pb.NewUserServiceClient(conn)

// 调用服务
resp, err := userClient.GetUser(ctx, &pb.GetUserRequest{
    UserId: "123",
})
```

#### 配置选项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `Service.Name` | string | - | 服务名称 |
| `Service.Addr` | string | - | 服务地址 |
| `Service.Scheme` | string | "direct" | 服务发现方案 |
| `DialTimeout` | time.Duration | 10s | 连接超时 |
| `Keepalive.Time` | time.Duration | 60s | 保活时间 |
| `Keepalive.Timeout` | time.Duration | 20s | 保活超时 |

### 1.2 HTTP 客户端 (resty)

#### 模块概述
提供功能丰富的 HTTP 客户端，支持重试、超时、代理等特性。

#### 核心功能
- **自动重试**：支持失败自动重试
- **超时管理**：支持连接和请求超时
- **代理支持**：支持 HTTP 代理
- **Cookie 管理**：内置 Cookie 管理
- **中间件支持**：支持 HTTP 中间件

#### 主要 API

```go
// New 创建 HTTP 客户端
func New(cfg *Config, p Params, middlewares ...lava.Middleware) Client

// Client HTTP 客户端接口
type Client interface {
    R() *resty.Request
    Get(url string) (*resty.Response, error)
    Post(url string) (*resty.Response, error)
    Put(url string) (*resty.Response, error)
    Delete(url string) (*resty.Response, error)
}
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/clients/resty"
)

// 创建 HTTP 客户端
client := resty.New(
    &resty.Config{
        BaseURL: "http://api.example.com",
        Timeout: 10 * time.Second,
    },
    resty.Params{
        Log:    logger,
        Metric: metric,
    },
)

// 发送 GET 请求
resp, err := client.R().
    SetPathParam("id", "123").
    SetQueryParam("foo", "bar").
    SetHeader("Authorization", "Bearer token").
    Get("/users/{id}")

// 发送 POST 请求
resp, err := client.R().
    SetBody(map[string]interface{}{
        "name": "John",
        "email": "john@example.com",
    }).
    Post("/users")
```

## 2. 核心模块 (core)

### 2.1 服务管理器 (supervisor)

#### 模块概述
统一管理所有服务的生命周期，提供服务健康监控和自动故障恢复。

#### 核心功能
- **服务生命周期管理**：启动、停止、重启服务
- **健康状态监控**：定期检查服务健康状态
- **自动故障恢复**：服务失败时自动重启
- **服务状态追踪**：记录服务运行状态和统计信息
- **调试接口**：提供 HTTP API 用于管理和监控服务

#### 主要 API

```go
// NewManager 创建服务管理器
func NewManager(name string, lc lifecycle.Getter) *Manager

// Default 创建默认服务管理器
func Default(lc lifecycle.Getter) *Manager

// Register 注册服务
func (m *Manager) Register(name string, service Service, cfg ServiceConfig)

// StartService 启动服务
func (m *Manager) StartService(name string) error

// StopService 停止服务
func (m *Manager) StopService(name string) error

// RestartService 重启服务
func (m *Manager) RestartService(name string) error
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/core/supervisor"
)

// 创建服务管理器
mgr := supervisor.Default(lc)

// 注册服务
mgr.Register("http-server", httpService, supervisor.ServiceConfig{
    MaxRestarts:       3,
    RestartWindow:     time.Minute,
    RestartDelay:      time.Second,
    MaxConsecFailures: 3,
})

// 启动所有服务
mgr.StartAll(context.Background())

// 监控服务状态
for {
    services := mgr.Services()
    for _, svc := range services {
        fmt.Printf("Service %s: %s\n", svc.Name, svc.Status)
    }
    time.Sleep(5 * time.Second)
}
```

### 2.2 隧道系统 (tunnel)

#### 模块概述
提供基于反向连接的服务代理网关系统，支持内网穿透和服务聚合。

#### 核心功能
- **反向连接**：Agent 主动连接 Gateway，无需开放服务端口
- **内网穿透**：服务可以在内网/防火墙后，只要能出站连接 Gateway
- **服务聚合**：多个服务通过同一个 Gateway 对外暴露
- **远程调试**：通过 Gateway 访问服务的 debug 接口进行远程监控
- **多种传输协议**：支持 Yamux、QUIC、HTTP、KCP 等传输协议

#### 主要 API

**Agent API**：
```go
// New 创建 Agent
func New(cfg *Config, transport Transport) (Agent, error)

// Start 启动 Agent
func (a *Agent) Start(ctx context.Context) error

// Register 注册服务
func (a *Agent) Register(ctx context.Context, service *ServiceInfo) error
```

**Gateway API**：
```go
// New 创建 Gateway
func New(cfg *Config, transport Transport) (Gateway, error)

// Start 启动 Gateway
func (g *Gateway) Start(ctx context.Context) error

// Services 获取服务列表
func (g *Gateway) Services() []*ServiceInfo
```

#### 使用示例

**启动 Gateway**：
```go
import (
    "github.com/pubgo/lava/v2/core/tunnel"
    "github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
    _ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

// 创建 Gateway
cfg := &tunnelgateway.Config{
    ListenAddr: ":7000",
    HTTPPort:   8080,
    GRPCPort:   9090,
    DebugPort:  6060,
    Transport:  "yamux",
}

gw := tunnelgateway.New(cfg)

// 启动 Gateway
ctx := context.Background()
if err := gw.Start(ctx); err != nil {
    log.Fatal(err)
}

// 等待信号
<-ctx.Done()
```

**启动 Agent**：
```go
import (
    "github.com/pubgo/lava/v2/core/tunnel"
    "github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
    _ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

// 创建 Agent
cfg := &tunnelagent.Config{
    GatewayAddr:    "gateway:7000",
    ServiceName:    "my-service",
    ServiceVersion: "1.0.0",
    Transport:      "yamux",
    Endpoints: []tunnel.EndpointConfig{
        {
            Type:      "http",
            LocalAddr: "localhost:8080",
            Path:      "/api",
        },
        {
            Type:      "debug",
            LocalAddr: "localhost:6060",
            Path:      "/debug",
        },
    },
}

agent := tunnelagent.New(cfg)

// 启动 Agent
ctx := context.Background()
if err := agent.Start(ctx); err != nil {
    log.Fatal(err)
}

// 等待信号
<-ctx.Done()
```

### 2.3 调度器 (scheduler)

#### 模块概述
基于 cron 表达式的任务调度系统，支持定时任务的管理和执行。

#### 核心功能
- **cron 表达式支持**：支持标准 cron 表达式
- **任务管理**：动态添加、删除和查询任务
- **任务执行**：并发执行任务，支持任务超时控制
- **执行历史**：记录任务执行状态和结果
- **调试接口**：提供 HTTP API 用于管理和监控任务

#### 主要 API

```go
// New 创建调度器
func New(cfg *Config) *Scheduler

// AddJob 添加任务
func (s *Scheduler) AddJob(job Job) error

// RemoveJob 删除任务
func (s *Scheduler) RemoveJob(id string) error

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) error

// Stop 停止调度器
func (s *Scheduler) Stop(ctx context.Context) error
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/core/scheduler"
)

// 定义任务
func backupHandler(ctx context.Context) error {
    // 执行备份逻辑
    log.Println("Running backup job...")
    return nil
}

// 创建调度器
cfg := &scheduler.Config{
    Jobs: []scheduler.JobConfig{
        {
            Name:     "backup",
            Schedule: "0 0 * * *", // 每天凌晨执行
            Handler:  "backupHandler",
        },
        {
            Name:     "cleanup",
            Schedule: "0 */6 * * *", // 每6小时执行
            Handler:  "cleanupHandler",
        },
    },
}

sched := scheduler.New(cfg)

// 注册任务处理器
sched.RegisterHandler("backupHandler", backupHandler)
sched.RegisterHandler("cleanupHandler", cleanupHandler)

// 启动调度器
ctx := context.Background()
if err := sched.Start(ctx); err != nil {
    log.Fatal(err)
}

// 等待信号
<-ctx.Done()
```

### 2.4 日志系统 (logging)

#### 模块概述
提供统一的日志接口，支持多种日志输出和结构化日志。

#### 核心功能
- **多种日志输出**：支持 slog、stdlog、grpclog 等
- **结构化日志**：支持 JSON 等结构化日志格式
- **日志级别**：支持 DEBUG、INFO、ERROR 等多种日志级别
- **日志字段**：支持添加自定义日志字段
- **上下文日志**：支持从 context 中提取日志信息

#### 主要 API

```go
// New 创建日志记录器
func New(cfg *Config) log.Logger

// GetLogger 获取日志记录器
func GetLogger(name string) log.Logger

// WithName 添加日志名称
func (l log.Logger) WithName(name string) log.Logger

// WithField 添加日志字段
func (l log.Logger) WithField(key string, value interface{}) log.Logger
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/core/logging"
    "github.com/pubgo/lava/v2/core/logging/logbuilder"
)

// 创建日志记录器
log := logbuilder.New(logging.Config{
    Level:  "info",
    Format: "json",
    Output: "stdout",
})

// 记录不同级别的日志
log.Info().Str("user", "john").Int("age", 30).Msg("User info")
log.Debug().Str("path", "/api/users").Msg("Request received")
log.Error().Err(err).Msg("Error occurred")

// 使用命名日志记录器
userLog := log.WithName("user-service")
userLog.Info().Str("action", "create").Msg("User created")

// 从 context 中获取日志记录器
ctx := context.Background()
ctxLog := log.FromContext(ctx)
ctxLog.Info().Msg("Logging from context")
```

### 2.5 指标系统 (metrics)

#### 模块概述
提供统一的指标接口，支持 Prometheus 等多种指标导出。

#### 核心功能
- **多种指标类型**：支持 Counter、Gauge、Histogram 等
- **Prometheus 集成**：支持 Prometheus 指标导出
- **指标标签**：支持添加自定义指标标签
- **指标聚合**：支持指标聚合和计算
- **内置指标**：提供系统和服务的内置指标

#### 主要 API

```go
// New 创建指标记录器
func New(cfg *Config) metrics.Metric

// Counter 创建计数器
func (m metrics.Metric) Counter(name string, labels ...string) metrics.Counter

// Gauge 创建 gauge
func (m metrics.Metric) Gauge(name string, labels ...string) metrics.Gauge

// Histogram 创建直方图
func (m metrics.Metric) Histogram(name string, labels ...string) metrics.Histogram
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/core/metrics"
    "github.com/pubgo/lava/v2/core/metrics/metricbuilder"
)

// 创建指标记录器
metric := metricbuilder.New(metrics.Config{
    Enabled: true,
    Exporter: "prometheus",
    Port: 9091,
})

// 创建计数器
requestCounter := metric.Counter("http_requests_total", "method", "path", "status")

// 记录指标
requestCounter.Inc("GET", "/api/users", "200")

// 创建 gauge
temperatureGauge := metric.Gauge("system_temperature", "sensor")
temperatureGauge.Set(25.5, "cpu")

// 创建直方图
requestLatency := metric.Histogram("http_request_duration_seconds", "method")
requestLatency.Observe(0.1, "GET")
```

### 2.6 链路追踪 (tracing)

#### 模块概述
提供统一的链路追踪接口，支持 OpenTelemetry 标准。

#### 核心功能
- **OpenTelemetry 集成**：支持标准的 OpenTelemetry 规范
- **多种导出器**：支持 Jaeger、Zipkin 等多种导出器
- **上下文传递**：通过 context 传递追踪信息
- **span 管理**：创建和管理追踪 span
- ** baggage 支持**：支持 baggage 传递

#### 主要 API

```go
// New 创建追踪器
func New(cfg *Config) tracing.Tracer

// Start 开始追踪
func (t tracing.Tracer) Start(ctx context.Context, name string) (context.Context, tracing.Span)

// Inject 注入追踪信息
func (t tracing.Tracer) Inject(ctx context.Context, carrier interface{}) error

// Extract 提取追踪信息
func (t tracing.Tracer) Extract(ctx context.Context, carrier interface{}) (context.Context, error)
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/core/tracing"
    "github.com/pubgo/lava/v2/core/tracing/tracingbuilder"
)

// 创建追踪器
cfg := &tracing.Config{
    Enabled:   true,
    Exporter:  "jaeger",
    Endpoint:  "http://jaeger:14268/api/traces",
    ServiceName: "user-service",
}

tracer := tracingbuilder.New(cfg)

// 开始追踪
ctx, span := tracer.Start(context.Background(), "user-service:get-user")
span.SetTag("user_id", "123")

// 执行操作
user, err := getUser(ctx, "123")
if err != nil {
    span.SetTag("error", true)
    span.LogKV("error.message", err.Error())
}

// 结束追踪
span.Finish()

// 嵌套追踪
func getUser(ctx context.Context, id string) (*User, error) {
    ctx, span := tracer.Start(ctx, "get-user")
    defer span.Finish()
    
    // 执行数据库操作
    ctx, dbSpan := tracer.Start(ctx, "database:query")
    user, err := db.Query(ctx, "SELECT * FROM users WHERE id = ?", id)
    dbSpan.Finish()
    
    return user, err
}
```

### 2.7 编码解码 (encoding)

#### 模块概述
提供统一的编码解码接口，支持多种编码格式。

#### 核心功能
- **多种编码格式**：支持 JSON、Protobuf、MsgPack 等
- **统一接口**：提供统一的编码解码接口
- **类型安全**：支持类型安全的编码解码
- **性能优化**：针对不同编码格式的性能优化

#### 主要 API

```go
// GetCodec 获取编解码器
func GetCodec(name string) Codec

// Encode 编码
func (c Codec) Encode(v interface{}) ([]byte, error)

// Decode 解码
func (c Codec) Decode(data []byte, v interface{}) error
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/core/encoding"
)

// 获取 JSON 编解码器
jsonCodec := encoding.GetCodec("json")

// 编码数据
data, err := jsonCodec.Encode(map[string]interface{}{
    "name": "John",
    "age": 30,
})

// 解码数据
var result map[string]interface{}
err = jsonCodec.Decode(data, &result)

// 获取 Protobuf 编解码器
protoCodec := encoding.GetCodec("protobuf")

// 编码 Protobuf 消息
msg := &pb.User{
    Name: "John",
    Age: 30,
}
data, err = protoCodec.Encode(msg)

// 解码 Protobuf 消息
var user pb.User
err = protoCodec.Decode(data, &user)
```

### 2.8 服务发现 (discovery)

#### 模块概述
提供统一的服务发现接口，支持多种注册中心。

#### 核心功能
- **多种注册中心**：支持 Consul、Etcd、Kubernetes 等
- **服务注册**：注册服务实例到注册中心
- **服务发现**：从注册中心发现服务实例
- **健康检查**：定期检查服务实例健康状态
- **服务监控**：监控服务实例变化

#### 主要 API

```go
// New 创建服务发现客户端
func New(cfg *Config) discovery.Discovery

// Register 注册服务
func (d discovery.Discovery) Register(ctx context.Context, service *ServiceInfo) error

// Deregister 注销服务
func (d discovery.Discovery) Deregister(ctx context.Context, service *ServiceInfo) error

// Discover 发现服务
func (d discovery.Discovery) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)

// Watch 监控服务变化
func (d discovery.Discovery) Watch(ctx context.Context, serviceName string, callback func([]*ServiceInstance)) error
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/core/discovery"
)

// 创建服务发现客户端
cfg := &discovery.Config{
    Type:     "consul",
    Address:  "consul:8500",
    Interval: "30s",
}

disco := discovery.New(cfg)

// 注册服务
ctx := context.Background()
service := &discovery.ServiceInfo{
    Name:    "user-service",
    ID:      "user-service-1",
    Address: "192.168.1.100",
    Port:    8080,
    Tags:    []string{"v1", "production"},
    Check: &discovery.HealthCheck{
        HTTP:     "http://192.168.1.100:8080/health",
        Interval: "10s",
        Timeout:  "5s",
    },
}

if err := disco.Register(ctx, service); err != nil {
    log.Fatal(err)
}

// 发现服务
instances, err := disco.Discover(ctx, "user-service")
if err != nil {
    log.Fatal(err)
}

for _, instance := range instances {
    fmt.Printf("Service instance: %s:%d\n", instance.Address, instance.Port)
}

// 监控服务变化
disco.Watch(ctx, "user-service", func(instances []*discovery.ServiceInstance) {
    fmt.Printf("Service instances changed: %d instances\n", len(instances))
})
```

## 3. 服务器模块 (servers)

### 3.1 HTTP 服务器 (https)

#### 模块概述
基于 Fiber 框架的高性能 HTTP 服务器，支持路由管理和中间件。

#### 核心功能
- **高性能**：基于 Fiber 框架，性能优异
- **路由管理**：支持 RESTful 路由和路由分组
- **中间件链**：支持 HTTP 中间件
- **静态文件**：支持静态文件服务
- **WebSocket**：支持 WebSocket 连接
- **调试接口**：自动挂载调试 API

#### 主要 API

```go
// New 创建 HTTP 服务器
func New(params Params) supervisor.Service

// Params HTTP 服务器参数
type Params struct {
    Handlers    []lava.HttpRouter
    Middlewares []lava.Middleware
    M           metrics.Metric
    Log         log.Logger
    Cfg         *Config
}
```

#### 使用示例

```go
import (
    "github.com/gofiber/fiber/v3"
    "github.com/pubgo/lava/v2/lava"
    "github.com/pubgo/lava/v2/servers/https"
)

// 定义 HTTP 路由
type UserRouter struct{}

func (r *UserRouter) Prefix() string {
    return "/api/users"
}

func (r *UserRouter) Middlewares() []lava.Middleware {
    return nil
}

func (r *UserRouter) Router(router fiber.Router) {
    router.Get("/", r.GetUsers)
    router.Get("/:id", r.GetUser)
    router.Post("/", r.CreateUser)
    router.Put("/:id", r.UpdateUser)
    router.Delete("/:id", r.DeleteUser)
}

func (r *UserRouter) GetUsers(c *fiber.Ctx) error {
    return c.JSON([]map[string]interface{}{
        {"id": "1", "name": "John"},
        {"id": "2", "name": "Jane"},
    })
}

// 创建 HTTP 服务器
httpService := https.New(https.Params{
    Handlers: []lava.HttpRouter{
        &UserRouter{},
    },
    Middlewares: []lava.Middleware{
        // 中间件...
    },
    M:    metric,
    Log:  log,
    Cfg:  &https.Config{},
})

// 注册到服务管理器
mgr.Register("http-server", httpService, supervisor.ServiceConfig{})
```

### 3.2 gRPC 服务器 (grpcs)

#### 模块概述
基于标准 gRPC 库的 gRPC 服务器，支持服务注册和中间件。

#### 核心功能
- **标准 gRPC**：基于官方 gRPC 库
- **服务注册**：自动注册 gRPC 服务
- **中间件链**：支持 gRPC 中间件
- **gRPC Gateway**：集成 gRPC Gateway，支持 HTTP/JSON
- **调试接口**：自动挂载调试 API

#### 主要 API

```go
// New 创建 gRPC 服务器
func New(params Params) supervisor.Service

// Params gRPC 服务器参数
type Params struct {
    GrpcRouters     []lava.GrpcRouter
    HttpRouters     []lava.HttpRouter
    GrpcHttpRouters []lava.GrpcHttpRouter
    DixMiddlewares  []lava.Middleware
    Metric          metrics.Metric
    Log             log.Logger
    Conf            *Config
    Gw              []*gateway.Mux
}
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/lava"
    "github.com/pubgo/lava/v2/servers/grpcs"
    pb "your/proto/package"
)

// 实现 gRPC 服务
type UserService struct{}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    return &pb.User{
        Id:   req.UserId,
        Name: "John",
        Age:  30,
    }, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
    return &pb.User{
        Id:   "123",
        Name: req.Name,
        Age:  req.Age,
    }, nil
}

func (s *UserService) ServiceDesc() *grpc.ServiceDesc {
    return &pb.UserService_ServiceDesc
}

// 创建 gRPC 服务器
grpcService := grpcs.New(grpcs.Params{
    GrpcRouters: []lava.GrpcRouter{
        &UserService{},
    },
    HttpRouters: []lava.HttpRouter{
        // HTTP 路由...
    },
    DixMiddlewares: []lava.Middleware{
        // 中间件...
    },
    Metric: metric,
    Log:    log,
    Conf:   &grpcs.Config{},
})

// 注册到服务管理器
mgr.Register("grpc-server", grpcService, supervisor.ServiceConfig{})
```

## 4. 公共包 (pkg)

### 4.1 gRPC Gateway (gateway)

#### 模块概述
提供 HTTP/JSON 到 gRPC 的协议转换，支持 gRPC Web。

#### 核心功能
- **HTTP Rule 解析**：支持 Google API HTTP 注解
- **协议转换**：自动处理 HTTP/JSON 与 gRPC/Protobuf 之间的转换
- **gRPC Web 支持**：允许浏览器直接调用 gRPC 服务
- **服务注册**：支持本地服务和代理服务的注册
- **中间件支持**：提供 Unary 和 Stream 拦截器
- **错误映射**：自动将 gRPC 错误码映射为 HTTP 状态码

#### 主要 API

```go
// NewMux 创建 Gateway
func NewMux(opts ...Option) *Mux

// RegisterService 注册服务
func (m *Mux) RegisterService(desc *grpc.ServiceDesc, srv interface{})

// RegisterProxy 注册代理服务
func (m *Mux) RegisterProxy(desc *grpc.ServiceDesc, srv interface{}, client grpc.ClientConnInterface)

// Handler HTTP 处理器
func (m *Mux) Handler(c *fiber.Ctx) error
```

#### 使用示例

```go
import (
    "github.com/gofiber/fiber/v3"
    "github.com/pubgo/lava/v2/pkg/gateway"
    pb "your/proto/package"
)

// 创建 Gateway
mux := gateway.NewMux()

// 注册服务
mux.RegisterService(&pb.UserService_ServiceDesc, &UserService{})

// 创建 Fiber 应用
app := fiber.New()

// 挂载 Gateway
apiPrefix := "/api"
app.Group(apiPrefix, func(ctx *fiber.Ctx) error {
    return httputil.StripPrefix(apiPrefix, mux.Handler)(ctx)
})

// 启动服务器
app.Listen(":8080")
```

### 4.2 命令行工具 (cliutil)

#### 模块概述
提供命令行工具的公共功能，支持命令注册和参数解析。

#### 核心功能
- **命令注册**：支持注册子命令
- **参数解析**：支持命令行参数解析
- **帮助信息**：自动生成帮助信息
- **命令补全**：支持命令补全

#### 主要 API

```go
// NewCommand 创建命令
func NewCommand(use, short string, run func(*redant.Command) error) *redant.Command

// RegisterCommand 注册命令
func RegisterCommand(cmd *redant.Command)

// GetFlags 获取全局标志
func GetFlags() []redant.Flag
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/pkg/cliutil"
)

// 创建命令
cmd := cliutil.NewCommand(
    "hello",
    "Say hello",
    func(cmd *redant.Command) error {
        name := cmd.Flag("name").String()
        cmd.Println("Hello,", name)
        return nil
    },
)

// 添加标志
cmd.Flag("name", "Your name").Default("World").String()

// 注册命令
cliutil.RegisterCommand(cmd)

// 运行命令
app := &redant.Command{
    Use:   "myapp",
    Short: "My application",
    Children: []*redant.Command{cmd},
}

app.Run(os.Args)
```

### 4.3 HTTP 工具 (httputil)

#### 模块概述
提供 HTTP 相关的工具函数和中间件。

#### 核心功能
- **CORS**：支持跨域资源共享
- **请求处理**：提供请求解析和处理工具
- **响应处理**：提供响应格式化和发送工具
- **路径处理**：提供路径操作工具
- **中间件**：提供常用 HTTP 中间件

#### 主要 API

```go
// Cors 跨域中间件
func Cors() fiber.Handler

// StripPrefix 路径前缀处理
func StripPrefix(prefix string, h fiber.Handler) fiber.Handler

// DefaultCfg 默认配置
func DefaultCfg(cfg *Config) *Config

// Build 构建 Fiber 配置
func (c *Config) Build() result.Result[fiber.Config]
```

#### 使用示例

```go
import (
    "github.com/gofiber/fiber/v3"
    "github.com/pubgo/lava/v2/pkg/httputil"
)

// 创建 Fiber 应用
app := fiber.New(httputil.DefaultCfg(&httputil.Config{
    EnablePrintRouter: true,
}).Build().Unwrap())

// 添加 CORS 中间件
app.Use(httputil.Cors())

// 处理路径前缀
apiPrefix := "/api"
app.Group(apiPrefix, func(ctx *fiber.Ctx) error {
    return httputil.StripPrefix(apiPrefix, func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "Hello from API",
        })
    })(ctx)
})
```

### 4.4 网络工具 (netutil)

#### 模块概述
提供网络相关的工具函数。

#### 核心功能
- **地址处理**：提供地址解析和格式化工具
- **连接管理**：提供连接处理工具
- **网络检测**：提供网络状态检测工具
- **错误处理**：提供网络错误处理工具

#### 主要 API

```go
// IsErrServerClosed 检查是否是服务器关闭错误
func IsErrServerClosed(err error) bool

// ParseAddr 解析地址
func ParseAddr(addr string) (string, error)

// GetFreePort 获取空闲端口
func GetFreePort() (int, error)

// IsIPv4 检查是否是 IPv4 地址
func IsIPv4(ip net.IP) bool

// IsIPv6 检查是否是 IPv6 地址
func IsIPv6(ip net.IP) bool
```

#### 使用示例

```go
import (
    "github.com/pubgo/lava/v2/pkg/netutil"
)

// 检查服务器关闭错误
if err != nil && !netutil.IsErrServerClosed(err) {
    log.Fatal(err)
}

// 解析地址
addr, err := netutil.ParseAddr("localhost:8080")
if err != nil {
    log.Fatal(err)
}

// 获取空闲端口
port, err := netutil.GetFreePort()
if err != nil {
    log.Fatal(err)
}

// 检查 IP 类型
ip := net.ParseIP("192.168.1.1")
if netutil.IsIPv4(ip) {
    fmt.Println("IPv4 address")
}
```

## 5. 命令行工具 (cmds)

### 5.1 配置命令 (configcmd)

#### 模块概述
提供配置管理的命令行工具。

#### 核心功能
- **配置查看**：查看当前配置
- **配置编辑**：编辑配置文件
- **配置验证**：验证配置有效性
- **配置导出**：导出配置到文件

#### 使用示例

```bash
# 查看当前配置
lava config show

# 编辑配置
lava config edit

# 验证配置
lava config validate

# 导出配置
lava config export > config.yaml
```

### 5.2 依赖管理 (depcmd)

#### 模块概述
提供依赖管理的命令行工具。

#### 核心功能
- **依赖查看**：查看当前依赖
- **依赖更新**：更新依赖版本
- **依赖清理**：清理未使用的依赖
- **依赖锁定**：锁定依赖版本

#### 使用示例

```bash
# 查看依赖
lava dep list

# 更新依赖
lava dep update

# 清理依赖
lava dep prune

# 锁定依赖
lava dep lock
```

### 5.3 环境管理 (envcmd)

#### 模块概述
提供环境管理的命令行工具。

#### 核心功能
- **环境查看**：查看当前环境变量
- **环境设置**：设置环境变量
- **环境导出**：导出环境变量到文件
- **环境加载**：从文件加载环境变量

#### 使用示例

```bash
# 查看环境变量
lava env show

# 设置环境变量
lava env set HTTP_PORT=8080

# 导出环境变量
lava env export > .env

# 加载环境变量
lava env load < .env
```

### 5.4 健康检查 (healthcmd)

#### 模块概述
提供健康检查的命令行工具。

#### 核心功能
- **服务健康检查**：检查服务健康状态
- **依赖健康检查**：检查依赖服务健康状态
- **健康状态报告**：生成健康状态报告

#### 使用示例

```bash
# 检查服务健康状态
lava health check

# 检查依赖服务
lava health dependencies

# 生成健康报告
lava health report
```

### 5.5 HTTP 客户端工具 (curl)

#### 模块概述
提供 HTTP 客户端工具，支持按 operation 或路径调用 API。

#### 核心功能
- **按 operation 调用**：自动匹配方法和路径
- **按路径调用**：支持显式路径和方法
- **参数传递**：支持 Header、Query、Path 参数
- **Token 管理**：支持登录和 Token 管理
- **路由发现**：自动发现网关路由

#### 使用示例

```bash
# 列出路由
lava curl --list

# 登录
lava curl login -t "YOUR_TOKEN"

# 按 operation 调用
lava curl UserService/GetUser -d '{"userId":"123"}'

# 按路径调用
lava curl --path /api/users/123

# 携带参数
lava curl UserService/GetUser -H "X-Req-Id=abc" -Q verbose=true -P id=123
```

### 5.6 调度器命令 (schedulercmd)

#### 模块概述
提供调度器管理的命令行工具。

#### 核心功能
- **任务管理**：查看、添加、删除任务
- **任务执行**：手动执行任务
- **任务状态**：查看任务执行状态
- **执行历史**：查看任务执行历史

#### 使用示例

```bash
# 查看任务列表
lava scheduler list

# 添加任务
lava scheduler add --name backup --schedule "0 0 * * *" --handler backupHandler

# 删除任务
lava scheduler remove --name backup

# 执行任务
lava scheduler run --name backup

# 查看执行历史
lava scheduler history
```

### 5.7 隧道命令 (tunnelcmd)

#### 模块概述
提供隧道管理的命令行工具。

#### 核心功能
- **Agent 管理**：启动、停止 Agent
- **Gateway 管理**：启动、停止 Gateway
- **服务管理**：查看、注册、注销服务
- **状态查看**：查看隧道状态

#### 使用示例

```bash
# 启动 Gateway
lava tunnel gateway --listen :7000 --http-port 8080 --grpc-port 9090

# 启动 Agent
lava tunnel agent --gateway localhost:7000 --service my-service

# 查看服务列表
lava tunnel services

# 查看隧道状态
lava tunnel status
```

## 6. 集成示例

### 6.1 完整服务示例

```go
import (
    "context"
    "github.com/pubgo/lava/v2/core/lavabuilder"
    "github.com/pubgo/lava/v2/core/logging/logbuilder"
    "github.com/pubgo/lava/v2/core/metrics/metricbuilder"
    "github.com/pubgo/lava/v2/core/supervisor"
    "github.com/pubgo/lava/v2/servers/grpcs"
    "github.com/pubgo/lava/v2/servers/https"
)

func main() {
    // 创建依赖注入容器
    di := lavabuilder.New()
    
    // 注册服务
    di.Provide(func() supervisor.Service {
        return https.New(https.Params{
            Handlers: []lava.HttpRouter{
                // HTTP 路由...
            },
            Middlewares: []lava.Middleware{
                // 中间件...
            },
            M:    metricbuilder.New(),
            Log:  logbuilder.New(),
            Cfg:  &https.Config{},
        })
    })
    
    di.Provide(func() supervisor.Service {
        return grpcs.New(grpcs.Params{
            GrpcRouters: []lava.GrpcRouter{
                // gRPC 服务...
            },
            Middlewares: []lava.Middleware{
                // 中间件...
            },
            M:    metricbuilder.New(),
            Log:  logbuilder.New(),
            Cfg:  &grpcs.Config{},
        })
    })
    
    // 运行应用
    lavabuilder.Run(di)
}
```

### 6.2 微服务示例

**服务 A (用户服务)**：
```go
// 实现用户服务
func main() {
    di := lavabuilder.New()
    
    // 注册 gRPC 服务
    di.Provide(func() supervisor.Service {
        return grpcs.New(grpcs.Params{
            GrpcRouters: []lava.GrpcRouter{
                &UserService{},
            },
            // 其他参数...
        })
    })
    
    lavabuilder.Run(di)
}
```

**服务 B (订单服务)**：
```go
// 实现订单服务
func main() {
    di := lavabuilder.New()
    
    // 注册 gRPC 服务
    di.Provide(func() supervisor.Service {
        return grpcs.New(grpcs.Params{
            GrpcRouters: []lava.GrpcRouter{
                &OrderService{},
            },
            // 其他参数...
        })
    })
    
    lavabuilder.Run(di)
}
```

**服务 C (网关服务)**：
```go
// 实现网关服务
func main() {
    di := lavabuilder.New()
    
    // 注册 HTTP 服务
    di.Provide(func() supervisor.Service {
        return https.New(https.Params{
            Handlers: []lava.HttpRouter{
                &GatewayRouter{},
            },
            // 其他参数...
        })
    })
    
    lavabuilder.Run(di)
}
```

## 7. 配置参考

### 7.1 全局配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `app.name` | string | "lava" | 应用名称 |
| `app.env` | string | "development" | 应用环境 |
| `app.version` | string | "1.0.0" | 应用版本 |
| `http_server.port` | int | 8080 | HTTP 服务器端口 |
| `grpc_server.port` | int | 9090 | gRPC 服务器端口 |
| `logging.level` | string | "info" | 日志级别 |
| `metrics.enabled` | bool | true | 是否启用指标 |
| `tracing.enabled` | bool | false | 是否启用追踪 |

### 7.2 隧道配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `tunnel.agent.enabled` | bool | false | 是否启用 Agent |
| `tunnel.agent.gateway_addr` | string | "localhost:7000" | Gateway 地址 |
| `tunnel.agent.service_name` | string | "" | 服务名称 |
| `tunnel.gateway.enabled` | bool | false | 是否启用 Gateway |
| `tunnel.gateway.listen_addr` | string | ":7000" | 监听地址 |
| `tunnel.gateway.http_port` | int | 8080 | HTTP 端口 |
| `tunnel.gateway.grpc_port` | int | 9090 | gRPC 端口 |

### 7.3 调度器配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `scheduler.enabled` | bool | false | 是否启用调度器 |
| `scheduler.jobs` | []JobConfig | [] | 任务配置 |
| `scheduler.jobs[].name` | string | "" | 任务名称 |
| `scheduler.jobs[].schedule` | string | "" | cron 表达式 |
| `scheduler.jobs[].handler` | string | "" | 任务处理器 |

## 8. 最佳实践

### 8.1 服务开发

1. **模块化设计**：将服务拆分为小的、可测试的模块
2. **接口定义**：使用接口定义清晰的模块边界
3. **依赖注入**：使用依赖注入管理对象生命周期
4. **错误处理**：统一的错误处理机制
5. **日志记录**：合理的日志记录，包括请求和响应
6. **指标监控**：为关键操作添加指标
7. **链路追踪**：为跨服务调用添加追踪
8. **健康检查**：实现健康检查接口

### 8.2 部署运维

1. **容器化**：使用 Docker 容器化部署
2. **编排**：使用 Kubernetes 进行服务编排
3. **配置管理**：使用配置中心管理配置
4. **服务发现**：使用服务发现机制
5. **负载均衡**：使用负载均衡分散请求
6. **监控告警**：配置监控和告警
7. **日志收集**：集中收集和分析日志
8. **备份恢复**：实现数据备份和恢复策略

### 8.3 性能优化

1. **连接池**：使用连接池复用数据库和网络连接
2. **缓存**：合理使用缓存减少外部调用
3. **异步处理**：对耗时操作使用异步处理
4. **批量操作**：使用批量 API 减少网络往返
5. **数据压缩**：启用 HTTP 和 gRPC 的压缩
6. **内存优化**：减少内存分配和垃圾回收
7. **并发控制**：合理控制并发度
8. **代码优化**：优化热点代码路径

### 8.4 安全实践

1. **认证授权**：实现认证和授权机制
2. **加密传输**：使用 TLS 加密网络传输
3. **输入验证**：验证所有输入数据
4. **防止注入**：防止 SQL 注入等攻击
5. **权限控制**：最小权限原则
6. **安全配置**：使用安全的默认配置
7. **定期更新**：定期更新依赖和安全补丁
8. **安全扫描**：定期进行安全扫描

## 9. 故障排查

### 9.1 常见问题

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| 服务启动失败 | 端口被占用 | 检查端口占用情况，修改配置 |
| 服务连接失败 | 网络问题 | 检查网络连接，防火墙设置 |
| 服务响应缓慢 | 资源不足 | 检查 CPU、内存使用情况 |
| 数据库连接失败 | 数据库问题 | 检查数据库状态和连接配置 |
| 配置错误 | 配置文件问题 | 检查配置文件格式和内容 |
| 依赖缺失 | 依赖问题 | 检查依赖是否正确安装 |

### 9.2 排查工具

1. **日志分析**：查看应用日志和系统日志
2. **指标监控**：使用 Prometheus 和 Grafana 监控指标
3. **链路追踪**：使用 Jaeger 查看请求链路
4. **网络工具**：使用 ping、telnet、curl 等工具检查网络
5. **系统工具**：使用 top、df、netstat 等工具检查系统状态
6. **调试接口**：使用应用的调试接口查看内部状态
7. **健康检查**：使用健康检查接口检查服务状态
8. **性能分析**：使用 pprof 进行性能分析

## 10. 总结

Lava 框架提供了一套完整的微服务开发和运行体系，包含丰富的功能模块和工具。通过本文档的说明，您应该对 Lava 的各个功能模块有了详细的了解，可以根据自己的需求选择合适的模块进行使用。

核心优势：

1. **统一的服务管理**：通过 Supervisor 统一管理所有服务的生命周期
2. **多协议支持**：同时支持 HTTP、gRPC 和 gRPC Web
3. **强大的隧道系统**：基于反向连接的服务代理和内网穿透
4. **完善的可观测性**：内置日志、指标和追踪系统
5. **调试友好**：提供丰富的调试接口和详细的系统状态信息
6. **插件化架构**：通过插件机制扩展框架功能
7. **配置驱动**：通过配置文件控制服务行为
8. **性能优化**：在设计阶段就考虑性能因素
9. **可靠性设计**：包含故障恢复、重试机制和健康检查
10. **安全设计**：提供认证、授权和安全保护机制

Lava 框架不仅是一个技术工具，更是一套完整的微服务开发方法论，帮助开发者专注于业务逻辑的实现，而将基础设施的复杂性交给框架处理。通过遵循本文档中的最佳实践，可以构建出高性能、可靠、可观测的微服务系统。