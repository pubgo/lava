# Lava 设计文档

## 1. 设计原则

### 1.1 核心设计理念

1. **约定优于配置**：通过合理的默认配置和约定，减少开发者的配置负担
2. **插件化架构**：所有功能通过插件机制实现，便于扩展和定制
3. **统一抽象**：为不同类型的服务提供统一的抽象接口
4. **可观测性优先**：内置完善的日志、指标和追踪系统
5. **调试友好**：提供丰富的调试接口和详细的系统状态信息
6. **性能优化**：在设计阶段就考虑性能因素，如连接池、异步处理等
7. **可靠性设计**：包含故障恢复、重试机制和健康检查
8. **向后兼容**：API 设计考虑向后兼容性，减少升级成本

### 1.2 架构设计原则

1. **分层架构**：清晰的分层设计，每层职责单一
2. **依赖注入**：使用依赖注入框架管理对象生命周期
3. **接口分离**：通过接口定义清晰的模块边界
4. **服务自治**：每个服务都是独立的、可管理的单元
5. **资源管理**：统一管理所有资源的生命周期
6. **配置驱动**：通过配置文件控制服务行为
7. **中间件链**：通过中间件机制实现横切关注点

## 2. 核心组件设计

### 2.1 服务管理器 (Supervisor)

#### 设计目标
- 统一管理所有服务的生命周期
- 提供服务健康状态监控
- 支持服务自动重启和故障恢复
- 提供调试和管理接口

#### 核心设计

```go
// Service 服务接口
type Service interface {
    String() string           // 服务名称
    Serve(ctx context.Context) error // 服务启动逻辑
}

// ServiceConfig 服务配置
type ServiceConfig struct {
    Name              string        // 服务名称
    MaxRestarts       int           // 最大重启次数
    RestartWindow     time.Duration // 重启窗口
    RestartDelay      time.Duration // 重启延迟
    MaxConsecFailures int           // 最大连续失败次数
}

// Manager 服务管理器
type Manager struct {
    name     string
    services map[string]*serviceRunner
    logger   log.Logger
    // ... 其他字段
}
```

#### 关键设计点
- **服务状态管理**：跟踪每个服务的运行状态、重启次数、失败次数等
- **重启策略**：实现指数退避重启策略，避免频繁重启
- **健康检查**：定期检查服务健康状态
- **调试接口**：提供 HTTP API 用于管理和监控服务
- **并发安全**：使用互斥锁保证并发操作安全

### 2.2 HTTP 服务器 (HTTPServer)

#### 设计目标
- 提供高性能的 HTTP 服务
- 支持路由管理和中间件
- 集成调试接口
- 支持 TLS 和其他 HTTP 特性

#### 核心设计

```go
// HttpRouter HTTP 路由接口
type HttpRouter interface {
    Router(router any)         // 注册路由
    Prefix() string            // 路由前缀
    Middlewares() []lava.Middleware // 中间件
}

// Config HTTP 服务器配置
type Config struct {
    Http *fiber.Config // Fiber 配置
    // ... 其他配置
}

// serviceImpl HTTP 服务实现
type serviceImpl struct {
    httpServer *fiber.App
    log        log.Logger
    // ... 其他字段
}
```

#### 关键设计点
- **Fiber 集成**：基于高性能的 Fiber 框架
- **中间件链**：统一的中间件处理机制
- **路由管理**：支持分组路由和前缀管理
- **调试接口**：自动挂载调试 API
- **配置灵活**：支持详细的 HTTP 服务器配置

### 2.3 gRPC 服务器 (gRPCServer)

#### 设计目标
- 提供标准的 gRPC 服务
- 集成 gRPC Gateway 支持 HTTP/JSON
- 支持中间件和拦截器
- 统一的服务注册机制

#### 核心设计

