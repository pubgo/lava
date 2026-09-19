# Gateway 目标设计与演进计划

本文档固化 Gateway 的目标架构与分阶段落地计划，作为后续改动的依据。  
现状实现细节见 [architecture.md](architecture.md)、[internals.md](internals.md)。

## 模块目标（不可偏离）

让 **一份 gRPC 服务实现** 同时对外提供多种协议访问，而不为每种协议再写一套 handler。

具体能力：

1. **协议归一**：HTTP/JSON、gRPC-Web、WebSocket、Native gRPC → `grpc.ServerStream` → `Dispatcher` → Backend  
2. **HTTP Rule 路由**：`google.api.http` → REST 路径/动词/body/query → 消息字段  
3. **统一调度**：unary / server-stream / client-stream / bidi 由 `Dispatcher` 一处实现  
4. **一次注册，多端复用**：`RegisterService` / `RegisterProxy` 注册一次，各前端与 zrpc 桥共用方法表  

## 设计原则

1. **Mux 只做方法表 + 调度目标**：协议细节留在 Frontend，不渗进核心。  
2. **横切逻辑挂在 Backend 边界**：本地（inproc）与 proxy 必须能套同一套「调用前/后」语义。  
3. **能力矩阵写进行为与文档**：HTTP/gRPC-Web 仅 unary / server-stream；client/bidi 走 WS / Native gRPC。  
4. **扩展点要么接线，要么不对外承诺**：Codec 与 Compressor（HTTP/gRPC-Web 帧路径）已接线。  
5. **包内只保留一条泵实现**：禁止与 `Dispatcher` 并行的第二套 bidi 转发逻辑。  

## 目标结构

```
                 ┌─ httpFrontend   (Fiber)      unary / server-stream
Client ────────►├─ wsFrontend     (net/http)   四流
                 ├─ grpcFrontend   (h2c)        四流
                 └─ zrpc bridge                 四流
                              │
                              ▼
                     Mux.DispatchFrontend
                              │
                     UseRPCMiddleware（整段 RPC）
                              │
                              ▼
                    Backend (Mux)
                     ├─ UseBackend* interceptor  ← 本地与 proxy 共用
                     ├─ local inproc
                     └─ remote proxy
```

**装配**：`gatewayserver` 一次 Register 后挂 `UseRPCMiddleware`；`NewGatewaySurface` 可从已注册 Mux 取出 HTTP / WS / gRPC 入口。  
**双栈**：接受 HTTP 用 Fiber、WS 用 net/http；不在 fasthttp 上强行跑 websocket。

**调度单一事实来源**：注册时固化 `Operation`（FullMethod / 消息类型 / 流模式 / Meta）。`routerTree` 只做 HTTP 路径 → FullMethod 索引；`methodWrapper` 仍持有 srv/codec/proxy 等实现绑定（非 schema 副本）。HTTP rule 的 body 映射细节留在路由 extras。

## 协议 × 流模式（契约）

| 前端 | Unary | Server-Stream | Client-Stream | Bidi |
| ---- | ----- | ------------- | ------------- | ---- |
| HTTP/REST + gRPC-Web | ✅ | ✅ | ❌ `Unimplemented` | ❌ |
| WebSocket | ✅ | ✅ | ✅ | ✅ |
| Native gRPC | ✅ | ✅ | ✅ | ✅ |

## 中间件模型（目标契约）

| API | 作用范围 | 用途 |
| --- | --- | --- |
| `UseRPCMiddleware` | 整段 `Mux.Dispatch`（所有流模式，本地 + proxy）；`Mux.DispatchFrontend` 在预读请求后进入它 | lava Middleware 等需观察完整 RPC 生命周期的横切 |
| `UseBackendUnaryInterceptor` / `UseBackendStreamInterceptor` | **所有** `Invoke`/`NewStream`（本地 + proxy） | 调用边界日志、鉴权、指标、超时 |
| `SetUnaryInterceptor` / `SetStreamInterceptor` | 仅 `RegisterService` 的 inproc **server** 拦截器 | 兼容层；新横切优先 `UseRPCMiddleware` / `UseBackend*` |

