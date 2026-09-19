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

## 多协议前端

底层 gRPC handler 注册一次后，可同时挂载多种协议前端（详见[架构设计](architecture.md)）：

```go
mux := gateway.NewMux()
mux.RegisterService(&pb.UserService_ServiceDesc, &userServiceImpl{})

// 前端 1：HTTP/REST + gRPC-Web（Fiber/fasthttp）
app := fiber.New()
app.All("/api/*", mux.Handler)

// 前端 2：WebSocket（coder/websocket，必须用 net/http）
go http.ListenAndServe(":8081", mux.WebSocketHandler())

// 前端 3：Native gRPC 透传
grpcServer := grpc.NewServer(mux.GRPCServerOptions()...)

// NATS/zrpc 见 pkg/zrpcbridge（非 gateway 前端）
```

| 前端 | 入口 | 运行栈 | 说明 |
| --- | --- | --- | --- |
| HTTP/REST | `mux.Handler` | Fiber/fasthttp | RESTful + 普通 JSON |
| gRPC-Web | `mux.Handler` | Fiber/fasthttp | 浏览器 gRPC，详见 [gRPC Web](grpcweb.md) |
| WebSocket | `mux.WebSocketHandler()` | net/http | 双向流，详见 [WebSocket](websocket.md) |
| Native gRPC | `mux.GRPCServerOptions()` | grpc.Server | 透传，详见 [Native gRPC](grpcnative.md) |

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
    // 按 Content-Type 覆盖/注册编解码器（默认已含 application/json、application/protobuf）
    gateway.WithCodec("application/json", gateway.CodecJSON{}),
)
```

> 消息压缩：HTTP/gRPC-Web 帧路径支持 `grpc-encoding` / `grpc-accept-encoding` 协商（默认注册 `gzip`）。请求帧压缩标志为 `0x01` 时按请求编码解压；响应在客户端接受时压缩并回写 `Grpc-Encoding`。

## 错误处理

HTTP/JSON 前端将 gRPC 错误码映射为 HTTP 状态码，并返回 JSON：`{"code":N,"message":"..."}`。
gRPC-Web 则写入 `grpc-status` / `grpc-message` 并由 trailer 帧带回客户端。
server-stream（NDJSON）的 200 已经发出，无法再改状态码，失败时流末尾追加一行 `{"error":{"code":N,"message":"..."}}`（包一层 `error` 是为了和可能长成 `{code,message}` 的数据行区分）。

| gRPC Code | HTTP Status |
|-----------|-------------|
| OK | 200 |
| Canceled | 499 |
| InvalidArgument | 400 |
| Unauthenticated | 401 |
| PermissionDenied | 403 |
| NotFound | 404 |
| AlreadyExists | 409 |
| ResourceExhausted | 429 |
| Unimplemented | 501（HTTP 前端拒绝 client-stream / bidi） |
| Internal | 500 |
| DeadlineExceeded | 504 |
| Unavailable | 503 |

完整映射见 [实现细节](internals.md#错误码映射)。

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

### 拦截器作用范围

- **`UseRPCMiddleware`**：包装整段 `Dispatch` / `DispatchFrontend`（unary + 全部流模式，本地与 proxy 一致）。需要完整 RPC 生命周期（如 lava Middleware）挂这里；请求体见 `IncomingPayload`。
- **`UseBackendUnaryInterceptor` / `UseBackendStreamInterceptor`**：包装 `Mux.Invoke` / `NewStream`，**本地与 `RegisterProxy` 都会经过**。调用边界横切挂这里。
- **`SetUnaryInterceptor` / `SetStreamInterceptor`**：仅作用于 `RegisterService` 的进程内 handler（`inprocgrpc` server interceptor），兼容层；新横切优先 RPC / Backend 链（见 [design-evolution.md](design-evolution.md)）。

> **直接以 Mux 当 gRPC client 的调用**（`pb.NewXxxClient(mux)`）只经过 Backend 链，**不经过** `UseRPCMiddleware`：RPC 中间件要在整段流结束后收尾，而 `Invoke`/`NewStream` 在流开始时就返回了。v2 把 lava 中间件挂在 inproc 通道上，因此这类调用也被覆盖；迁移时若依赖这一点，把横切改挂 `UseBackend*`（或让调用方走前端 / `DispatchFrontend`）。

## 最佳实践

1. **使用 HTTP Rule 注解**：在 Protobuf 中定义路由，而不是手动注册
2. **合理使用 body 映射**：只映射需要的字段，减少数据传输
3. **使用拦截器**：完整 RPC 横切用 `UseRPCMiddleware`；调用边界用 `UseBackend*`；本地 handler 细节可用 `SetUnary/StreamInterceptor`
4. **错误处理**：使用标准的 gRPC 错误码，HTTP/JSON 前端会自动映射状态码
5. **进程内调用**：优先使用本地服务注册，避免网络开销
6. **流式 RPC**：client/bidi 请走 WebSocket 或 Native gRPC，不要依赖 HTTP/REST
