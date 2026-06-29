# gRPC Web 支持

Gateway 模块提供了完整的 gRPC Web 支持，允许浏览器直接调用 gRPC 服务，无需额外的代理服务器。

## 功能特性

- ✅ 支持 `application/grpc-web+proto` (二进制格式，推荐)
- ✅ 支持 `application/grpc-web-text+proto` (Base64 文本格式)
- ✅ 支持 `application/grpc-web+json` (JSON 格式)
- ✅ 自动处理 gRPC 帧格式
- ✅ 正确返回 Trailer (grpc-status)
- ✅ 与 Fiber 框架无缝集成
- ✅ 兼容 protobuf-ts、grpc-web、Connect-Web 等主流客户端库

## 快速开始

### 1. 服务端配置

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/cors"
    "github.com/pubgo/lava/v2/pkg/gateway"
    
    pb "your/proto/package"
)

// 实现 gRPC 服务
type greeterService struct {
    pb.UnimplementedGreeterServiceServer
}

func (s *greeterService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
    return &pb.HelloResponse{
        Message:   "Hello, " + req.GetName() + "!",
        Timestamp: time.Now().Unix(),
    }, nil
}

func main() {
    // 创建 Gateway Mux
    mux := gateway.NewMux()

    // 注册服务
    mux.RegisterService(&pb.GreeterService_ServiceDesc, &greeterService{})

    // 创建 Fiber 应用
    app := fiber.New()

    // 添加 CORS 中间件 (gRPC Web 必需)
    app.Use(cors.New(cors.Config{
        AllowOrigins:  []string{"*"},
        AllowMethods:  []string{"GET", "POST", "OPTIONS"},
        AllowHeaders:  []string{"Content-Type", "X-Grpc-Web", "X-User-Agent", "Grpc-Timeout"},
        ExposeHeaders: []string{"Grpc-Status", "Grpc-Message", "Grpc-Status-Details-Bin"},
    }))

    // 注册 HTTP/JSON 路由 (REST API)
    app.All("/v1/*", mux.Handler)

    // 注册 gRPC Web 路由
    // 路由格式: /<package>.<service>/<method>
    app.Post("/example.v1.GreeterService/*", mux.Handler)

    log.Fatal(app.Listen(":8080"))
}
```

### 2. 前端集成 (protobuf-ts)

**安装依赖：**

```bash
npm install @protobuf-ts/plugin @protobuf-ts/grpcweb-transport
```

**生成 TypeScript 代码：**

```bash
npx protoc --ts_out ./src/generated --proto_path ./proto ./proto/greeter.proto
```

**使用客户端：**

```typescript
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport';
import { GreeterServiceClient } from './generated/greeter.client';

// 创建 Transport
const transport = new GrpcWebFetchTransport({
    baseUrl: 'http://localhost:8080',
    format: 'binary', // 推荐使用二进制格式
});

// 创建客户端
const client = new GreeterServiceClient(transport);

// 调用方法
async function sayHello(name: string) {
    try {
        const call = client.sayHello({ name });
        const response = await call.response;
        
        console.log('Message:', response.message);
        console.log('Timestamp:', response.timestamp);
        
        // 获取状态
        const status = await call.status;
        console.log('Status:', status.code); // "OK"
    } catch (error) {
        console.error('gRPC Error:', error);
    }
}

