# Lava 快速开始文档

## 1. 环境准备

### 1.1 系统要求

| 环境 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.25+ | 必须使用 Go 1.25 或更高版本 |
| Git | 2.0+ | 用于代码管理 |
| Task | 3.0+ | 用于运行构建和开发命令 |
| Protobuf | 3.0+ | 用于生成 gRPC 代码（可选） |

### 1.2 安装依赖工具

#### 安装 Go

**macOS**：
```bash
brew install go
```

**Linux**：
```bash
sudo apt-get update
sudo apt-get install golang-go
```

**Windows**：
下载并安装 [Go 安装包](https://golang.org/dl/)

#### 安装 Task

**macOS**：
```bash
brew install go-task/tap/go-task
brew install go-task
eval "$(task --completion zsh)"
```

**Linux**：
```bash
sh -c "$(curl -ss https://taskfile.dev/install.sh)"
```

**Windows**：
```bash
scoop install task
```

#### 安装 Protobuf (可选)

**macOS**：
```bash
brew install protobuf
```

**Linux**：
```bash
sudo apt-get install protobuf-compiler
```

**Windows**：
下载并安装 [Protobuf 编译器](https://github.com/protocolbuffers/protobuf/releases)

## 2. 安装 Lava

### 2.1 作为依赖引入

在你的 Go 项目中，添加 Lava 作为依赖：

```bash
go get github.com/pubgo/lava/v2
```

### 2.2 克隆源码

如果你想直接使用 Lava 的示例或进行开发：

```bash
git clone https://github.com/pubgo/lava.git
cd lava
go mod tidy
```

## 3. 快速创建第一个服务

### 3.1 创建项目结构

```bash
mkdir -p my-lava-service
cd my-lava-service
go mod init my-lava-service
go get github.com/pubgo/lava/v2
```

### 3.2 创建 main.go 文件

```go
package main

import (
    "github.com/pubgo/lava/v2/core/lavabuilder"
)

func main() {
    // 创建依赖注入容器
    di := lavabuilder.New()
    
    // 运行应用
    lavabuilder.Run(di)
}
```

### 3.3 创建配置文件

创建 `config.yaml` 文件：

```yaml
app:
  name: "my-lava-service"
  env: "development"
  version: "1.0.0"

http_server:
  base_url: "/api"
  enable_print_router: true

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  exporter: "prometheus"
  port: 9091
```

## 4. 添加 HTTP 服务

### 4.1 创建 HTTP 路由

创建 `routes/user_router.go` 文件：

```go
package routes

import (
    "github.com/gofiber/fiber/v2"
    "github.com/pubgo/lava/v2/lava"
)

// UserRouter 用户路由
type UserRouter struct{}

// Prefix 路由前缀
func (r *UserRouter) Prefix() string {
    return "/users"
}

// Middlewares 中间件
func (r *UserRouter) Middlewares() []lava.Middleware {
    return nil
}

// Router 路由注册
func (r *UserRouter) Router(router fiber.Router) {
    router.Get("/", r.GetUsers)
    router.Get("/:id", r.GetUser)
    router.Post("/", r.CreateUser)
    router.Put("/:id", r.UpdateUser)
    router.Delete("/:id", r.DeleteUser)
}

// GetUsers 获取用户列表
func (r *UserRouter) GetUsers(c *fiber.Ctx) error {
    return c.JSON([]map[string]interface{}{
        {"id": "1", "name": "John", "age": 30},
        {"id": "2", "name": "Jane", "age": 28},
    })
}

// GetUser 获取单个用户
func (r *UserRouter) GetUser(c *fiber.Ctx) error {
    id := c.Params("id")
    return c.JSON(map[string]interface{}{
        "id":   id,
        "name": "John",
        "age":  30,
    })
}

// CreateUser 创建用户
func (r *UserRouter) CreateUser(c *fiber.Ctx) error {
    var req map[string]interface{}
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(map[string]interface{}{
            "error": "Invalid request",
        })
    }
    
    return c.Status(201).JSON(map[string]interface{}{
        "id":   "3",
        "name": req["name"],
        "age":  req["age"],
    })
}

// UpdateUser 更新用户
func (r *UserRouter) UpdateUser(c *fiber.Ctx) error {
    id := c.Params("id")
    var req map[string]interface{}
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(map[string]interface{}{
            "error": "Invalid request",
        })
    }
    
    return c.JSON(map[string]interface{}{
        "id":   id,
        "name": req["name"],
        "age":  req["age"],
    })
}

// DeleteUser 删除用户
func (r *UserRouter) DeleteUser(c *fiber.Ctx) error {
    id := c.Params("id")
    return c.JSON(map[string]interface{}{
        "message": "User deleted",
        "id":      id,
    })
}
```

### 4.2 更新 main.go 文件

```go
package main

import (
    "github.com/pubgo/lava/v2/core/lavabuilder"
    "github.com/pubgo/lava/v2/core/logging/logbuilder"
    "github.com/pubgo/lava/v2/core/metrics/metricbuilder"
    "github.com/pubgo/lava/v2/lava"
    "github.com/pubgo/lava/v2/servers/https"
    "my-lava-service/routes"
)

func main() {
    // 创建依赖注入容器
    di := lavabuilder.New()
    
    // 注册 HTTP 服务器
    di.Provide(func() lava.Service {
        return https.New(https.Params{
            Handlers: []lava.HttpRouter{
                &routes.UserRouter{},
            },
            Middlewares: nil,
            M:           metricbuilder.New(),
            Log:         logbuilder.New(),
            Cfg:         nil,
        })
    })
    
    // 运行应用
    lavabuilder.Run(di)
}
```

### 4.3 运行服务

```bash
go run main.go
```

服务启动后，你可以访问以下端点：
- `http://localhost:8080/api/users` - 获取用户列表
- `http://localhost:8080/api/users/1` - 获取单个用户
- `http://localhost:8080/debug` - 调试接口

## 5. 添加 gRPC 服务

### 5.1 创建 Protobuf 文件

创建 `proto/user.proto` 文件：

```protobuf
syntax = "proto3";

package user;

option go_package = "my-lava-service/proto/user";

// User 用户消息
message User {
  string id = 1;
  string name = 2;
  int32 age = 3;
}

// GetUserRequest 获取用户请求
message GetUserRequest {
  string user_id = 1;
}

// CreateUserRequest 创建用户请求
message CreateUserRequest {
  string name = 1;
  int32 age = 2;
}

// UpdateUserRequest 更新用户请求
message UpdateUserRequest {
  string user_id = 1;
  string name = 2;
  int32 age = 3;
}

// DeleteUserRequest 删除用户请求
message DeleteUserRequest {
  string user_id = 1;
}

// UserListResponse 用户列表响应
message UserListResponse {
  repeated User users = 1;
}

// UserResponse 用户响应
message UserResponse {
  User user = 1;
}

// MessageResponse 消息响应
message MessageResponse {
  string message = 1;
  string id = 2;
}

// UserService 用户服务
service UserService {
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc GetUsers(google.protobuf.Empty) returns (UserListResponse);
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (MessageResponse);
}

import "google/protobuf/empty.proto";
```

### 5.2 生成 gRPC 代码

```bash
mkdir -p proto/user
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
protoc --go_out=proto/user --go-grpc_out=proto/user proto/user.proto
```

### 5.3 实现 gRPC 服务

创建 `services/user_service.go` 文件：

```go
package services

import (
    "context"
    "my-lava-service/proto/user"
    "google.golang.org/protobuf/types/known/emptypb"
)

// UserService 用户服务实现
type UserService struct{}

// GetUser 获取用户
func (s *UserService) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
    return &user.UserResponse{
        User: &user.User{
            Id:   req.UserId,
            Name: "John",
            Age:  30,
        },
    }, nil
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(ctx context.Context, req *emptypb.Empty) (*user.UserListResponse, error) {
    return &user.UserListResponse{
        Users: []*user.User{
            {
                Id:   "1",
                Name: "John",
                Age:  30,
            },
            {
                Id:   "2",
                Name: "Jane",
                Age:  28,
            },
        },
    }, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserResponse, error) {
    return &user.UserResponse{
        User: &user.User{
            Id:   "3",
            Name: req.Name,
            Age:  req.Age,
        },
    }, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UserResponse, error) {
    return &user.UserResponse{
        User: &user.User{
            Id:   req.UserId,
            Name: req.Name,
            Age:  req.Age,
        },
    }, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.MessageResponse, error) {
    return &user.MessageResponse{
        Message: "User deleted",
        Id:      req.UserId,
    }, nil
}

// ServiceDesc 获取服务描述
func (s *UserService) ServiceDesc() *grpc.ServiceDesc {
    return &user.UserService_ServiceDesc
}
```

### 5.4 更新 main.go 文件

```go
package main

import (
    "github.com/pubgo/lava/v2/core/lavabuilder"
    "github.com/pubgo/lava/v2/core/logging/logbuilder"
    "github.com/pubgo/lava/v2/core/metrics/metricbuilder"
    "github.com/pubgo/lava/v2/lava"
    "github.com/pubgo/lava/v2/servers/grpcs"
    "github.com/pubgo/lava/v2/servers/https"
    "google.golang.org/grpc"
    "my-lava-service/routes"
    "my-lava-service/services"
)

func main() {
    // 创建依赖注入容器
    di := lavabuilder.New()
    
    // 注册 HTTP 服务器
    di.Provide(func() lava.Service {
        return https.New(https.Params{
            Handlers: []lava.HttpRouter{
                &routes.UserRouter{},
            },
            Middlewares: nil,
            M:           metricbuilder.New(),
            Log:         logbuilder.New(),
            Cfg:         nil,
        })
    })
    
    // 注册 gRPC 服务器
    di.Provide(func() lava.Service {
        return grpcs.New(grpcs.Params{
            GrpcRouters: []lava.GrpcRouter{
                &services.UserService{},
            },
            HttpRouters:     nil,
            GrpcHttpRouters: nil,
            DixMiddlewares:  nil,
            Metric:          metricbuilder.New(),
            Log:             logbuilder.New(),
            Conf:            nil,
            Gw:              nil,
        })
    })
    
    // 运行应用
    lavabuilder.Run(di)
}
```

### 5.5 运行服务

```bash
go run main.go
```

服务启动后，你可以：
- 使用 gRPC 客户端调用 `UserService` 服务
- 通过 `http://localhost:8080/api` 访问 gRPC Gateway
- 访问 `http://localhost:8080/debug` 调试接口

## 6. 使用命令行工具

### 6.1 列出可用命令

```bash
go run main.go help
```

### 6.2 健康检查

```bash
go run main.go health check
```

### 6.3 查看配置

```bash
go run main.go config show
```

### 6.4 查看环境变量

```bash
go run main.go env show
```

### 6.5 使用 lavacurl 调用 API

```bash
# 安装 lavacurl
go install ./cmds/lavacurl

# 调用 HTTP API
lavacurl --path /api/users

# 调用 gRPC API（通过 Gateway）
lavacurl UserService/GetUsers
```

## 7. 添加中间件

### 7.1 创建日志中间件

创建 `middlewares/logger.go` 文件：

```go
package middlewares

import (
    "context"
    "github.com/pubgo/lava/v2/lava"
    "github.com/pubgo/lava/v2/core/logging"
)

// LoggerMiddleware 日志中间件
type LoggerMiddleware struct {
    log logging.Logger
}

// NewLoggerMiddleware 创建日志中间件
func NewLoggerMiddleware(log logging.Logger) lava.Middleware {
    return &LoggerMiddleware{
        log: log,
    }
}

// Handle 处理请求
func (m *LoggerMiddleware) Handle(ctx context.Context, req interface{}) (interface{}, error) {
    m.log.Info().Msg("Request received")
    
    // 继续执行下一个中间件
    return req, nil
}
```

### 7.2 更新 main.go 文件

```go
package main

import (
    "github.com/pubgo/lava/v2/core/lavabuilder"
    "github.com/pubgo/lava/v2/core/logging/logbuilder"
    "github.com/pubgo/lava/v2/core/metrics/metricbuilder"
    "github.com/pubgo/lava/v2/lava"
    "github.com/pubgo/lava/v2/servers/https"
    "my-lava-service/middlewares"
    "my-lava-service/routes"
)

func main() {
    // 创建依赖注入容器
    di := lavabuilder.New()
    
    // 创建日志记录器
    log := logbuilder.New()
    
    // 注册 HTTP 服务器
    di.Provide(func() lava.Service {
        return https.New(https.Params{
            Handlers: []lava.HttpRouter{
                &routes.UserRouter{},
            },
            Middlewares: []lava.Middleware{
                middlewares.NewLoggerMiddleware(log),
            },
            M:           metricbuilder.New(),
            Log:         log,
            Cfg:         nil,
        })
    })
    
    // 运行应用
    lavabuilder.Run(di)
}
```

## 8. 配置管理

### 8.1 使用环境变量

你可以通过环境变量覆盖配置：

```bash
HTTP_PORT=8081 GRPC_PORT=9091 go run main.go
```

### 8.2 使用配置文件

创建 `config.yaml` 文件：

```yaml
app:
  name: "my-lava-service"
  env: "production"
  version: "1.0.0"

http_server:
  base_url: "/api"
  http:
    port: 8080

grpc_server:
  grpc:
    port: 9090

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  exporter: "prometheus"
  port: 9091
```

运行服务时指定配置文件：

```bash
go run main.go -c config.yaml
```

## 9. 调试和监控

### 9.1 调试接口

服务启动后，你可以访问以下调试端点：

| 端点 | 说明 |
|------|------|
| `/debug/pprof/` | pprof 性能分析 |
| `/debug/vars` | 系统变量 |
| `/debug/version` | 版本信息 |
| `/debug/runtime` | 运行时信息 |
| `/debug/goroutines` | goroutine 信息 |
| `/debug/process` | 进程信息 |
| `/debug/statsviz` | 统计可视化 |
| `/debug/supervisor` | 服务管理器 |

### 9.2 指标监控

Lava 默认集成了 Prometheus 指标导出：

1. 启动服务后，访问 `http://localhost:9091/metrics` 查看指标
2. 配置 Prometheus 抓取该端点
3. 使用 Grafana 可视化指标

### 9.3 日志查看

默认情况下，日志输出到标准输出。你可以通过配置文件修改日志输出位置：

```yaml
logging:
  level: "info"
  format: "json"
  output: "/var/log/lava.log"
```

## 10. 部署服务

### 10.1 使用 Docker

创建 `Dockerfile` 文件：

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o my-lava-service .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/my-lava-service .
COPY config.yaml .

EXPOSE 8080 9090 9091

CMD ["./my-lava-service"]
```

构建和运行 Docker 容器：

```bash
docker build -t my-lava-service .
docker run -p 8080:8080 -p 9090:9090 -p 9091:9091 my-lava-service
```

### 10.2 使用 Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3'

services:
  my-lava-service:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
      - "9091:9091"
    environment:
      - APP_ENV=production
      - HTTP_PORT=8080
      - GRPC_PORT=9090
    volumes:
      - ./config.yaml:/app/config.yaml
```

启动服务：

```bash
docker-compose up -d
```

### 10.3 使用 Kubernetes

创建 `kubernetes/deployment.yaml` 文件：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-lava-service
  labels:
    app: my-lava-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-lava-service
  template:
    metadata:
      labels:
        app: my-lava-service
    spec:
      containers:
      - name: my-lava-service
        image: my-lava-service:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        - containerPort: 9091
        env:
        - name: APP_ENV
          value: "production"
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
```

创建 `kubernetes/service.yaml` 文件：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-lava-service
spec:
  selector:
    app: my-lava-service
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  - name: metrics
    port: 9091
    targetPort: 9091
  type: LoadBalancer
```

部署到 Kubernetes：

```bash
kubectl apply -f kubernetes/deployment.yaml
kubectl apply -f kubernetes/service.yaml
```

## 11. 常见问题

### 11.1 服务启动失败

**问题**：服务启动失败，提示端口被占用

**解决方案**：
- 检查端口是否被其他进程占用
- 修改配置文件中的端口设置
- 使用环境变量覆盖端口设置

**命令**：
```bash
# 检查端口占用
lsof -i :8080

# 修改端口
HTTP_PORT=8081 go run main.go
```

### 11.2 依赖解析失败

**问题**：`go mod tidy` 失败，提示依赖解析错误

**解决方案**：
- 检查网络连接
- 检查 Go 版本是否正确
- 清理并重新生成 go.mod 文件

**命令**：
```bash
# 清理依赖
go clean -modcache

# 重新生成 go.mod
go mod init my-lava-service
go get github.com/pubgo/lava/v2
```

### 11.3 gRPC 代码生成失败

**问题**：Protobuf 代码生成失败

**解决方案**：
- 检查 Protobuf 编译器是否安装
- 检查 proto 文件语法是否正确
- 确保安装了正确的 Go 插件

**命令**：
```bash
# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成代码
protoc --go_out=proto --go-grpc_out=proto proto/user.proto
```

### 11.4 调试接口无法访问

**问题**：无法访问 `/debug` 端点

**解决方案**：
- 检查服务是否正常运行
- 检查防火墙设置
- 检查调试接口是否正确挂载

**命令**：
```bash
# 检查服务状态
go run main.go health check

# 检查端口
telnet localhost 8080
```

### 11.5 中间件不生效

**问题**：添加的中间件没有生效

**解决方案**：
- 检查中间件注册是否正确
- 检查中间件实现是否正确
- 检查中间件顺序是否合理

**示例**：
```go
// 正确注册中间件
https.New(https.Params{
    Handlers: []lava.HttpRouter{
        &routes.UserRouter{},
    },
    Middlewares: []lava.Middleware{
        middlewares.NewLoggerMiddleware(log),
    },
    // 其他参数...
})
```

## 12. 下一步

### 12.1 学习更多功能

- **隧道系统**：了解如何使用 Lava 的隧道系统进行内网穿透
- **调度器**：学习如何使用 Lava 的调度器进行任务调度
- **服务发现**：了解如何集成服务发现机制
- **配置中心**：学习如何使用配置中心管理配置

### 12.2 查看示例代码

Lava 提供了丰富的示例代码：
- `internal/examples/grpcweb` - gRPC Web 示例
- `internal/examples/scheduler` - 调度器示例
- `internal/examples/tunnel` - 隧道系统示例
- `internal/examples/fileserver` - 文件服务器示例

### 12.3 参考文档

- **架构文档**：了解 Lava 的整体架构
- **设计文档**：了解 Lava 的设计原则
- **功能模块说明**：了解各个模块的详细功能
- **开发者指南**：学习如何为 Lava 贡献代码

## 13. 总结

通过本快速开始文档，你已经学会了：

1. **环境搭建**：安装了 Go、Task 等必要工具
2. **创建服务**：创建了 HTTP 和 gRPC 服务
3. **添加中间件**：实现了自定义中间件
4. **使用命令行工具**：使用了 Lava 的命令行工具
5. **调试和监控**：访问了调试接口和指标端点
6. **部署服务**：使用 Docker 和 Kubernetes 部署服务

Lava 框架提供了一套完整的微服务开发解决方案，通过统一的抽象和丰富的功能，大大简化了微服务的开发和运维。现在你可以开始使用 Lava 构建自己的微服务系统了！