```go
// GrpcRouter gRPC 路由接口
type GrpcRouter interface {
    ServiceDesc() *grpc.ServiceDesc // 服务描述
}

// GrpcHttpRouter 同时支持 gRPC 和 HTTP 的路由接口
type GrpcHttpRouter interface {
    HttpRouter
    GrpcRouter
}

// Config gRPC 服务器配置
type Config struct {
    GrpcConfig *grpc.Config // gRPC 配置
    Http       *fiber.Config // HTTP 配置
    // ... 其他配置
}
```

#### 关键设计点
- **标准 gRPC**：基于官方 gRPC 库
- **gRPC Gateway**：自动转换 HTTP/JSON 到 gRPC
- **服务注册**：统一的服务注册机制
- **中间件集成**：支持 gRPC 拦截器
- **路由管理**：自动生成和管理路由

### 2.4 调度器 (Scheduler)

#### 设计目标
- 基于 cron 表达式的任务调度
- 支持任务管理和监控
- 提供任务执行历史
- 集成调试接口

#### 核心设计

```go
// Job 任务接口
type Job interface {
    ID() string        // 任务 ID
    Name() string      // 任务名称
    Schedule() string  // cron 表达式
    Run(ctx context.Context) error // 任务执行逻辑
}

// Config 调度器配置
type Config struct {
    Jobs []JobConfig // 任务配置
    // ... 其他配置
}

// Scheduler 调度器接口
type Scheduler interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    AddJob(job Job) error
    RemoveJob(id string) error
    Jobs() []Job
}
```

#### 关键设计点
- **Cron 表达式**：支持标准 cron 表达式
- **任务管理**：动态添加、删除和查询任务
- **执行历史**：记录任务执行状态和结果
- **错误处理**：包含任务执行错误处理和重试机制
- **调试接口**：提供任务状态和执行历史查询

### 2.5 隧道系统 (Tunnel)

#### 设计目标
- 实现基于反向连接的服务代理
- 支持内网穿透
- 提供多种传输协议
- 支持服务注册和发现

#### 核心设计

```go
// Transport 传输层接口
type Transport interface {
    Name() string
    Dial(ctx context.Context, addr string) (Session, error)
    Listen(ctx context.Context, addr string) (Listener, error)
}

// Session 会话接口
type Session interface {
    io.Closer
    Open(ctx context.Context) (Stream, error)
    Accept() (Stream, error)
    IsClosed() bool
    // ... 其他方法
}

// Agent 代理客户端接口
type Agent interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Register(ctx context.Context, service *ServiceInfo) error
    // ... 其他方法
}

// Gateway 网关接口
type Gateway interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Services() []*ServiceInfo
    // ... 其他方法
}
```

#### 关键设计点
- **传输层抽象**：支持多种传输协议（Yamux、QUIC、HTTP、KCP）
- **反向连接**：Agent 主动连接 Gateway，实现内网穿透
- **服务注册**：Agent 向 Gateway 注册服务信息
- **心跳机制**：保持连接活跃和检测故障
- **多路复用**：在单个连接上复用多个流
- **服务发现**：Gateway 提供服务列表和状态

### 2.6 中间件系统 (Middleware)

#### 设计目标
- 提供统一的中间件接口
- 支持 HTTP 和 gRPC 共享中间件
- 实现中间件链和顺序控制
- 支持中间件嵌套和组合

#### 核心设计

```go
// Middleware 中间件接口
type Middleware interface {
    Handle(ctx context.Context, req interface{}) (interface{}, error)
}

// Middlewares 中间件链
type Middlewares []Middleware

// Handle 执行中间件链
func (ms Middlewares) Handle(ctx context.Context, req interface{}) (interface{}, error) {
    // 实现中间件链执行逻辑
}
```

#### 关键设计点
- **统一接口**：HTTP 和 gRPC 使用相同的中间件接口
- **中间件链**：按顺序执行中间件
- **上下文传递**：通过 context 传递请求上下文
- **错误处理**：统一的错误处理机制
- **内置中间件**：提供常用中间件如日志、指标、恢复等

### 2.7 可观测性系统

#### 设计目标
- 提供统一的日志、指标和追踪系统
- 支持多种输出和导出器
- 集成 OpenTelemetry
- 提供详细的系统状态信息