sayHello('World');
```

## Content-Type 说明

Gateway 通过 `Content-Type` 头判断请求类型：

| Content-Type                      | 描述                       |
| --------------------------------- | -------------------------- |
| `application/grpc-web+proto`      | gRPC Web 二进制格式 (推荐) |
| `application/grpc-web-text+proto` | gRPC Web Base64 文本格式   |
| `application/grpc-web+json`       | gRPC Web JSON 格式         |
| `application/json`                | 普通 HTTP/JSON (REST API)  |

## gRPC Web 协议格式

### 请求格式

```
[1 byte: compression flag] [4 bytes: message length (big-endian)] [message bytes]
```

- Compression flag: `0x00` = 无压缩
- Message length: 4 字节大端整数
- Message: Protobuf 编码的消息

### 响应格式

响应由数据帧和 Trailer 帧组成：

```
数据帧:   [0x00] [4 bytes: length] [protobuf message]
Trailer:  [0x80] [4 bytes: length] [HTTP headers format]
```

Trailer 示例: `Grpc-Status: 0\r\n`

## CORS 配置

gRPC Web 是跨域请求，必须配置 CORS：

```go
app.Use(cors.New(cors.Config{
    AllowOrigins:  []string{"*"},  // 生产环境应该限制具体域名
    AllowMethods:  []string{"GET", "POST", "OPTIONS"},
    AllowHeaders:  []string{"Content-Type", "X-Grpc-Web", "X-User-Agent", "Grpc-Timeout", "Authorization"},
    ExposeHeaders: []string{"Grpc-Status", "Grpc-Message", "Grpc-Status-Details-Bin"},
}))
```

## 与其他客户端库兼容

### grpc-web (Google 官方)

```typescript
import { GreeterServiceClient } from './generated/greeter_grpc_web_pb';

const client = new GreeterServiceClient('http://localhost:8080');

client.sayHello({ name: 'World' }, {}, (err, response) => {
    if (err) {
        console.error(err);
        return;
    }
    console.log(response.getMessage());
});
```

### Connect-Web

```typescript
import { createPromiseClient } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";
import { GreeterService } from "./generated/greeter_connect";

const transport = createGrpcWebTransport({
    baseUrl: "http://localhost:8080",
});

const client = createPromiseClient(GreeterService, transport);

const response = await client.sayHello({ name: "World" });
console.log(response.message);
```

## 调试方法

### 使用 curl 测试 HTTP/JSON

```bash
curl -X POST http://localhost:8080/v1/greeter/hello \
  -H "Content-Type: application/json" \
  -d '{"name": "World"}'
```

### 使用 curl 测试 gRPC Web

```bash
# 发送请求并查看原始响应
printf '\x00\x00\x00\x00\x07\n\x05World' | \
curl -X POST http://localhost:8080/example.v1.GreeterService/SayHello \
  -H "Content-Type: application/grpc-web+proto" \
  --data-binary @- \
  --output - | xxd

# 查看响应头
curl -X POST http://localhost:8080/example.v1.GreeterService/SayHello \
  -H "Content-Type: application/grpc-web+proto" \
  --data-binary @- -i
```

## 完整示例

参见 `internal/examples/grpcweb/` 目录：

```
internal/examples/grpcweb/
├── proto/
│   └── greeter.proto       # Proto 定义
├── main.go                 # Go 服务端实现
├── static/                 # 编译后的前端静态文件
└── frontend/               # TypeScript 前端源码
    ├── package.json
    ├── tsconfig.json
    ├── vite.config.ts
    ├── index.html
    └── src/
        ├── main.ts         # 客户端代码
        └── generated/      # 生成的 TypeScript 代码
```

**运行示例：**

```bash
# 启动服务端
go run ./internal/examples/grpcweb/

# 访问测试页面
open http://localhost:8080/
```

## 流式支持

gRPC-Web 前端复用统一的 `Dispatcher`：

- ✅ Unary
- ✅ 服务端响应流（Server Streaming）
- ⚠️ 客户端流 / 双向流（Client / Bidi Streaming）：受 gRPC-Web 协议本身限制，浏览器侧难以原生支持上行流。若需要完整双向流能力，请改用 [WebSocket 前端](websocket.md)（基于同一套后端 handler）。

> 注意：`Dispatcher` 在调度层已实现四种流模式，gRPC-Web 的客户端流/双向流限制来自浏览器与 gRPC-Web 协议，而非 Gateway 调度能力。
