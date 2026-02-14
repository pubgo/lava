# Gateway

Gateway 是一个 gRPC Gateway 实现，提供 HTTP/JSON 到 gRPC 的协议转换功能。它支持服务注册、中间件、HTTP Rule 解析等特性，让开发者可以轻松地将 gRPC 服务暴露为 RESTful API。

## 核心特性

- **HTTP Rule 解析**：支持 `google.api.http` 注解，自动解析 RESTful 路径模板
- **协议转换**：自动处理 HTTP/JSON 与 gRPC/Protobuf 之间的双向转换
- **gRPC Web 支持**：允许浏览器直接调用 gRPC 服务
- **服务注册**：支持本地服务和代理服务的注册
- **中间件支持**：提供 Unary 和 Stream 拦截器
- **自定义编解码**：支持 JSON、Protobuf 等多种编码格式
- **错误映射**：自动将 gRPC 错误码映射为 HTTP 状态码

## 快速开始

### 1. 定义 Proto 文件

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
      body: "*"
    };
  }
}
```

### 2. 实现并注册服务

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
    mux.RegisterService(&pb.UserService_ServiceDesc, &userServiceImpl{})

    // 创建 Fiber 应用
    app := fiber.New()
    app.All("/v1/*", mux.Handler)

    app.Listen(":8080")
}
```

### 3. 调用服务

```bash
# GET 请求
curl http://localhost:8080/v1/users/123

# POST 请求
curl -X POST http://localhost:8080/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com"}'
```

## 文档

| 文档 | 说明 |
|------|------|
| [使用指南](docs/usage.md) | 服务注册、路由配置、中间件、错误处理等 |
| [gRPC Web](docs/grpcweb.md) | 浏览器端 gRPC Web 集成 |
| [架构设计](docs/architecture.md) | 核心组件、数据结构、处理流程 |
| [实现细节](docs/internals.md) | 路径解析、元数据转换、流式处理等 |

## 支持的协议

| 协议 | Content-Type | 说明 |
|------|--------------|------|
| HTTP/JSON | `application/json` | RESTful API |
| gRPC Web | `application/grpc-web+proto` | 浏览器 gRPC (二进制) |
| gRPC Web Text | `application/grpc-web-text+proto` | 浏览器 gRPC (Base64) |

## 路径匹配

| 模式 | 示例 | 说明 |
|------|------|------|
| `{field}` | `/users/{id}` | 路径变量 |
| `*` | `/files/*` | 单段通配符 |
| `**` | `/files/**` | 多段通配符 |
| `:verb` | `/users/{id}:get` | 动词后缀 |

## 错误码映射

| gRPC Code | HTTP Status |
|-----------|-------------|
| OK | 200 |
| InvalidArgument | 400 |
| Unauthenticated | 401 |
| PermissionDenied | 403 |
| NotFound | 404 |
| Internal | 500 |

完整映射表见 [实现细节](docs/internals.md#错误码映射)。

## 示例

- [gRPC Web 示例](../../internal/examples/grpcweb/) - 完整的 gRPC Web 前后端示例

## 参考

- [Google API HTTP Annotation](https://cloud.google.com/service-infrastructure/docs/service-management/reference/rpc/google.api)
- [gRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway)
- [gRPC Web Protocol](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)