规则：

- 宣称「Mux 中间件」时，默认指 **RPC / Backend 链**（本地与 proxy 一致）。  
- inproc server 拦截器是实现细节/兼容层，不应当作「统一后端」的唯一挂点。  
- Unary / server-stream 的请求体经 `IncomingPayload(ctx)` 对 RPC 中间件可见。预读发生在 RPC 链之前，因此请求体解码失败不会被任何 RPC 中间件观察到。  
- RPC 链只挂在 `Dispatch`：`Invoke`/`NewStream` 在流开始即返回，包装它们会让中间件在流结束前收尾（`defer` 清理、`timeout` 的 `context.CancelFunc` 直接杀掉流）。直接以 Mux 为 client 的横切走 Backend 链。  

## 元数据契约（目标）

- 请求：白名单 header → gRPC metadata；`-bin` 按标准 base64 编解码。  
- 响应：metadata → HTTP 响应头（多值保留）；trailer 在 HTTP/JSON 与 gRPC-Web 上的暴露方式固定并测覆盖。  
- WebSocket：错误用 close reason 携带 gRPC status（已有）；成功路径 meta 后续补契约。  

## 分阶段计划

### Phase 1 — 收口语义（已完成）

- [x] HTTP 错误码映射 + gRPC-Web error trailer  
- [x] 协议 × 流模式文档化；HTTP 拒绝 client/bidi  
- [x] `WithCodec` 接线；压缩能力诚实说明  
- [x] **Backend interceptor 链**：`UseBackend*`，在 `Invoke`/`NewStream` 包装本地与 proxy  
- [x] **`TransparentHandler` 并入 `Dispatcher`**：删除并行 `forward*` 泵；远程场景 `WithPropagateBackendHeaders`  
- [x] 设计文档（本文）落地并链到 README / architecture  

### Phase 2 — 接入面整理（已完成）

- [x] `gatewayserver` 装配面：`NewGatewaySurface`；横切统一到 RPC 中间件（见 Phase 3）  
- [x] 明确 `ServeHTTP` 仅 REST/gRPC-Web  
- [x] 元数据白名单 / `-bin`：HTTP+WS 入口走 `newIncomingContext`；响应 `applyResponseMetadata` 过滤 reserved 并编码 `-bin`  
- [x] Stream lava Middleware 完整迁到 Dispatcher 级钩子（`UseRPCMiddleware`，proxy 流式与本地对齐）

### Phase 3 — 协议完整度（已完成核心项）

- [x] Dispatcher 级 stream/RPC 中间件：`Mux.UseRPCMiddleware` + `IncomingPayload`；`gatewayserver` 经 `handlerRPCMiddle` 挂载；前端统一走 `Mux.DispatchFrontend`  
- [x] gRPC-Web：成功路径保证 trailer 帧（默认 `grpc-status=0`）；`applyGRPCWebMetadata` 允许 `grpc-*`  
- [x] 注册模型：`Operation` 为调度 SSOT（注册时固化）；`routerTree` 仅作 HTTP 路径 → `FullMethod` 索引；查找走 `LookupOperation` / `findMethod`  
- [x] gRPC-Web：压缩帧协商（`grpc-encoding` / `grpc-accept-encoding` + 帧压缩标志；默认 gzip）  
- [ ] HTTP framed client-stream（若产品需要）单独设计，不假装 REST body 可表示多消息  

## 明确不做

- 为「统一」把 WebSocket 塞进 Fiber/fasthttp  
- 在 HTTP/REST 上硬撑 client/bidi  
- 新增未接线的 Option / 扩展点  
- 用删「看似无用」的辅助实现代替契约收口  

## 相关文档

- [architecture.md](architecture.md) — 现状分层与组件  
- [internals.md](internals.md) — Dispatcher、路由、错误码细节  
- [usage.md](usage.md) — 使用与拦截器作用范围  
- [deploy.md](deploy.md) — 部署与 TLS  