#### 核心设计

**日志系统**：
```go
// Logger 日志接口
type Logger interface {
    Info() log.Logger
    Debug() log.Logger
    Error() log.Logger
    // ... 其他方法
}

// Config 日志配置
type Config struct {
    Level  string // 日志级别
    Format string // 日志格式
    Output string // 输出位置
    // ... 其他配置
}
```

**指标系统**：
```go
// Metric 指标接口
type Metric interface {
    Counter(name string, labels ...string) Counter
    Gauge(name string, labels ...string) Gauge
    Histogram(name string, labels ...string) Histogram
    // ... 其他方法
}
```

**追踪系统**：
```go
// Tracer 追踪接口
type Tracer interface {
    Start(ctx context.Context, name string) (context.Context, Span)
    // ... 其他方法
}
```

#### 关键设计点
- **统一接口**：提供一致的日志、指标和追踪接口
- **OpenTelemetry 集成**：支持标准的可观测性规范
- **多输出支持**：支持多种日志输出和指标导出
- **结构化日志**：支持 JSON 等结构化日志格式
- **上下文传递**：通过 context 传递追踪信息

### 2.8 配置系统 (Config)

#### 设计目标
- 提供统一的配置管理
- 支持本地配置和配置中心
- 支持配置热更新
- 支持配置验证和默认值

#### 核心设计

```go
// Config 配置接口
type Config interface {
    Load() error
    Watch() error
    Get(key string) interface{}
    Set(key string, value interface{})
    // ... 其他方法
}

// ConfigSource 配置源接口
type ConfigSource interface {
    Load() (map[string]interface{}, error)
    Watch(callback func(map[string]interface{})) error
}
```

#### 关键设计点
- **配置源抽象**：支持多种配置源如文件、环境变量、配置中心
- **配置合并**：支持多层配置合并
- **配置验证**：验证配置有效性
- **默认值**：提供合理的默认配置
- **热更新**：支持配置变更自动生效

### 2.9 gRPC Gateway

#### 设计目标
- 提供 HTTP/JSON 到 gRPC 的协议转换
- 支持 Google API HTTP 注解
- 支持 gRPC Web
- 提供路由管理和服务注册

#### 核心设计

```go
// Mux gRPC Gateway 核心结构
type Mux struct {
    routerTree *routertree.RouterTree
    codecs     map[string]Codec
    services   map[string]*serviceInfo
    // ... 其他字段
}

// RegisterService 注册服务
func (m *Mux) RegisterService(desc *grpc.ServiceDesc, srv interface{}) {
    // 实现服务注册逻辑
}

// Handler HTTP 处理器
func (m *Mux) Handler(c *fiber.Ctx) error {
    // 实现 HTTP 请求处理逻辑
}
```

#### 关键设计点
- **路由树**：高效的路由匹配算法
- **协议转换**：自动处理 HTTP/JSON 和 gRPC 之间的转换
- **服务注册**：支持本地服务和代理服务注册
- **中间件支持**：集成中间件机制
- **错误映射**：自动将 gRPC 错误码映射为 HTTP 状态码

## 3. API 设计

### 3.1 服务 API

#### HTTP 服务 API

| API 路径 | 方法 | 模块 | 描述 |
|----------|------|------|------|
| `/debug/services` | GET | Supervisor | 服务列表 |
| `/debug/service/{name}` | GET | Supervisor | 服务详情 |
| `/debug/service/{name}/restart` | POST | Supervisor | 重启服务 |
| `/debug/service/{name}/stop` | POST | Supervisor | 停止服务 |
| `/debug/service/{name}/start` | POST | Supervisor | 启动服务 |

#### gRPC 服务 API

```protobuf
service Health {
    rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
    rpc Watch(HealthCheckRequest) returns (stream HealthCheckResponse);
}

service Debug {
    rpc GetConfig(ConfigRequest) returns (ConfigResponse);
    rpc GetMetrics(MetricsRequest) returns (MetricsResponse);
    rpc GetServices(ServicesRequest) returns (ServicesResponse);
}
```

