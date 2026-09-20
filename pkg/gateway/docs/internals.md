# Gateway 实现细节

本文档介绍 Gateway 模块的内部实现细节，适合需要深入了解或扩展 Gateway 的开发者。

## 统一调度器（Dispatcher）

`Dispatcher`（`dispatcher.go`）是连接「前端流」与「后端连接」的核心泵，借鉴自 vanguard-go 的 transcoder 设计。它把所有协议前端归一化后的 `grpc.ServerStream`，按方法的流类型对接到后端 `grpc.ClientConnInterface`。

### 入口

```go
func (d *Dispatcher) Dispatch(
    ctx context.Context,
    backend Backend,          // = grpc.ClientConnInterface（Mux）
    frontend FrontendStream,  // = grpc.ServerStream（streamHTTP/streamWS/...）
    op *Operation,            // 调度 SSOT（注册时固化；LookupOperation / findMethod）
    in any,                   // unary/server-stream 的预读请求；client/bidi 为 nil
) (header, trailer metadata.MD, err error)
```

### 注册表

- **`Operation`**：调度单一事实来源（FullMethod / 消息类型 / StreamDesc / Meta），在 `registerRouter` 时写入 `methodWrapper.op`。
- **`handlers` / `customOperationNames`**：按 FullMethod（及 RpcMeta.Name）索引到内部 `methodWrapper`（含 codec/proxy 绑定）。
- **`routerTree`**：仅 HTTP 路径 → FullMethod 索引，不复制 schema。

### 四种流模式

| 模式 | 判定（`op.StreamDesc`） | 实现要点 |
| --- | --- | --- |
| Unary | `nil` | `backend.Invoke` → `frontend.SendMsg`，回传 header/trailer |
| Server-Stream | `ServerStreams && !ClientStreams` | `NewStream` → `SendMsg(in)` + `CloseSend` → 循环 `RecvMsg`/`SendMsg`，首帧前发送 header |
| Client-Stream | `ClientStreams && !ServerStreams` | 循环 `frontend.RecvMsg` → `localStream.SendMsg`，`CloseSend` 后取单次响应 |
| Bidi | `ClientStreams && ServerStreams` | 启动 `pumpFrontendToBackend` 与 `pumpBackendToFrontend` 双向泵（`dispatcher.go`），`select` 等待任一方向结束 |

> Bidi 泵默认**不调用** `ClientStream.Header()`：inprocgrpc 后端可能不发送显式 header 帧，此时 `Header()` 会阻塞。各前端在首次 `SendMsg` 时自行发送响应 header。远程透明代理（`TransparentHandler`）通过 `WithPropagateBackendHeaders` 开启首帧 header 转发；已并入 `Dispatcher`，不再维护独立 `forward*` 泵。

### 入口与请求预读约定

各前端的统一入口是 `Mux.DispatchFrontend`（HTTP/gRPC-Web / native gRPC / WebSocket 以及 zrpc 的流式分支）：它按流模式决定是否预读请求，再把 `Dispatch` 包在 RPC 中间件里执行。zrpc 的 unary 分支请求体已在消息里拿到，直接调 `Mux.Dispatch`。`Dispatcher.DispatchFrontend` 是不含中间件的同一套预读+调度逻辑，供透明代理后端（`stream.proxy.go`）复用。

- **Unary / Server-Stream**：先 `RecvMsg` 读入请求消息，作为 `in` 传给 `Dispatch`。
- **Client-Stream / Bidi**：`in` 传 `nil`，由泵内部通过 `frontend.RecvMsg` 持续读取，直到返回 `io.EOF`。

> HTTP/gRPC-Web 前端（`httpFrontend`）在进入调度前会拒绝 `ClientStreams` 方法（返回 `codes.Unimplemented`），因为单次 HTTP 请求体无法可靠表达 client/bidi 多消息语义；请改用 WebSocket 或 Native gRPC。

### 拦截器

