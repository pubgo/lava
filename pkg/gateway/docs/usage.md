# Gateway 使用指南

本文档介绍如何使用 Gateway 模块，包括服务注册、路由配置、中间件等。

## 快速开始

### 1. 定义 Protobuf 服务

```protobuf
syntax = "proto3";

package example.v1;

import "google/api/annotations.proto";

service UserService {
  rpc GetUser(GetUserRequest) returns (User) {
    option (google.api.http) = {
      get: "/v1/users/{user_id}"
    };
  }
  
  rpc CreateUser(CreateUserRequest) returns (User) {
    option (google.api.http) = {
      post: "/v1/users"
      body: "user"
    };
  }
}

message GetUserRequest {
  string user_id = 1;
}

message CreateUserRequest {
  User user = 1;
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
}
```

### 2. 实现服务

```go
type userService struct {
    pb.UnimplementedUserServiceServer
}

func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    return &pb.User{
        Id:    req.UserId,
        Name:  "John Doe",
        Email: "john@example.com",
    }, nil
}

func (s *userService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
    return req.User, nil
}
```

### 3. 配置 Gateway

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "github.com/pubgo/lava/v2/pkg/gateway"
    pb "your/proto/package"
)

func main() {
    // 创建 Gateway
    mux := gateway.NewMux()

    // 注册服务
    mux.RegisterService(&pb.UserService_ServiceDesc, &userService{})

    // 创建 Fiber 应用
    app := fiber.New()
    app.All("/v1/*", mux.Handler)

    app.Listen(":8080")
}
```

### 4. 测试调用

```bash
# GET 请求
curl http://localhost:8080/v1/users/123

# POST 请求
curl -X POST http://localhost:8080/v1/users \
  -H "Content-Type: application/json" \
  -d '{"user": {"name": "Jane", "email": "jane@example.com"}}'
```

## 服务注册

### 本地服务

本地服务运行在同一进程中，通过进程内通道调用：

```go
mux := gateway.NewMux()
mux.RegisterService(&pb.UserService_ServiceDesc, &userServiceImpl{})
```

### 代理服务

代理服务将请求转发到远程 gRPC 服务：

```go
// 连接到远程服务
conn, _ := grpc.Dial("remote-service:50051", grpc.WithInsecure())

// 注册代理
mux.RegisterProxy(&pb.UserService_ServiceDesc, router, conn)
```

## HTTP Rule 配置

### 请求体映射

```protobuf
// 整个消息作为请求体
rpc CreateUser(CreateUserRequest) returns (User) {
  option (google.api.http) = {
    post: "/v1/users"
    body: "*"
  };
}

// 指定字段作为请求体
rpc UpdateUser(UpdateUserRequest) returns (User) {
  option (google.api.http) = {
    patch: "/v1/users/{user_id}"
    body: "user"  // 只有 user 字段从请求体解析
  };
}
```

### 响应体映射

```protobuf
// 只返回指定字段
rpc GetUser(GetUserRequest) returns (GetUserResponse) {
  option (google.api.http) = {
    get: "/v1/users/{user_id}"
    response_body: "user"  // 只返回 user 字段
  };
}
```

### 多路由绑定

```protobuf
rpc UpdateUser(UpdateUserRequest) returns (User) {
  option (google.api.http) = {
    patch: "/v1/users/{user.id}"
    body: "user"
    additional_bindings: {
      put: "/v1/users/{user.id}"
      body: "user"
    }
  };
}
```

## 路径匹配规则

| 模式 | 示例 | 说明 |
|------|------|------|
| `{field}` | `/users/{id}` | 匹配单个路径段 |
| `{field.subfield}` | `/users/{user.id}` | 嵌套字段 |
| `{field=pattern}` | `{id=projects/*/users/*}` | 带模式的变量 |
| `*` | `/files/*` | 单段通配符 |
| `**` | `/files/**` | 多段通配符 |
| `:verb` | `/users/{id}:get` | 动词后缀 |

## 中间件

### Unary 拦截器

```go
mux.SetUnaryInterceptor(func(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    // 前置处理（日志、认证等）
    start := time.Now()
    
    // 调用实际处理
    resp, err := handler(ctx, req)
    
    // 后置处理（记录耗时等）
    log.Printf("Method: %s, Duration: %v", info.FullMethod, time.Since(start))
    
    return resp, err
})
```

### Stream 拦截器

```go
mux.SetStreamInterceptor(func(
    srv interface{},
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error {
    log.Printf("Starting stream: %s", info.FullMethod)
    return handler(srv, ss)
})
```

## 请求/响应拦截器

### 请求解码器

```go
mux.SetRequestDecoder(
    protoreflect.FullName("example.UserRequest"),
    func(ctx *fiber.Ctx, msg proto.Message) error {
        // 从 HTTP 头读取用户信息
        userID := ctx.Get("X-User-ID")
        if userID != "" {
            msg.ProtoReflect().Set(
                msg.ProtoReflect().Descriptor().Fields().ByName("user_id"),
                protoreflect.ValueOfString(userID),
            )
        }
        return nil
    },
)
```

### 响应编码器

```go
mux.SetResponseEncoder(
    protoreflect.FullName("example.UserResponse"),
    func(ctx *fiber.Ctx, msg proto.Message) error {
        // 添加自定义响应头
        ctx.Set("X-Custom-Header", "value")
        return nil
    },
)
```

## 配置选项

```go
mux := gateway.NewMux(
    // 最大接收消息大小（默认 4MB）
    gateway.MaxReceiveMessageSizeOption(4 * 1024 * 1024),
    
    // 最大发送消息大小
    gateway.MaxSendMessageSizeOption(4 * 1024 * 1024),
    
    // 连接超时（默认 120s）
    gateway.ConnectionTimeoutOption(120 * time.Second),
    
    // 自定义编解码器
    gateway.CodecOption("application/xml", xmlCodec),
    
    // 自定义压缩器
    gateway.CompressorOption("gzip", gzipCompressor),
)
```

## 错误处理

Gateway 自动将 gRPC 错误码映射为 HTTP 状态码：

| gRPC Code | HTTP Status |
|-----------|-------------|
| OK | 200 |
| InvalidArgument | 400 |
| Unauthenticated | 401 |
| PermissionDenied | 403 |
| NotFound | 404 |
| AlreadyExists | 409 |
| ResourceExhausted | 429 |
| Internal | 500 |
| Unavailable | 503 |

在服务中返回 gRPC 错误：

```go
import "google.golang.org/grpc/status"

func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }
    // ...
}
```

## 最佳实践

1. **使用 HTTP Rule 注解**：在 Protobuf 中定义路由，而不是手动注册
2. **合理使用 body 映射**：只映射需要的字段，减少数据传输
3. **使用拦截器**：统一处理日志、认证、监控等横切关注点
4. **错误处理**：使用标准的 gRPC 错误码，Gateway 会自动映射
5. **进程内调用**：优先使用本地服务注册，避免网络开销