### 3.2 隧道系统 API

#### Agent API

| API 方法 | 模块 | 描述 |
|----------|------|------|
| `Register` | Agent | 注册服务 |
| `Deregister` | Agent | 注销服务 |
| `Heartbeat` | Agent | 发送心跳 |

#### Gateway API

| API 路径 | 方法 | 模块 | 描述 |
|----------|------|------|------|
| `/tunnel/` | GET | Gateway | 隧道概览 |
| `/tunnel/gateway` | GET | Gateway | 网关状态 |
| `/tunnel/gateway/services` | GET | Gateway | 服务列表 |
| `/tunnel/gateway/services/{name}` | GET | Gateway | 服务详情 |

### 3.3 调度器 API

| API 路径 | 方法 | 模块 | 描述 |
|----------|------|------|------|
| `/scheduler/jobs` | GET | Scheduler | 任务列表 |
| `/scheduler/job/{id}` | GET | Scheduler | 任务详情 |
| `/scheduler/job/{id}/run` | POST | Scheduler | 立即执行任务 |
| `/scheduler/job/{id}/stop` | POST | Scheduler | 停止任务 |
| `/scheduler/history` | GET | Scheduler | 执行历史 |

### 3.4 调试 API

| API 路径 | 方法 | 模块 | 描述 |
|----------|------|------|------|
| `/debug/pprof/` | GET | Debug | pprof 调试接口 |
| `/debug/vars` | GET | Debug | 系统变量 |
| `/debug/version` | GET | Debug | 版本信息 |
| `/debug/runtime` | GET | Debug | 运行时信息 |
| `/debug/goroutines` | GET | Debug | goroutine 信息 |
| `/debug/process` | GET | Debug | 进程信息 |
| `/debug/statsviz` | GET | Debug | 统计可视化 |

## 4. 配置设计

### 4.1 配置结构

#### 核心配置结构

```yaml
# 应用配置
app:
  name: "my-service"
  env: "production"
  version: "1.0.0"

# HTTP 服务器配置
http_server:
  base_url: "/api"
  enable_print_router: true
  http:
    port: 8080
    read_timeout: 10s
    write_timeout: 10s

# gRPC 服务器配置
grpc_server:
  grpc:
    port: 9090
    max_concurrent_streams: 100

# 日志配置
logging:
  level: "info"
  format: "json"
  output: "stdout"

# 指标配置
metrics:
  enabled: true
  exporter: "prometheus"
  port: 9091

# 追踪配置
tracing:
  enabled: true
  exporter: "jaeger"
  endpoint: "http://jaeger:14268/api/traces"

# 隧道配置
tunnel:
  agent:
    enabled: true
    gateway_addr: "gateway:7000"
    service_name: "my-service"
    endpoints:
      - type: "http"
        local_addr: "localhost:8080"
        path: "/api"

# 调度器配置
scheduler:
  enabled: true
  jobs:
    - name: "backup"
      schedule: "0 0 * * *"
      handler: "backupHandler"

# 服务发现配置
discovery:
  type: "consul"
  address: "consul:8500"
  interval: "30s"
```

### 4.2 配置加载顺序

1. **默认配置**：框架内置的默认配置
2. **环境变量**：通过环境变量覆盖配置
3. **配置文件**：本地配置文件
4. **配置中心**：从配置中心获取的配置
5. **运行时配置**：通过 API 动态修改的配置

### 4.3 配置验证

- **结构验证**：验证配置结构是否正确
- **类型验证**：验证配置值类型是否正确
- **范围验证**：验证配置值是否在合理范围内
- **依赖验证**：验证配置之间的依赖关系

## 5. 插件系统设计

### 5.1 设计目标
- 提供统一的插件注册机制
- 支持插件生命周期管理
- 允许插件扩展框架功能
- 提供插件依赖管理

### 5.2 核心设计