- `UseRPCMiddleware`：包装整段 `Mux.Dispatch`（所有流模式，本地与 proxy 一致）；`Mux.DispatchFrontend` 预读请求后才进入它，因此预读失败（请求体解不开）不会被中间件看到。unary/server-stream 请求体见 `IncomingPayload`。
- `UseBackendUnaryInterceptor` / `UseBackendStreamInterceptor`：挂在 Backend 边界，本地与 proxy 共用（见 `backend.go`）。
- `SetUnaryInterceptor` / `SetStreamInterceptor`：挂在 `inprocgrpc.Channel`，只影响 `RegisterService` 本地实现（兼容层）。

目标契约与演进见 [design-evolution.md](design-evolution.md)。

各前端只需实现 `grpc.ServerStream`（编解码/帧处理），即可复用以上全部流模式，这是「底层 handler 注册一次、多协议复用」的关键。

## 指标

`pkg/gateway` 自己不产指标（不依赖 metrics 包）。RPC 指标由 `pkg/middleware/metric` 在 RPC 中间件链内产生——`servers/gatewayserver` 把它装进 `UseRPCMiddleware` 的链——所以 HTTP/gRPC-Web、native gRPC、WebSocket 四条前端与本地、proxy 两种后端共用同一份计数。

| 序列 | 类型 | label |
|------|------|-------|
| `lava_rpc_total` | counter | `side` `kind` `service` `method` `stream` `proto` |
| `lava_rpc_failed_total` | counter | 同上，另加 `code` |
| `lava_rpc_handling_seconds` | histogram | 同 `lava_rpc_total` |

- `side`、`kind` 不能省：这个 middleware 同时被 `servers/zrpcs`、`clients/zrpcc`、`clients/grpcc`、`clients/resty` 安装，没有它们，客户端调用会和服务端调用落进同一条序列。
- `proto` 是 content type 归一后的闭集（`grpc` / `grpc-web` / `json` / `other`）。content type 来自调用方的 metadata，原样当 label 就是无界基数。
- 抓取路径是 `/debug/metrics`：`core/metrics/drivers/prometheus` 把 reporter 的 handler 注册在 debug app 上，而各 server 把 debug app 挂在 `/debug` 下。

链之外的事件不进指标，只有日志：路由未命中（`GetRouterTarget`）、`httpFrontend` 的 426 拒绝、`Mux.DispatchFrontend` 的预读失败，都在进入 `Mux.Dispatch` 之前返回；`gateway.TransparentHandler` 用的是免中间件的 `Dispatcher.DispatchFrontend`，那条路径连 accesslog 都不经过。`servers/https` 的 REST 路由与 `clients/resty` 同样不产指标（它们的操作名是 `VERB /path`）。时延从链内起算，不含路由匹配与响应成帧。

## 路径解析

