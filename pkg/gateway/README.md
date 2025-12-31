# Gateway

Gateway 是一个 gRPC Gateway 实现，提供 HTTP/JSON 到 gRPC 的协议转换功能。它支持服务注册、中间件、HTTP Rule 解析等特性，让开发者可以轻松地将 gRPC 服务暴露为 RESTful API。

## 概述

Gateway 模块实现了完整的 gRPC Gateway 功能，允许客户端通过 HTTP/JSON 调用 gRPC 服务。它基于 [Google API HTTP Annotation](https://cloud.google.com/service-infrastructure/docs/service-management/reference/rpc/google.api#google.api.DocumentationRule.FIELDS.string.google.api.DocumentationRule.selector) 规范，支持灵活的路径模板和请求/响应体映射。

### 核心特性

- **HTTP Rule 解析**：支持 `google.api.http` 注解，自动解析 RESTful 路径模板
- **协议转换**：自动处理 HTTP/JSON 与 gRPC/Protobuf 之间的双向转换
- **服务注册**：支持本地服务和代理服务的注册
- **中间件支持**：提供 Unary 和 Stream 拦截器
- **自定义编解码**：支持 JSON、Protobuf 等多种编码格式
- **请求/响应拦截器**：支持针对特定消息类型的自定义编解码逻辑
- **错误映射**：自动将 gRPC 错误码映射为 HTTP 状态码
- **流式支持**：支持 HTTP、WebSocket、in-process、proxy 等多种流类型

### 模块结构

```
pkg/gateway/
├── mux.go              # 核心路由器 Mux，实现 Gateway 接口
├── routertree/         # 路由树实现，负责路径匹配
│   ├── router.go       # 路由树核心逻辑
│   ├── parser.go       # HTTP Rule 路径模板解析器
│   └── lex.go          # 词法分析器
├── codec.go            # 编解码器接口和实现（JSON、Protobuf）
├── stream.go           # Stream 接口定义
├── stream.http.go      # HTTP 流实现
├── stream.websocket.go # WebSocket 流实现
├── stream.inprocess.go # 进程内流实现
├── stream.proxy.go     # 代理流实现
├── context.go          # 上下文和元数据管理
├── util.go             # 工具函数（HTTP Rule 解析、元数据转换等）
├── fieldmask.go        # FieldMask 支持
├── wrapper.go          # 服务和方法包装器
├── grpccodes.go        # gRPC 错误码到 HTTP 状态码映射
├── gatewayutils/       # Gateway 工具函数
│   ├── query_params.go # 查询参数处理
│   └── trie.go         # Trie 树实现
└── internal/           # 内部实现（压缩器等）
```

## 架构设计

### 设计理念

Gateway 模块基于 **Google API HTTP Annotation** 规范，实现了从 HTTP/REST 到 gRPC 的透明转换。其设计遵循以下原则：

1. **声明式路由**：通过 Protobuf 注解定义 HTTP 路由，无需手动编写路由代码
2. **协议透明**：客户端使用标准的 HTTP/JSON，后端使用 gRPC，Gateway 自动处理转换
3. **类型安全**：基于 Protobuf 的类型系统，保证请求/响应的类型安全
4. **可扩展性**：支持自定义编解码器、拦截器、压缩器等扩展点

### 核心组件

```
┌─────────────────────────────────────────────────────────┐
│                      Gateway (Mux)                       │
│             核心路由器，管理整个请求生命周期              │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │ RouterTree   │  │   Codec      │  │  Stream      │ │
│  │ (路径匹配)    │  │  (编解码)    │  │  (流处理)    │ │
│  │              │  │              │  │              │ │
│  │ - 路径解析    │  │ - JSON       │  │ - HTTP       │ │
│  │ - 变量提取    │  │ - Protobuf   │  │ - WebSocket  │ │
│  │ - 路由匹配    │  │ - 自定义     │  │ - InProcess  │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │  Context     │  │  FieldMask   │  │  Wrapper     │ │
│  │ (元数据管理)  │  │  (字段掩码)  │  │  (服务包装)  │ │
│  │              │  │              │  │              │ │
│  │ - HTTP↔gRPC  │  │ - 字段过滤   │  │ - 本地服务   │ │
│  │ - Metadata   │  │ - 路径解析   │  │ - 代理服务   │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 请求处理流程

完整的请求处理流程如下：

```
HTTP Request (JSON)
    │
    ├─> [1. Handler] 接收 Fiber Context
    │
    ├─> [2. RouterTree.Match] 
    │      ├─ 解析 HTTP 方法和路径
    │      ├─ 匹配路由规则（支持通配符、变量、动词）
    │      └─ 提取路径变量和查询参数
    │
    ├─> [3. Method Lookup]
    │      └─ 根据 gRPC 方法名查找对应的 methodWrapper
    │
    ├─> [4. Metadata Conversion]
    │      └─ 将 HTTP 头转换为 gRPC metadata（支持二进制编码）
    │
    ├─> [5. streamHTTP.RecvMsg]
    │      ├─ 根据 body 规则解析请求体（JSON → Protobuf）
    │      ├─ 合并路径变量和查询参数到消息字段
    │      └─ 执行请求拦截器（如果配置）
    │
    ├─> [6. Mux.Invoke]
    │      ├─ 判断是本地服务还是代理服务
    │      ├─ 本地服务：使用 inprocgrpc.Channel 调用
    │      └─ 代理服务：转发到远程 gRPC 客户端
    │
    ├─> [7. gRPC Service Execution]
    │      ├─ 执行 Unary/Stream 拦截器
    │      └─ 调用实际的 gRPC 服务方法
    │
    ├─> [8. streamHTTP.SendMsg]
    │      ├─ 根据 response_body 规则定位响应字段（通过 getRspBodyDesc）
    │      ├─ 执行响应编码器（如果配置）
    │      └─ Protobuf → JSON 编码并写入响应
    │
    ├─> [9. Metadata Conversion]
    │      └─ 将 gRPC header/trailer 转换为 HTTP 响应头
    │
    └─> [10. HTTP Response] 返回 JSON 响应
```

### 关键数据结构

#### 1. Mux（核心路由器）

`Mux` 是 Gateway 的核心结构，实现了 `Gateway` 接口和 `grpc.ClientConnInterface`：

**主要字段：**
- `localClient`: 进程内 gRPC 通道（`inprocgrpc.Channel`），用于调用本地服务
- `routerTree`: 路由树，存储和匹配 HTTP 路由规则
- `opts`: 配置选项，包括编解码器、拦截器、压缩器等

**核心职责：**
- 管理服务注册（本地服务和代理服务）
- 路由匹配和方法查找
- 处理 HTTP 请求到 gRPC 调用的转换
- 管理编解码器和压缩器
- 提供拦截器支持（Unary 和 Stream）

#### 2. RouterTree（路由树）

`RouterTree` 基于树形结构实现高效的路径匹配：

**数据结构：**
```go
type RouteTree struct {
    nodeMap map[string]*nodeTree  // 按 HTTP 方法分组的节点树
}

type nodeTree struct {
    nodeMap map[string]*nodeTree  // 子节点（路径段）
    verbMap map[string]*routeTarget  // 动词到路由目标的映射
}
```

**特性：**
- **路径解析**：使用 `participle` 解析 HTTP Rule 路径模板
- **变量提取**：支持路径变量（`{field}`）、带模式的变量（`{field=pattern}`）
- **通配符匹配**：支持 `*`（单段）和 `**`（多段贪婪匹配）
- **动词支持**：支持 `:verb` 后缀用于区分操作
- **多路由绑定**：支持 `additional_bindings` 同一方法映射到多个 HTTP 路径

**匹配算法：**
1. 按 HTTP 方法分组路由树（通过 `handlerMethod` 转换方法名）
2. 逐段匹配路径，支持精确匹配、路径变量和通配符匹配
3. 提取路径变量并映射到 Protobuf 字段路径
4. 匹配动词（如果有）确定最终路由

**匹配规则验证：**
- ✅ 路由重复注册检查：防止意外覆盖已有路由
- ✅ 路径变量边界检查：防止数组越界
- ✅ 空路径验证：确保路由路径有效
- ✅ 通配符正确处理：`*` 匹配单段，`**` 贪婪匹配多段

#### 3. Codec（编解码器）

Gateway 支持多种编解码器：

**内置编解码器：**
- `CodecJSON`：基于 `protojson` 的 JSON 编解码，支持 Protobuf 与 JSON 的双向转换
- `CodecProto`：Protobuf 二进制格式编解码
- `codecHTTPBody`：原始 HTTP Body 处理（用于 `google.api.HttpBody` 类型）

**编解码器接口：**
```go
type Codec interface {
    encoding.Codec
    MarshalAppend([]byte, any) ([]byte, error)  // 追加式序列化
}

type StreamCodec interface {
    Codec
    ReadNext(buf []byte, r io.Reader, limit int) (dst []byte, n int, err error)
    WriteNext(w io.Writer, src []byte) (n int, err error)
}
```

#### 4. Stream（流处理）

Gateway 实现了 `grpc.ServerStream` 接口，支持不同类型的流式 RPC：

**流类型：**
- `streamHTTP`：基于 Fiber Context 的 HTTP 请求/响应流
- `streamWebSocket`：WebSocket 双向流
- `streamInProcess`：进程内流（用于本地服务调用）
- `streamProxy`：代理流（转发到远程 gRPC 服务）

**关键方法：**
- `RecvMsg`: 从 HTTP 请求中接收并解码消息
- `SendMsg`: 编码并发送消息到 HTTP 响应
- `SetHeader/SendHeader`: 设置 HTTP 响应头
- `SetTrailer`: 设置 HTTP trailer

#### 5. ServiceWrapper 和 MethodWrapper（服务包装）

**ServiceWrapper：**
- 包装 gRPC 服务描述符和实现
- 区分本地服务和代理服务
- 管理服务的编解码器和拦截器配置

**MethodWrapper：**
- 包装单个 gRPC 方法
- 存储输入/输出消息类型
- 关联 HTTP Rule 配置和路径变量信息
- 支持自定义操作名称（通过 `RpcMeta` 扩展）

#### 6. Context（上下文管理）

Gateway 提供了丰富的上下文管理功能：

**元数据转换：**
- HTTP 请求头 → gRPC metadata
- gRPC header/trailer → HTTP 响应头
- 支持二进制元数据（通过 `-bin` 后缀和 base64 编码）

**保留的 HTTP 头：**
- `Content-Type`, `User-Agent`
- `grpc-*` 相关的头（如 `grpc-status`, `grpc-message` 等）

**上下文扩展：**
- `RPCMethod`: 从上下文获取 gRPC 方法名
- `HTTPPathPattern`: 获取匹配的 HTTP 路径模板
- `ServerMetadata`: 传递 gRPC 服务的 header 和 trailer

#### 7. FieldMask 支持

Gateway 支持 Google FieldMask 协议，允许客户端指定要返回的字段：

**功能：**
- `FieldMaskFromRequestBody`: 从 JSON 请求体自动生成 FieldMask
- 支持嵌套字段路径（如 `user.profile.name`）
- 自动处理动态消息类型（`google.protobuf.Struct`）
- 支持 `google.protobuf.Any` 类型的字段过滤

**使用场景：**
- 部分字段更新（PATCH 请求）
- 减少响应数据量（客户端只请求需要的字段）
- 支持 GraphQL 风格的字段选择

## 主要功能

### 1. 服务注册

Gateway 支持两种服务注册方式：

#### 注册本地服务

本地服务运行在同一个进程中，通过 `inprocgrpc.Channel` 进行进程内通信，避免了网络开销：

```go
mux := gateway.NewMux()
mux.RegisterService(&pb.UserService_ServiceDesc, &userServiceImpl{})
```

**注册流程：**
1. 验证服务实现是否满足接口要求
2. 注册到进程内 gRPC 通道（`inprocgrpc.Channel`）
3. 解析 Protobuf 服务描述符，提取方法定义
4. 解析每个方法的 HTTP Rule 注解
5. 构建路由树，注册 HTTP 路径到 gRPC 方法的映射
6. 处理 `additional_bindings`，支持多个 HTTP 路径映射到同一方法

#### 注册代理服务

代理服务将请求转发到远程 gRPC 服务，适用于微服务架构：

```go
// 创建 GrpcRouter 实现（用于提供中间件和服务描述符）
type userServiceProxy struct{}

func (p *userServiceProxy) ServiceDesc() *grpc.ServiceDesc {
    return &pb.UserService_ServiceDesc
}

func (p *userServiceProxy) Middlewares() []lava.Middleware {
    return []lava.Middleware{
        // 可以添加中间件，如认证、日志等
    }
}

// 连接到远程 gRPC 服务（可以配置负载均衡、服务发现等）
conn, _ := grpc.Dial(
    "remote-service:50051",
    grpc.WithInsecure(),
    // 可以添加更多选项，如负载均衡、超时等
)

// 注册代理服务
proxy := &userServiceProxy{}
mux.RegisterProxy(&pb.UserService_ServiceDesc, proxy, conn)
```

**代理模式特点：**
- 使用 `grpc.ClientConnInterface` 连接到远程 gRPC 服务
- 可以配置不同的连接策略（负载均衡、服务发现等）
- 保持 HTTP/JSON 接口的统一性
- 支持多实例负载均衡

**GrpcRouter 接口说明：**
`lava.GrpcRouter` 接口提供了 `Middlewares()` 和 `ServiceDesc()` 方法，用于配置服务的中间件和获取服务描述符。实际的代理路由由 `grpc.ClientConnInterface` 管理，可以通过配置 gRPC 客户端连接来实现负载均衡、服务发现等功能。

### 2. HTTP Rule 配置

在 Protobuf 定义中使用 `google.api.http` 注解：

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User) {
    option (google.api.http) = {
      get: "/v1/users/{user_id}"
      body: "*"
    };
  }
  
  rpc CreateUser(CreateUserRequest) returns (User) {
    option (google.api.http) = {
      post: "/v1/users"
      body: "user"
    };
  }
  
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
}
```

### 3. 请求/响应体映射

HTTP Rule 中的 `body` 和 `response_body` 字段控制请求和响应消息的映射方式：

#### 请求体映射（`body`）

- `"*"` 或 `""`：整个请求消息作为请求体，所有字段都从 JSON body 中解析
- `"field_name"`：仅指定字段作为请求体，其他字段从路径变量或查询参数中获取

**示例：**
```protobuf
message UpdateUserRequest {
  string user_id = 1;      // 从路径变量获取
  User user = 2;           // 从请求体获取
  string update_mask = 3;  // 从查询参数获取
}

rpc UpdateUser(UpdateUserRequest) returns (User) {
  option (google.api.http) = {
    patch: "/v1/users/{user_id}"
    body: "user"  // 只有 user 字段从请求体解析
  };
}
```

**实现细节：**
- 路径变量通过 `resolveBodyDesc` 解析字段描述符
- 查询参数通过 `gatewayutils.PopulateQueryParameters` 填充到消息字段
- 支持嵌套字段路径（如 `"user.profile"`）
- 自动处理类型转换（string → int/bool/double 等）

#### 响应体映射（`response_body`）

- `"*"` 或 `""`：返回整个响应消息
- `"field_name"`：只返回指定字段的内容

**示例：**
```protobuf
message GetUserResponse {
  User user = 1;
  Metadata metadata = 2;
}

rpc GetUser(GetUserRequest) returns (GetUserResponse) {
  option (google.api.http) = {
    get: "/v1/users/{user_id}"
    response_body: "user"  // 只返回 user 字段，metadata 被丢弃
  };
}
```

**实现细节：**
- 通过字段描述符链（`getRspBodyDesc`）定位到目标字段
- 只序列化目标字段的内容
- 支持嵌套字段路径

### 4. 中间件支持

Gateway 支持两种类型的拦截器，可以用于日志记录、认证、授权、限流、监控等场景：

#### Unary 拦截器

用于拦截 Unary（一元）RPC 调用，适用于同步的请求-响应模式：

```go
mux.SetUnaryInterceptor(func(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    // 前置处理
    start := time.Now()
    log.Info().Msgf("Calling method: %s", info.FullMethod)
    
    // 认证检查
    if err := authenticate(ctx); err != nil {
        return nil, err
    }
    
    // 调用实际处理函数
    resp, err := handler(ctx, req)
    
    // 后置处理
    duration := time.Since(start)
    if err != nil {
        log.Error().Err(err).
            Dur("duration", duration).
            Msg("Method call failed")
    } else {
        log.Info().
            Dur("duration", duration).
            Msg("Method call succeeded")
    }
    
    return resp, err
})
```

**适用场景：**
- 请求日志记录
- 认证和授权
- 请求限流
- 性能监控和指标收集
- 错误处理和转换
- 请求追踪（Trace）

#### Stream 拦截器

用于拦截 Stream（流式）RPC 调用，可以拦截双向流、客户端流和服务器流：

```go
mux.SetStreamInterceptor(func(
    srv interface{},
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error {
    // 包装 ServerStream 以拦截消息
    wrappedStream := &wrappedServerStream{
        ServerStream: ss,
    }
    
    log.Info().
        Str("method", info.FullMethod).
        Bool("client_stream", info.IsClientStream).
        Bool("server_stream", info.IsServerStream).
        Msg("Starting stream")
    
    err := handler(srv, wrappedStream)
    
    if err != nil {
        log.Error().Err(err).Msg("Stream ended with error")
    } else {
        log.Info().Msg("Stream completed successfully")
    }
    
    return err
})
```

**适用场景：**
- 流式数据的监控和统计
- 流式消息的日志记录
- 流式连接的认证和授权
- 流式数据的转换和过滤

### 5. 请求/响应拦截器

Gateway 支持针对特定消息类型的自定义编解码逻辑，可以在标准编解码前后进行自定义处理：

#### 请求解码器（Request Decoder）

请求解码器在标准 JSON → Protobuf 解码之前执行，可以：

- 从 HTTP 头、Cookie 等位置读取数据并填充到消息字段
- 修改请求消息内容
- 验证和转换请求数据
- 添加额外的上下文信息

```go
mux.SetRequestDecoder(
    protoreflect.FullName("example.UserRequest"),
    func(ctx *fiber.Ctx, msg proto.Message) error {
        // 从 HTTP 头读取用户信息
        userID := ctx.Get("X-User-ID")
        if userID != "" {
            // 将用户 ID 填充到消息字段
            msg.ProtoReflect().Set(
                msg.ProtoReflect().Descriptor().Fields().ByName("user_id"),
                protoreflect.ValueOfString(userID),
            )
        }
        
        // 从 Cookie 读取额外信息
        token := ctx.Cookies("auth_token")
        // ... 处理 token
        
        return nil
    },
)
```

#### 响应编码器（Response Encoder）

响应编码器在标准 Protobuf → JSON 编码之后执行，可以：

- 修改响应格式
- 添加额外的响应头
- 过滤敏感信息
- 包装响应数据

```go
mux.SetResponseEncoder(
    protoreflect.FullName("example.UserResponse"),
    func(ctx *fiber.Ctx, msg proto.Message) error {
        // 添加自定义响应头
        ctx.Response().Header.Set("X-Custom-Header", "value")
        
        // 可以访问和修改消息内容
        // msg.ProtoReflect()...
        
        // 如果返回 nil，将使用标准编码
        // 如果返回错误，将中断响应处理
        return nil
    },
)
```

**执行时机：**
- 请求解码器：在 `streamHTTP.RecvMsg` 中，定位请求体字段之后，JSON 解码之前执行
- 响应编码器：在 `streamHTTP.SendMsg` 中，定位响应字段之后，JSON 编码之前执行

**注意：**
- 解码器和编码器都是可选的，如果没有配置，将使用标准的 JSON 编解码
- 解码器/编码器的参数类型必须与消息类型的 FullName 完全匹配

**使用场景：**
- 从 HTTP 头/Cookie 中提取认证信息
- 数据脱敏和隐私保护
- 响应格式定制
- 添加额外的元数据

### 6. 错误处理

Gateway 提供了完整的 gRPC 错误码到 HTTP 状态码的自动映射，基于 [Google RPC Code](https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto) 规范。

#### 错误码映射表

完整的 gRPC 错误码到 HTTP 状态码映射：

| gRPC Code | HTTP Status | 说明 |
|-----------|-------------|------|
| OK | 200 OK | 成功 |
| Canceled | 499 Client Closed Request | 客户端取消请求（非标准 HTTP 状态码） |
| Unknown | 500 Internal Server Error | 未知错误 |
| InvalidArgument | 400 Bad Request | 无效的参数 |
| DeadlineExceeded | 504 Gateway Timeout | 请求超时 |
| NotFound | 404 Not Found | 资源未找到 |
| AlreadyExists | 409 Conflict | 资源已存在 |
| PermissionDenied | 403 Forbidden | 权限不足 |
| ResourceExhausted | 429 Too Many Requests | 资源耗尽（通常用于限流） |
| FailedPrecondition | 400 Bad Request | 前置条件失败（不使用 412 Precondition Failed） |
| Aborted | 409 Conflict | 操作被中止 |
| OutOfRange | 400 Bad Request | 参数超出范围 |
| Unimplemented | 501 Not Implemented | 方法未实现 |
| Internal | 500 Internal Server Error | 内部服务器错误 |
| Unavailable | 503 Service Unavailable | 服务不可用 |
| DataLoss | 500 Internal Server Error | 数据丢失 |
| Unauthenticated | 401 Unauthorized | 未认证 |

#### 错误处理机制

**自动映射：**
- Gateway 自动将 gRPC 错误转换为对应的 HTTP 状态码
- 使用 `HTTPStatusFromCode` 函数进行转换
- 未知的错误码会映射为 `500 Internal Server Error`

**在拦截器中处理错误：**
```go
mux.SetUnaryInterceptor(func(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    resp, err := handler(ctx, req)
    
    // 检查错误并可以自定义处理
    if err != nil {
        if st, ok := status.FromError(err); ok {
            // 可以记录错误、添加额外信息等
            log.Error().
                Str("code", st.Code().String()).
                Str("message", st.Message()).
                Msg("gRPC error occurred")
            
            // 可以转换错误码或添加错误详情
            // return nil, status.Errorf(st.Code(), "custom error message: %v", st.Message())
        }
    }
    
    return resp, err
})
```

**自定义错误响应：**
- 在响应编码器中可以检查错误并自定义响应格式
- 通过 Fiber Context 设置自定义的 HTTP 状态码和响应头

### 7. 元数据处理

Gateway 处理 HTTP 头和 gRPC 元数据之间的转换：

- **HTTP → gRPC**：HTTP 请求头转换为 gRPC metadata
- **gRPC → HTTP**：gRPC header 和 trailer 转换为 HTTP 响应头
- **二进制元数据**：支持 base64 编码的二进制元数据（`-bin` 后缀）

保留的 HTTP 头（不会转换为元数据）：
- `Content-Type`
- `User-Agent`
- `grpc-*` 相关的头

### 8. FieldMask 支持

Gateway 支持 Google FieldMask 协议，允许客户端指定要返回或更新的字段：

**功能：**
- `FieldMaskFromRequestBody`: 从 JSON 请求体自动生成 FieldMask
- 支持嵌套字段路径（如 `user.profile.name`）
- 自动处理动态消息类型（`google.protobuf.Struct`、`google.protobuf.Value`）
- 支持 `google.protobuf.Any` 类型的字段过滤

**使用示例：**
```go
import (
    "github.com/pubgo/lava/v2/pkg/gateway"
    "google.golang.org/protobuf/types/known/fieldmaskpb"
)

// 从请求体中提取 FieldMask
func handleUpdateUser(ctx *fiber.Ctx) error {
    req := &pb.UpdateUserRequest{}
    
    // 解析请求体
    if err := ctx.BodyParser(req); err != nil {
        return err
    }
    
    // 从请求体中提取 FieldMask（如果请求体包含字段，则生成对应的 FieldMask）
    fm, err := gateway.FieldMaskFromRequestBody(
        bytes.NewReader(ctx.Body()),
        req,
    )
    if err != nil {
        return err
    }
    
    // 使用 FieldMask 进行部分更新
    req.UpdateMask = fm
    // ... 继续处理
    return nil
}
```

**使用场景：**
- **部分字段更新（PATCH）**：客户端只传递需要更新的字段，生成对应的 FieldMask
- **减少响应数据量**：客户端通过 FieldMask 指定只返回需要的字段
- **GraphQL 风格的字段选择**：实现类似 GraphQL 的字段选择功能

## 使用示例

### 完整示例

以下是一个完整的示例，展示如何使用 Gateway：

**1. Protobuf 定义（user.proto）**

```protobuf
syntax = "proto3";

package example;

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
  
  rpc UpdateUser(UpdateUserRequest) returns (User) {
    option (google.api.http) = {
      patch: "/v1/users/{user.id}"
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

message UpdateUserRequest {
  User user = 1;
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
}
```

**2. 服务实现**

```go
package main

import (
    "context"
    "github.com/pubgo/lava/v2/pkg/gateway"
    pb "your/proto/package"
)

type userServiceImpl struct {
    pb.UnimplementedUserServiceServer
    // 可以添加依赖，如数据库客户端等
}

func (s *userServiceImpl) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // 实现逻辑
    return &pb.User{
        Id:    req.UserId,
        Name:  "John Doe",
        Email: "john@example.com",
    }, nil
}

func (s *userServiceImpl) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
    // 实现逻辑
    return req.User, nil
}

func (s *userServiceImpl) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
    // 实现逻辑
    return req.User, nil
}
```

**3. Gateway 配置和使用**

```go
package main

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/pubgo/lava/v2/pkg/gateway"
    pb "your/proto/package"
)

func main() {
    // 创建 Gateway，配置选项
    mux := gateway.NewMux(
        gateway.MaxReceiveMessageSizeOption(4 * 1024 * 1024), // 4MB
        gateway.ConnectionTimeoutOption(120 * time.Second),
    )
    
    // 可选：配置拦截器
    mux.SetUnaryInterceptor(func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // 日志、认证等处理
        return handler(ctx, req)
    })
    
    // 注册服务
    mux.RegisterService(
        &pb.UserService_ServiceDesc,
        &userServiceImpl{},
    )
    
    // 集成到 Fiber 应用
    app := fiber.New()
    app.Use("/api", func(c *fiber.Ctx) error {
        return mux.Handler(c)
    })
    
    // 启动服务器
    if err := app.Listen(":8080"); err != nil {
        panic(err)
    }
}
```

**4. 客户端调用示例**

```bash
# GET 请求
curl http://localhost:8080/api/v1/users/123

# POST 请求
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"user": {"name": "Jane Doe", "email": "jane@example.com"}}'

# PATCH 请求
curl -X PATCH http://localhost:8080/api/v1/users/123 \
  -H "Content-Type: application/json" \
  -d '{"user": {"id": "123", "name": "Jane Smith"}}'
```

### 基本使用（简化版）

### 与 gRPC 服务器集成

```go
// 在同一个进程中同时运行 gRPC 和 HTTP 服务
func main() {
    // 创建 Gateway（使用进程内通道）
    mux := gateway.NewMux()
    mux.RegisterService(&pb.Service_ServiceDesc, &serviceImpl{})
    
    // HTTP 服务器
    httpApp := fiber.New()
    httpApp.Use("/", mux.Handler)
    go httpApp.Listen(":8080")
    
    // gRPC 服务器（可选，用于直接 gRPC 调用）
    grpcServer := grpc.NewServer()
    pb.RegisterServiceServer(grpcServer, &serviceImpl{})
    go grpcServer.Serve(lis)
    
    select {}
}
```

### 代理模式

```go
// 将 HTTP 请求代理到远程 gRPC 服务
func main() {
    mux := gateway.NewMux()
    
    // 创建 GrpcRouter 实现（可选，如果需要配置中间件）
    type userServiceProxy struct{}
    
    func (p *userServiceProxy) ServiceDesc() *grpc.ServiceDesc {
        return &pb.UserService_ServiceDesc
    }
    
    func (p *userServiceProxy) Middlewares() []lava.Middleware {
        return []lava.Middleware{
            // 可以添加代理服务的中间件
        }
    }
    
    // 连接到远程 gRPC 服务（可以配置负载均衡、服务发现等）
    conn, err := grpc.Dial(
        "remote-service:50051",
        grpc.WithInsecure(),
        grpc.WithBalancerName("round_robin"), // 负载均衡
        // 更多选项...
    )
    if err != nil {
        log.Fatal().Err(err).Msg("failed to connect to gRPC service")
    }
    defer conn.Close()
    
    // 注册代理服务
    proxy := &userServiceProxy{}
    mux.RegisterProxy(
        &pb.UserService_ServiceDesc,
        proxy,
        conn,
    )
    
    app := fiber.New()
    app.Use("/api", mux.Handler)
    app.Listen(":8080")
}
```

## 配置选项

### MuxOption

```go
// 最大接收消息大小（默认 4MB）
MaxReceiveMessageSizeOption(size int)

// 最大发送消息大小（默认 math.MaxInt32）
MaxSendMessageSizeOption(size int)

// 连接超时（默认 120s）
ConnectionTimeoutOption(d time.Duration)

// 自定义类型解析器
TypesOption(resolver protoregistry.MessageTypeResolver)

// 自定义文件注册表
FilesOption(files *protoregistry.Files)

// 注册自定义编解码器
CodecOption(contentType string, codec Codec)

// 注册自定义压缩器
CompressorOption(contentEncoding string, compressor Compressor)
```

## 路径匹配规则

### 路径变量

- `{field}`：匹配单个路径段，例如 `/users/{id}` 匹配 `/users/123`
- `{field=pattern}`：匹配指定的路径模式，例如 `/users/{id=projects/*/users/*}` 匹配 `/users/projects/p1/users/u1`
- `{field.subfield}`：嵌套字段路径，用于映射到 Protobuf 消息的嵌套字段

### 通配符

- `*`：匹配单个路径段（非贪婪），例如 `/files/*` 匹配 `/files/image.jpg` 但不匹配 `/files/images/2023/photo.jpg`
- `**`：匹配多个路径段（贪婪匹配），例如 `/files/**` 匹配 `/files/images/2023/photo.jpg` 以及任意深度的路径
- 通配符与路径变量的组合：`{field=**}` 表示变量匹配多个路径段

### 动词

- `:verb`：路径末尾的动词，用于区分相同路径的不同操作
- 例如 `/users/{id}:get` 和 `/users/{id}:delete` 可以区分不同的操作
- 动词是可选的，如果没有动词，路由通过 HTTP 方法区分

### 匹配优先级

路由匹配按照以下优先级：

1. **精确匹配**：完全匹配的路径段优先
2. **路径变量**：`{field}` 形式的变量匹配
3. **单段通配符**：`*` 匹配单个路径段
4. **多段通配符**：`**` 贪婪匹配剩余所有路径段

### 示例

| 路径模板 | 匹配示例 | 变量 | 说明 |
|---------|---------|------|------|
| `/v1/users/{user_id}` | `/v1/users/123` | `user_id=123` | 简单路径变量 |
| `/v1/users/{user.id=projects/*/users/*}` | `/v1/users/projects/p1/users/u1` | `user.id=projects/p1/users/u1` | 带模式的路径变量 |
| `/v1/files/*` | `/v1/files/image.jpg` | - | 单段通配符 |
| `/v1/files/{name=**}` | `/v1/files/images/2023/photo.jpg` | `name=images/2023/photo.jpg` | 多段通配符变量 |
| `/v1/users/{user_id}/profile:get` | `/v1/users/123/profile:get` | `user_id=123`, verb=`get` | 带动词的路由 |

### 性能特性

路由匹配具有以下性能特性：

- **时间复杂度**：O(d)，其中 d 是路径深度（路径段数），与路由数量无关
- **空间复杂度**：O(n)，其中 n 是路由数量，使用前缀树结构存储
- **匹配速度**：每个路径段只需要一次 map 查找，非常高效
- **内存优化**：使用指针共享公共路径前缀，减少内存占用

## 实现细节

### 路径解析

Gateway 使用 [participle](https://github.com/alecthomas/participle) 解析 HTTP Rule 路径模板：

**解析流程：**
1. **词法分析**：将路径字符串分解为 tokens（标识符、标点符号等）
2. **语法解析**：根据 HTTP Rule 语法规则构建路径 AST
3. **路径规范化**：提取路径变量、通配符、动词等信息
4. **路由树构建**：将解析后的路径信息添加到路由树中

**支持的路径元素：**
- 字面量路径段（如 `/v1/users`）
- 路径变量（`{field}` 或 `{field.subfield}`）
- 带模式的路径变量（`{field=pattern}`，如 `{user.id=projects/*/users/*}`）
- 单段通配符（`*`）
- 多段通配符（`**`，贪婪匹配）
- 动词（`:verb`，如 `/users/{id}:get`）

### 进程内调用

Gateway 使用 [inprocgrpc](https://github.com/fullstorydev/grpchan) 实现进程内的 gRPC 调用：

**优势：**
- **零网络开销**：进程内直接调用，无需序列化/反序列化网络数据
- **类型安全**：编译时类型检查，运行时无类型转换开销
- **调试友好**：可以直接进行调试和堆栈跟踪
- **性能最优**：避免了网络延迟和协议开销

**实现方式：**
```go
localClient := new(inprocgrpc.Channel)
localClient.RegisterService(sd, ss)  // 注册服务到进程内通道
```

**使用场景：**
- 同一进程内的服务调用
- API Gateway 与后端服务在同一进程
- 减少服务间调用的延迟和开销

### 查询参数处理

Gateway 支持将 HTTP 查询参数映射到 Protobuf 消息字段：

**处理流程：**
1. 从 HTTP 请求 URL 中提取查询参数（`?key=value&key2=value2`）
2. 合并路径变量（从路径中提取的变量）
3. 根据字段名和 JSON 名称匹配 Protobuf 字段
4. 执行类型转换（string → int32/int64/bool/double/float 等）
5. 处理数组参数（重复的查询参数键）

**特性：**
- **嵌套字段支持**：使用点号分隔的字段路径（如 `user.profile.name`）
- **类型自动转换**：自动将字符串转换为目标类型
- **数组参数**：支持 `?ids=1&ids=2&ids=3` 这样的数组参数
- **默认值处理**：支持 Protobuf 字段的默认值

**实现细节：**
- 使用 `gatewayutils.PopulateQueryParameters` 函数处理参数填充
- 通过反射和 Protobuf 描述符系统访问和设置字段值
- 支持基本类型、枚举、消息类型的转换

### 元数据转换

Gateway 实现了 HTTP 头和 gRPC metadata 之间的双向转换：

**HTTP → gRPC：**
- 过滤保留的 HTTP 头（`Content-Type`、`User-Agent`、`grpc-*` 等）
- 将 HTTP 头转换为小写的 gRPC metadata key
- 处理二进制元数据（`-bin` 后缀的头，使用 base64 解码）

**gRPC → HTTP：**
- 将 gRPC header 和 trailer 转换为 HTTP 响应头
- 二进制元数据使用 base64 编码，并添加 `-bin` 后缀
- 保留的 gRPC 元数据不会转换为 HTTP 头

**特殊处理：**
- `authorization` 头：直接传递，不加前缀
- `X-Forwarded-*` 头：自动处理代理场景
- 二进制数据：使用 `base64.RawStdEncoding` 进行编解码

### 错误处理

Gateway 提供了完整的 gRPC 错误码到 HTTP 状态码的映射：

**映射规则：**
- 基于 [Google RPC Code](https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto) 规范
- 提供 `HTTPStatusFromCode` 函数进行转换
- 支持所有标准的 gRPC 错误码

**特殊处理：**
- `Canceled` → `499 Client Closed Request`（非标准但常用）
- `FailedPrecondition` → `400 Bad Request`（不使用 412 Precondition Failed）
- `Unknown` → `500 Internal Server Error`（兜底处理）

### 流式处理

Gateway 实现了 `grpc.ServerStream` 接口，支持不同类型的流式 RPC：

**流类型实现：**

1. **HTTP Stream（streamHTTP）**
   - 基于 Fiber Context 的同步请求/响应
   - 实现 `RecvMsg` 和 `SendMsg` 方法
   - 处理请求体解析和响应体编码
   - 管理 HTTP 头和 trailer

2. **WebSocket Stream（streamWebSocket）**
   - 双向通信流
   - 支持客户端和服务器双向消息传输
   - 适用于实时通信场景

3. **In-Process Stream（streamInProcess）**
   - 进程内流，用于本地服务调用
   - 直接传递消息，无需序列化
   - 性能最优的流式处理方式

4. **Proxy Stream（streamProxy）**
   - 代理流，转发到远程 gRPC 服务
   - 支持负载均衡和服务发现
   - 适用于微服务架构

**流式处理特点：**
- 统一的 `grpc.ServerStream` 接口
- 支持 header 和 trailer 的设置
- 消息边界由编解码器决定（使用 `StreamCodec.ReadNext/WriteNext`）

## 依赖关系

Gateway 模块依赖以下核心库：

- `google.golang.org/grpc`：gRPC 核心库
- `google.golang.org/protobuf`：Protobuf 运行时
- `github.com/gofiber/fiber/v2`：HTTP 框架
- `github.com/fullstorydev/grpchan/inprocgrpc`：进程内 gRPC 通道
- `github.com/alecthomas/participle/v2`：路径模板解析器

## 设计总结

### 核心优势

1. **声明式配置**：通过 Protobuf 注解定义路由，代码即文档，减少维护成本
2. **类型安全**：基于 Protobuf 的类型系统，编译时和运行时都有类型保障
3. **高性能**：进程内调用零网络开销，编解码使用高效的 Protobuf 和 protojson
4. **灵活扩展**：支持自定义编解码器、拦截器、压缩器等，满足各种定制需求
5. **标准兼容**：遵循 Google API HTTP Annotation 规范，与现有工具链兼容

### 适用场景

- **API Gateway**：将内部 gRPC 服务暴露为 RESTful API
- **微服务架构**：统一 HTTP/JSON 接口，后端使用 gRPC 通信
- **混合部署**：同时支持 HTTP 和 gRPC 客户端访问
- **前端集成**：为前端应用提供标准的 REST API
- **API 版本管理**：通过路径前缀管理不同版本的 API

### 最佳实践

1. **使用 HTTP Rule 注解**：在 Protobuf 定义中使用 `google.api.http` 注解，而不是手动注册路由
2. **合理使用 body 映射**：对于包含大量字段的消息，使用 `body: "field_name"` 只映射需要的字段
3. **利用进程内调用**：对于同一进程的服务，使用本地服务注册而不是代理模式
4. **使用拦截器**：通过 Unary/Stream 拦截器实现通用的横切关注点（日志、认证、监控等）
5. **错误处理**：利用 Gateway 的错误映射，保持 gRPC 错误码的一致性
6. **元数据管理**：使用 HTTP 头传递认证信息和元数据，Gateway 会自动转换为 gRPC metadata

### 性能考虑

- **进程内调用**：相比网络调用，进程内调用延迟更低，吞吐量更高
- **编解码开销**：JSON 编解码相比二进制 Protobuf 有额外开销，但提供了更好的可读性
- **路由匹配**：路由树使用高效的树形结构，匹配时间复杂度为 O(path_depth)
- **并发处理**：Gateway 本身是无状态的，可以安全地处理并发请求

## 参考

- [Google API HTTP Annotation](https://cloud.google.com/service-infrastructure/docs/service-management/reference/rpc/google.api#google.api.DocumentationRule.FIELDS.string.google.api.DocumentationRule.selector)
- [gRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway)
- [AIP-123: Resource-oriented design](https://google.aip.dev/123)
- [gRPC HTTP/2 Protocol](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md)