```go
// Plugin 插件接口
type Plugin interface {
    Name() string              // 插件名称
    Init(ctx context.Context) error // 插件初始化
    Start(ctx context.Context) error // 插件启动
    Stop(ctx context.Context) error  // 插件停止
    Dependencies() []string    // 插件依赖
}

// PluginManager 插件管理器
type PluginManager struct {
    plugins map[string]Plugin
    // ... 其他字段
}

// Register 注册插件
func (pm *PluginManager) Register(plugin Plugin) error {
    // 实现插件注册逻辑
}

// Start 启动所有插件
func (pm *PluginManager) Start(ctx context.Context) error {
    // 实现插件启动逻辑
}
```

### 5.3 插件类型

1. **核心插件**：框架内置的插件，如日志、指标、追踪等
2. **服务插件**：提供特定服务的插件，如 HTTP 服务器、gRPC 服务器等
3. **工具插件**：提供工具功能的插件，如调试工具、配置工具等
4. **第三方插件**：由第三方开发的插件

### 5.4 插件加载机制

- **自动加载**：通过 init() 函数自动注册插件
- **手动加载**：通过 API 手动注册插件
- **配置加载**：根据配置文件加载插件

## 6. 错误处理设计

### 6.1 设计目标
- 提供统一的错误处理机制
- 支持错误分类和错误码
- 提供详细的错误信息
- 支持错误链和错误包装
- 集成日志和监控系统

### 6.2 核心设计

```go
// Error 错误接口
type Error interface {
    Error() string           // 错误信息
    Code() int               // 错误码
    Message() string         // 用户友好的错误信息
    Details() map[string]interface{} // 错误详情
    Cause() error            // 原始错误
}

// New 创建错误
func New(code int, message string, details ...map[string]interface{}) Error {
    // 实现错误创建逻辑
}

// Wrap 包装错误
func Wrap(err error, code int, message string, details ...map[string]interface{}) Error {
    // 实现错误包装逻辑
}
```

### 6.3 错误分类

| 错误类型 | 错误码范围 | 描述 |
|----------|------------|------|
| 系统错误 | 1000-1999 | 框架内部错误 |
| 配置错误 | 2000-2999 | 配置相关错误 |
| 网络错误 | 3000-3999 | 网络相关错误 |
| 服务错误 | 4000-4999 | 服务相关错误 |
| 业务错误 | 5000-5999 | 业务逻辑错误 |
| 资源错误 | 6000-6999 | 资源相关错误 |
| 权限错误 | 7000-7999 | 权限相关错误 |
|  validation 错误 | 8000-8999 | 数据验证错误 |

### 6.4 错误处理最佳实践

1. **错误包装**：使用 Wrap 函数包装错误，保留错误上下文
2. **错误分类**：根据错误类型返回适当的错误码
3. **错误日志**：记录详细的错误信息和上下文
4. **错误监控**：监控错误率和错误类型分布
5. **用户友好**：向用户返回友好的错误信息，避免暴露内部错误细节
6. **错误重试**：对可重试的错误实现重试机制

## 7. 安全设计

### 7.1 设计目标
- 提供安全的服务访问控制
- 保护敏感信息
- 防止常见的安全攻击
- 支持 TLS 和加密
- 提供审计日志

### 7.2 核心设计

**认证和授权**：
```go
// Authenticator 认证接口
type Authenticator interface {
    Authenticate(ctx context.Context, req interface{}) (context.Context, error)
}

// Authorizer 授权接口
type Authorizer interface {
    Authorize(ctx context.Context, resource string, action string) error
}
```

**加密和安全**：
- **TLS 支持**：所有网络通信支持 TLS
- **敏感信息保护**：配置中的敏感信息加密存储
- **密码哈希**：使用安全的密码哈希算法
- **防止注入**：防止 SQL 注入、命令注入等攻击
- **CSRF 保护**：防止跨站请求伪造攻击
- **XSS 保护**：防止跨站脚本攻击

**审计日志**：
- 记录所有认证和授权操作
- 记录敏感操作和配置变更
- 提供日志完整性保护

### 7.3 安全最佳实践