Gateway 使用 [participle](https://github.com/alecthomas/participle) 解析 HTTP Rule 路径模板。

### 解析流程

1. **词法分析**：将路径字符串分解为 tokens
2. **语法解析**：根据 HTTP Rule 语法规则构建路径 AST
3. **路径规范化**：提取路径变量、通配符、动词等信息
4. **路由树构建**：将解析后的路径信息添加到路由树中

### 支持的路径元素

- 字面量路径段（如 `/v1/users`）
- 路径变量（`{field}` 或 `{field.subfield}`）
- 带模式的路径变量（`{field=pattern}`）
- 单段通配符（`*`）
- 多段通配符（`**`，贪婪匹配）
- 动词（`:verb`，如 `/users/{id}:get`）

### 匹配优先级

1. **精确匹配**：完全匹配的路径段优先
2. **路径变量**：`{field}` 形式的变量匹配
3. **单段通配符**：`*` 匹配单个路径段
4. **多段通配符**：`**` 贪婪匹配剩余所有路径段

### 性能特性

- **时间复杂度**：O(d)，d 是路径深度
- **空间复杂度**：O(n)，n 是路由数量
- **内存优化**：使用前缀树共享公共路径前缀

## 进程内调用

Gateway 使用 [inprocgrpc](https://github.com/fullstorydev/grpchan) 实现进程内的 gRPC 调用。

### 优势

- **零网络开销**：进程内直接调用
- **类型安全**：编译时类型检查
- **调试友好**：可以直接调试和堆栈跟踪
- **性能最优**：避免网络延迟和协议开销

### 实现方式

```go
localClient := new(inprocgrpc.Channel)
localClient.RegisterService(sd, ss)  // 注册服务到进程内通道
```

## 查询参数处理

### 处理流程

1. 从 HTTP 请求 URL 中提取查询参数
2. 合并路径变量
3. 根据字段名和 JSON 名称匹配 Protobuf 字段
4. 执行类型转换
5. 处理数组参数

### 特性

- **嵌套字段支持**：`user.profile.name`
- **类型自动转换**：string → int/bool/double
- **数组参数**：`?ids=1&ids=2&ids=3`
- **默认值处理**：支持 Protobuf 默认值

## 元数据转换

### HTTP → gRPC

- 过滤保留的 HTTP 头（`Content-Type`、`User-Agent`、`grpc-*`）
- 将 HTTP 头转换为小写的 gRPC metadata key
- 处理二进制元数据（`-bin` 后缀，base64 解码）

### gRPC → HTTP

- 将 gRPC header/trailer 转换为 HTTP 响应头
- 二进制元数据使用 base64 编码，添加 `-bin` 后缀

### 保留的 HTTP 头

- `Content-Type`
- `User-Agent`
- `grpc-*` 相关的头

## 错误码映射

HTTP/JSON 前端通过 `HTTPStatusFromCode`（`grpccodes.go`）将 gRPC status 写成对应 HTTP 状态码，unary 响应体为 `{"code":N,"message":"..."}`。

server-stream（NDJSON）的 200 响应头在 `Dispatch` 失败之前就已写出，状态码改不了了，因此在流的末尾追加一行 `{"error":{"code":N,"message":"..."}}`。包一层 `error` 是为了和数据的 `{code,message}` 区分：schemaless（`google.protobuf.Struct`）流完全可能发出后者的数据行。同理会一并丢弃的还有响应 header/trailer 中的自定义 metadata（该框架无法送达），实现只记一条 Debug 日志。

gRPC-Web 前端把 `grpc-status` / `grpc-message` 放进 trailer 帧返回（key 一律小写，`http.Header` 的规范化会把它变成首字母大写，客户端就读不到了）；即使没有响应体也会强制写出 trailer。

完整的 gRPC 错误码到 HTTP 状态码映射：

| gRPC Code | HTTP Status | 说明 |
|-----------|-------------|------|
| OK | 200 | 成功 |
| Canceled | 499 | 客户端取消 |
| Unknown | 500 | 未知错误 |
| InvalidArgument | 400 | 无效参数 |
| DeadlineExceeded | 504 | 超时 |
| NotFound | 404 | 未找到 |
| AlreadyExists | 409 | 已存在 |
| PermissionDenied | 403 | 权限不足 |
| ResourceExhausted | 429 | 资源耗尽 |
| FailedPrecondition | 400 | 前置条件失败 |
| Aborted | 409 | 被中止 |
| OutOfRange | 400 | 超出范围 |
| Unimplemented | 501 | 未实现 |
| Internal | 500 | 内部错误 |
| Unavailable | 503 | 服务不可用 |
| DataLoss | 500 | 数据丢失 |
| Unauthenticated | 401 | 未认证 |

## gRPC Web 实现

### 协议检测

通过 `Content-Type` 头检测 gRPC Web 请求：

```go
func isWebRequestFromContentType(ct, method string) (typ, enc string, ok bool) {
    if !strings.HasPrefix(ct, grpcWeb) || method != http.MethodPost {
        return "", "", false
    }
    typ, enc, ok = strings.Cut(ct, "+")
    if !ok {
        enc = "proto"
    }
    ok = typ == grpcWeb || typ == grpcWebText
    return typ, enc, ok
}
```

### fiberWebWriter

专门为 Fiber 框架设计的 gRPC Web 响应写入器：

```go
type fiberWebWriter struct {
    ctx         fiber.Ctx
    fctx        *fasthttp.RequestCtx // 直播流使用，Fiber ctx 会被池化
    resp        io.Writer
    respCloser  io.Closer   // grpc-web-text 的流式 base64 编码器
    flushWriter http.Flusher
    typ         string      // grpcWeb 或 grpcWebText
    enc         string      // proto 或 json
    wroteHeader bool

    // 响应头可能已提交，这些字段暂存 trailer 内容，由 writeTrailer 统一写出
    errCode      *uint32
    errMessage   string
    extraTrailer http.Header

    headersCommitted bool // 直播流：响应头已序列化，只能写 trailer 帧
}
```

**主要功能：**
- 设置正确的 `Content-Type` 响应头
- 写入 gRPC 数据帧
- 写入 Trailer 帧（包含 grpc-status）

### Trailer 帧格式

`writeTrailer` 先把 trailer 收集到一个 key 全部小写的 `http.Header`，来源有三处：

1. 响应头里 `grpc-` 前缀的项，由 `isGRPCWebTrailerHeader` 过滤掉 `grpc-encoding`、`grpc-accept-encoding`、`grpc-timeout`、`grpc-message-type`（它们是 header，不该在 trailer 里重复）；
2. `addTrailers` 排入的自定义 metadata；
3. `markErrorTrailer` 记录的 `grpc-status` / `grpc-message`，缺省时补一条 `grpc-status: 0`。

随后按普通 gRPC 帧写出，只是 flags 用 trailer 标记（`0x80`，即最高位）：

```go
head := []byte{grpcFrameTrailerFlag, 0, 0, 0, 0}
binary.BigEndian.PutUint32(head[1:grpcFrameHeaderSize], uint32(buf.Len()))
```

`flushWithTrailer` 即使 trailer 写入失败也会 `Close` 编码器并 `Flush`：响应已经丢了，但不能把流式编码器留在打开状态。

## 流式处理实现

### streamHTTP

基于 Fiber Context 的 HTTP 请求/响应流：

**RecvMsg 流程：**
1. 检查 HTTP 方法是否允许请求体
2. 定位请求体字段（根据 body 规则）
3. 执行请求拦截器
4. 解析 gRPC 帧（如果是 gRPC Content-Type）
5. JSON/Protobuf 解码
6. 合并路径变量和查询参数

**SendMsg 流程：**
1. 定位响应字段（根据 response_body 规则）
2. 执行响应拦截器
3. JSON/Protobuf 编码
4. 添加 gRPC 帧头（如果是 gRPC Content-Type）
5. 写入响应

### gRPC 帧格式

```
[1 byte: flags] [4 bytes: length (big-endian)] [message bytes]
```

- flags: `0x00` = 数据帧，`0x01` = 压缩帧（`grpc-encoding` 协商），`0x80` = trailer 帧
- length: 消息长度（大端序）
- message: 实际的 protobuf 消息

## FieldMask 支持

### FieldMaskFromRequestBody

从 JSON 请求体自动生成 FieldMask：

```go
func FieldMaskFromRequestBody(r io.Reader, msg proto.Message) (*fieldmaskpb.FieldMask, error)
```

**特性：**
- 支持嵌套字段路径
- 自动处理 `google.protobuf.Struct`
- 支持 `google.protobuf.Any` 类型

**使用场景：**
- 部分字段更新（PATCH 请求）
- 减少响应数据量
- GraphQL 风格的字段选择

## 参考资料

- [Google API HTTP Annotation](https://cloud.google.com/service-infrastructure/docs/service-management/reference/rpc/google.api)
- [gRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway)
- [gRPC Web Protocol](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)
- [AIP-123: Resource-oriented design](https://google.aip.dev/123)
