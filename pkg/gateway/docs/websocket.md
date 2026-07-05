# WebSocket 支持

Gateway 提供基于 [coder/websocket](https://github.com/coder/websocket) 的 WebSocket 前端，让 WebSocket 客户端直接调用已注册的 gRPC handler，并天然支持双向流（Bidi Streaming）。

WebSocket 前端是 Gateway 三层架构中的一个「前端协议层」：它把 WebSocket 连接归一化为 `grpc.ServerStream`，再交给统一的 `Dispatcher` 调度到后端 gRPC handler。底层 handler 注册一次，即可同时被 HTTP/REST、gRPC-Web 和 WebSocket 复用。

## 重要前提：运行在 net/http 栈

coder/websocket 依赖标准库的 `http.Hijacker`，**无法在 fasthttp/Fiber 栈上运行**。因此：

- HTTP/REST 与 gRPC-Web 前端：运行在 Fiber/fasthttp 上（`Mux.Handler`）
- WebSocket 前端：运行在标准 `net/http` 上（`Mux.WebSocketHandler()`）

通常做法是为 WebSocket 单开一个 `net/http` 监听端口，或在已有的 net/http server 上挂载该 handler。

## 快速开始

### 1. 服务端配置

```go
package main

import (
    "net/http"

    "github.com/pubgo/lava/v2/pkg/gateway"
    pb "your/proto/package"
)

func main() {
    mux := gateway.NewMux()
    mux.RegisterService(&pb.GreeterService_ServiceDesc, &greeterService{})

    // WebSocket 前端运行在独立的 net/http 端口
    wsHandler := mux.WebSocketHandler(
        gateway.WithWSOriginPatterns("example.com", "*.example.com"),
        // gateway.WithWSInsecureSkipVerify(), // 仅开发期，关闭跨域校验
    )

    http.ListenAndServe(":8081", wsHandler)
}
```

### 2. 浏览器客户端

```javascript
// 路径 = gRPC 全方法名 /<package>.<service>/<method>
const ws = new WebSocket("ws://localhost:8081/example.v1.GreeterService/SayHello");

ws.onopen = () => {
    ws.send(JSON.stringify({ name: "World" }));
};

ws.onmessage = (event) => {
    const resp = JSON.parse(event.data);
    console.log("message:", resp.message);
};

ws.onclose = (event) => {
    console.log("closed:", event.code, event.reason);
};
```

## 协议约定

### 路由

请求路径被解释为 gRPC 全方法名 `/<package>.<service>/<method>`，直接在已注册 handler 中查找。若直查未命中，会回退到 `routerTree` 的 REST 风格路由匹配（此时路径变量与查询参数会合并进首个请求消息）。

WebSocket 握手使用 **GET**，而 `google.api.http` 路由常为 **POST**。回退匹配时会依次尝试 GET、POST、PUT、PATCH、DELETE；也可显式指定：

```
ws://localhost:8081/v1/greeter/hello?http_method=POST
```

### 编码格式

| 选择方式 | 值 | 编码 | WebSocket 帧类型 |
| --- | --- | --- | --- |
| 默认 | — | protojson | Text |
| 查询参数 | `?encoding=proto`（或 `protobuf`/`binary`） | protobuf | Binary |
| 子协议 | `grpc-ws-proto` | protobuf | Binary |
| 子协议 | `grpc-ws-json` | protojson | Text |

示例：

```javascript
// 使用 protobuf 二进制
const ws = new WebSocket(
  "ws://localhost:8081/example.v1.EchoService/Echo?encoding=proto"
);
ws.binaryType = "arraybuffer";
```

### 流模式映射

WebSocket 前端复用统一 `Dispatcher`，因此支持全部四种流模式：

| RPC 类型 | 行为 |
| --- | --- |
| Unary | 客户端发送 1 条消息 → 服务端回 1 条 WS 帧 → 关闭 |
| Server-Stream | 客户端发送 1 条请求 → 服务端连续推送多条 WS 帧 → 关闭 |
| Client-Stream | 客户端流式发送多条 → 正常关闭连接表示结束 → 服务端回 1 条响应 |
| Bidi | 双向并发收发，互不阻塞 |

示例中的 `WatchHello` RPC 演示 **Server-Stream**；`Chat` RPC 演示 **Bidi**（客户端连续发、服务端逐条 echo，发送 `bye` 结束）。

> 客户端的「正常关闭」（WebSocket Close 帧）会被映射为 gRPC 的 `io.EOF`，作为 Client-Stream / Bidi 的客户端发送结束信号。

### Metadata

HTTP 握手请求头会被转换为 gRPC 的 incoming metadata，供服务端通过 `metadata.FromIncomingContext` 读取。

### 关闭码

调度成功后服务端以 `StatusNormalClosure (1000)` 关闭；调度出错时以 `StatusInternalError (1011)` 关闭。WebSocket 没有原生 trailer 概念，`grpc-status` 暂通过关闭码体现。

## 配置选项

| 选项 | 说明 |
| --- | --- |
| `WithWSOriginPatterns(patterns...)` | 允许的 Origin 模式（对应 `AcceptOptions.OriginPatterns`），为空表示仅同源 |
| `WithWSInsecureSkipVerify()` | 关闭 Origin 校验，仅建议开发环境使用 |
| `WithWSSubprotocols(subprotocols...)` | 握手时声明的子协议列表 |

## 实现说明

- `streamWS`（`stream.websocket.go`）实现 `grpc.ServerStream`，负责 WebSocket 帧的编解码与读写。
- `wsFrontend`（`frontend_ws.go`）实现 `http.Handler`，负责握手、解析 Operation、构建 `streamWS` 并调用 `Dispatcher.DispatchFrontend`。
- 通过 REST 注解路径访问时，`streamWS` 会带上匹配到的 `MatchOperation`，因此 `body:"field"` / `response_body` 字段映射对 WebSocket 同样生效；直查 gRPC 全方法名时整条消息即请求/响应体。
- 结束时通过 WebSocket Close 帧回传结构化 gRPC 状态：close code 由 gRPC code 映射，reason 为 JSON `{"grpcStatus":N,"grpcMessage":"..."}`，客户端可解析 `event.reason` 获取状态。

## 在 gateway 中启用

在 `gateway_server` 配置里设置 `websocket_port` 即可在独立端口启动 WebSocket 前端（与 Fiber HTTP 端口分离）：

```yaml
gateway_server:
  websocket_port: 8081
  # 生产环境建议配置 Origin 白名单
  websocket_origin_patterns:
    - "example.com"
    - "*.example.com"
  # 开发环境可临时关闭 Origin 校验（未配置 origin_patterns 时默认跳过校验）
  websocket_insecure_skip_verify: false
  http:
    # ...
```

代码侧也可通过 `gateway.WSOptionsFromConfig` 构建选项：

```go
mux.WebSocketHandler(gateway.WSOptionsFromConfig(gateway.WSConfig{
    OriginPatterns: []string{"example.com"},
})...)
```

服务端会使用同一个 `gateway.Mux` 实例，已注册的 gRPC handler 自动可被 WebSocket 客户端调用。

### Origin 校验（生产环境）

| 配置 | 行为 |
| --- | --- |
| 未配置 `websocket_origin_patterns` 且 `websocket_insecure_skip_verify: false` | **开发默认**：跳过 Origin 校验，启动时输出 WARN 日志 |
| 配置 `websocket_origin_patterns` | 仅允许匹配的 Origin（推荐生产环境） |
| `websocket_insecure_skip_verify: true` | 显式关闭校验，仅用于本地调试 |

生产环境务必配置 `websocket_origin_patterns`，例如：

```yaml
gateway_server:
  websocket_port: 8081
  websocket_origin_patterns:
    - "example.com"
    - "*.example.com"
```

## 完整示例

参见 `internal/examples/grpcwebsocket/` 目录：

```
internal/examples/grpcwebsocket/
├── main.go           # Go 服务端：Fiber(8080) + WebSocket net/http(8081) + gRPC(50051)
├── verify/main.go    # 自动化验证（需先启动 main）
└── static/
    └── index.html    # 浏览器测试页
```

**运行示例：**

```bash
go run ./internal/examples/grpcwebsocket/
open http://localhost:8080/

# 另开终端，自动化验证全部前端
go run ./internal/examples/grpcwebsocket/verify/
```

WebSocket 端点示例：

```
ws://localhost:8081/grpcweb.example.v1.GreeterService/SayHello
```

该示例复用了 `internal/examples/grpcweb/proto` 中的 Greeter 服务定义，底层 handler 注册一次，同时提供 HTTP/gRPC-Web（8080）与 WebSocket（8081）两种前端。

## 当前限制

- ⏳ 没有内置心跳/超时管理，需结合业务在 handler 内处理
- ✅ `grpc-status` / `grpc-message` 通过 Close 帧 reason 以结构化 JSON 回传
- ✅ REST body 字段映射（`body:"field"`）在 WebSocket 流已启用（需经 REST 注解路径访问）