1. **最小权限原则**：每个服务和用户只拥有必要的权限
2. **安全配置**：使用安全的默认配置
3. **定期更新**：定期更新依赖和安全补丁
4. **安全扫描**：定期进行安全扫描和渗透测试
5. **安全监控**：监控异常访问和安全事件
6. **安全文档**：提供详细的安全配置和最佳实践文档

## 8. 性能设计

### 8.1 设计目标
- 提供高性能的服务处理
- 优化资源使用
- 减少延迟和提高吞吐量
- 支持高并发处理
- 提供性能监控和分析

### 8.2 核心设计

**连接池管理**：
```go
// ConnectionPool 连接池接口
type ConnectionPool interface {
    Get() (Connection, error)
    Put(conn Connection)
    Close() error
    // ... 其他方法
}
```

**异步处理**：
- **协程池**：使用协程池处理并发任务
- **消息队列**：使用消息队列处理异步任务
- **非阻塞 I/O**：使用非阻塞 I/O 提高性能

**缓存设计**：
- **本地缓存**：使用内存缓存减少外部调用
- **分布式缓存**：使用分布式缓存提高扩展性
- **缓存策略**：实现合理的缓存过期和更新策略

**性能监控**：
- **关键指标**：监控请求延迟、QPS、错误率等关键指标
- **性能分析**：提供 pprof 等性能分析工具
- **瓶颈检测**：自动检测性能瓶颈

### 8.3 性能优化最佳实践

1. **连接复用**：使用连接池复用网络连接
2. **批量操作**：使用批量 API 减少网络往返
3. **数据压缩**：启用 HTTP 和 gRPC 的压缩
4. **内存优化**：减少内存分配和垃圾回收
5. **并发控制**：合理控制并发度，避免过度并发
6. **负载均衡**：使用负载均衡分散请求
7. **预热机制**：启动时预热缓存和连接池
8. **超时设置**：合理设置请求超时，避免长时间阻塞

## 9. 部署设计

### 9.1 设计目标
- 支持多种部署方式
- 提供标准化的部署配置
- 支持容器化部署
- 提供健康检查和就绪探针
- 支持滚动更新和回滚

### 9.2 部署架构

**容器化部署**：
- **Dockerfile**：标准化的 Docker 构建文件
- **Docker Compose**：开发和测试环境的服务编排
- **Kubernetes**：生产环境的容器编排

**Kubernetes 部署**：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-service
  template:
    metadata:
      labels:
        app: my-service
    spec:
      containers:
      - name: my-service
        image: my-service:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        readinessProbe:
          httpGet:
            path: /debug/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /debug/health
            port: 8080
          initialDelaySeconds: 15
          periodSeconds: 20
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
          requests:
            cpu: "500m"
            memory: "512Mi"
```

**服务编排**：
- **服务发现**：使用 Kubernetes Service 或注册中心
- **配置管理**：使用 ConfigMap 和 Secret 管理配置
- **存储管理**：使用 PersistentVolume 管理存储
- **网络管理**：使用 Service 和 Ingress 管理网络

### 9.3 部署最佳实践

1. **环境隔离**：开发、测试、生产环境隔离
2. **标准化部署**：使用标准化的部署配置和脚本
3. **自动化部署**：使用 CI/CD 流水线自动化部署
4. **健康检查**：配置适当的健康检查和就绪探针
5. **资源限制**：设置合理的资源限制和请求
6. **监控集成**：集成 Prometheus 和 Grafana 监控
7. **日志收集**：使用 ELK 或 Loki 收集和分析日志
8. **备份策略**：实现数据备份和恢复策略

## 10. 总结

Lava 框架的设计遵循了现代软件工程的最佳实践，通过清晰的分层架构、统一的抽象接口、完善的可观测性系统和丰富的扩展点，为微服务开发提供了一套完整的解决方案。

核心设计亮点包括：

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

Lava 框架不仅是一个技术工具，更是一套完整的微服务开发方法论，帮助开发者专注于业务逻辑的实现，而将基础设施的复杂性交给框架处理。通过遵循本文档中的设计原则和最佳实践，可以构建出高性能、可靠、可观测的微服务系统。