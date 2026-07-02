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
    op *Operation,            // 方法元信息
    in any,                   // unary/server-stream 的预读请求；client/bidi 为 nil
) (header, trailer metadata.MD, err error)
```

### 四种流模式

| 模式 | 判定（`op.StreamDesc`） | 实现要点 |
| --- | --- | --- |
| Unary | `nil` | `backend.Invoke` → `frontend.SendMsg`，回传 header/trailer |
| Server-Stream | `ServerStreams && !ClientStreams` | `NewStream` → `SendMsg(in)` + `CloseSend` → 循环 `RecvMsg`/`SendMsg`，首帧前发送 header |
| Client-Stream | `ClientStreams && !ServerStreams` | 循环 `frontend.RecvMsg` → `localStream.SendMsg`，`CloseSend` 后取单次响应 |
| Bidi | `ClientStreams && ServerStreams` | 启动 `pumpFrontendToBackend` 与 `pumpBackendToFrontend` 双向泵（`dispatcher.go`），`select` 等待任一方向结束 |

> Bidi 泵刻意**不调用** `ClientStream.Header()`：inprocgrpc 后端可能不发送显式 header 帧，此时 `Header()` 会阻塞等待一个永不到来的帧。各前端在首次 `SendMsg` 时会自行发送响应 header，因此无需在泵内转发后端 header。`stream.proxy.go` 中的 `TransparentHandler` / `forwardClientToServer` 仍保留给独立的透明代理场景与单测使用。

### 入口与请求预读约定

`Dispatcher.DispatchFrontend` 是各前端的统一入口（native gRPC / WebSocket / zrpc 均使用）：它根据流模式自动决定是否预读请求，再委托给 `Dispatch`。

- **Unary / Server-Stream**：先 `RecvMsg` 读入请求消息，作为 `in` 传给 `Dispatch`。
- **Client-Stream / Bidi**：`in` 传 `nil`，由泵内部通过 `frontend.RecvMsg` 持续读取，直到返回 `io.EOF`。

各前端只需实现 `grpc.ServerStream`（编解码/帧处理），即可复用以上全部流模式，这是「底层 handler 注册一次、多协议复用」的关键。

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
func isWebRequestFromContentType(ct, method string) (typ string, enc string, ok bool) {
    if !strings.HasPrefix(ct, "application/grpc-web") || method != http.MethodPost {
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
    ctx         *fiber.Ctx
    resp        io.Writer
    typ         string // grpcWeb or grpcWebText
    enc         string // proto or json
    wroteHeader bool
    wroteResp   bool
}
```

**主要功能：**
- 设置正确的 `Content-Type` 响应头
- 写入 gRPC 数据帧
- 写入 Trailer 帧（包含 grpc-status）

### Trailer 帧格式

```go
func (w *fiberWebWriter) writeTrailer() error {
    // 收集 grpc-* headers
    tr := make(http.Header)
    w.ctx.Response().Header.VisitAll(func(key, value []byte) {
        k := string(key)
        if strings.HasPrefix(strings.ToLower(k), "grpc-") {
            tr[strings.ToLower(k)] = []string{string(value)}
        }
    })
    // 默认 grpc-status
    if tr.Get("grpc-status") == "" {
        tr.Set("grpc-status", "0")
    }
    
    // 写入 trailer 帧
    head := []byte{1 << 7, 0, 0, 0, 0} // MSB=1 表示 trailer
    binary.BigEndian.PutUint32(head[1:5], uint32(buf.Len()))
    w.resp.Write(head)
    w.resp.Write(buf.Bytes())
}
```

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

- flags: `0x00` = 数据帧，`0x80` = trailer 帧
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
