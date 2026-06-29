# Gateway

Gateway 是一个 gRPC Gateway 实现，提供 HTTP/JSON 到 gRPC 的协议转换功能。它支持服务注册、中间件、HTTP Rule 解析等特性，让开发者可以轻松地将 gRPC 服务暴露为 RESTful API。

## 核心特性

- **分层架构**：前端协议层（Frontend）/ 统一调度层（Dispatcher）/ 后端 gRPC 层（Backend）三层解耦，底层 gRPC handler 注册一次，多种协议前端复用
- **统一调度**：`Dispatcher` 统一处理 Unary / Server-Stream / Client-Stream / Bidi 四种流模式
- **多协议前端**：HTTP/REST、gRPC-Web、WebSocket、Native gRPC、NATS/zrpc 共享同一套后端 handler
- **HTTP Rule 解析**：支持 `google.api.http` 注解，自动解析 RESTful 路径模板
- **协议转换**：自动处理 HTTP/JSON 与 gRPC/Protobuf 之间的双向转换
- **gRPC Web 支持**：允许浏览器直接调用 gRPC 服务
- **服务注册**：支持本地服务和代理服务的注册
- **中间件支持**：提供 Unary 和 Stream 拦截器
- **自定义编解码**：支持 JSON、Protobuf 等多种编码格式
- **错误映射**：自动将 gRPC 错误码映射为 HTTP 状态码

## 分层架构概览

```
[ 前端协议层 Frontend ]      [ 核心调度层 Dispatcher ]      [ 后端 gRPC 层 Backend ]
  HTTP/REST                                                   inprocgrpc.Channel
  gRPC-Web          ──►   归一化为 grpc.ServerStream    ──►   （本地 handler）
  WebSocket                统一泵：unary/server/             remoteProxyCli
  Native gRPC              client/bidi 四种流模式             （远程代理）
  NATS/zrpc
```

设计灵感来自 [connectrpc/vanguard-go](https://github.com/connectrpc/vanguard-go)：所有前端协议最终归一化为 gRPC 语义的 `ServerStream`，由统一的 `Dispatcher` 对接后端，从而让「底层注册一次的 gRPC handler」服务于多种上层协议。详见 [架构设计](docs/architecture.md)。

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

| 文档                             | 说明                                   |
| -------------------------------- | -------------------------------------- |
| [使用指南](docs/usage.md)        | 服务注册、路由配置、中间件、错误处理等 |
| [gRPC Web](docs/grpcweb.md)      | 浏览器端 gRPC Web 集成                 |
| [WebSocket](docs/websocket.md)   | 基于 coder/websocket 的 WebSocket 前端 |
| [Native gRPC](docs/grpcnative.md) | 原生 gRPC 透传，RegisterService 一次多协议复用 |
| [NATS/zrpc](docs/grpcnats.md)     | NATS 队列订阅前端，protobuf 帧               |
| [架构设计](docs/architecture.md) | 分层架构、核心组件、调度流程           |
| [实现细节](docs/internals.md)    | 路径解析、调度器、元数据转换、流式处理 |
| [部署/TLS](docs/deploy.md)       | 边缘 TLS 终止与 Traefik 多协议路由     |

## 支持的协议

| 协议             | Content-Type / 入口               | 说明                         |
| ---------------- | --------------------------------- | ---------------------------- |
| HTTP/JSON        | `application/json`                | RESTful API                  |
| HTTP/JSON (别名) | `application/grpc-web-json`       | 前端命名兼容（按 JSON 处理） |
| gRPC Web         | `application/grpc-web+proto`      | 浏览器 gRPC (二进制)         |
| gRPC Web Text    | `application/grpc-web-text+proto` | 浏览器 gRPC (Base64)         |
| WebSocket        | `Mux.WebSocketHandler()`          | net/http 监听，支持双向流    |
| Native gRPC      | `Mux.GRPCServerOptions()`         | 标准 grpc.Server 透传        |
| NATS/zrpc        | `Mux.RegisterZrpc()`              | NATS 队列订阅，protobuf 帧   |

> 说明：HTTP/REST 与 gRPC-Web 前端运行在 Fiber/fasthttp 栈上（`Mux.Handler`）；WebSocket 前端基于 coder/websocket，必须运行在标准 `net/http` 栈上（`Mux.WebSocketHandler()`），详见 [WebSocket 文档](docs/websocket.md)。

## 路径匹配

| 模式      | 示例              | 说明       |
| --------- | ----------------- | ---------- |
| `{field}` | `/users/{id}`     | 路径变量   |
| `*`       | `/files/*`        | 单段通配符 |
| `**`      | `/files/**`       | 多段通配符 |
| `:verb`   | `/users/{id}:get` | 动词后缀   |

## 错误码映射

| gRPC Code        | HTTP Status |
| ---------------- | ----------- |
| OK               | 200         |
| InvalidArgument  | 400         |
| Unauthenticated  | 401         |
| PermissionDenied | 403         |
| NotFound         | 404         |
| Internal         | 500         |

完整映射表见 [实现细节](docs/internals.md#错误码映射)。

## 示例

- [gRPC Web 示例](../../internal/examples/grpcweb/) - HTTP/gRPC-Web 前后端示例
- [多协议示例](../../internal/examples/grpcwebsocket/) - 同一套 handler 同时暴露 HTTP/gRPC-Web(:8080)、WebSocket(:8081)、原生 gRPC(:50051)
  - `internal/examples/grpcwebsocket/verify/` 提供自动化验证（先启动 main，再运行 verify）

## 部署与 TLS

Gateway **不内置 HTTPS/TLS**，全程明文（`http` / `h2c`），TLS 在边缘代理（如 Traefik）终止。
三类协议需分别路由，其中原生 gRPC 回源必须用 `h2c`：

| 协议 | 默认端口 | 回源 scheme |
| --- | --- | --- |
| HTTP/REST + gRPC-Web | `http_port`(8080) | `http` |
| WebSocket | `websocket_port`(8081) | `http` |
| 原生 gRPC | `grpc_port`(50051) | `h2c` |

开箱即用的 Traefik 配置见 [`deploy/traefik/`](../../deploy/traefik/)，说明见 [部署/TLS 文档](docs/deploy.md)。

## 参考

- [Google API HTTP Annotation](https://cloud.google.com/service-infrastructure/docs/service-management/reference/rpc/google.api)
- [gRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway)
- [gRPC Web Protocol](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)
