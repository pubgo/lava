# 部署与 TLS

Gateway **不内置 HTTPS/TLS**。框架内全程明文（`http` / `h2c`），TLS 与 HTTP/3（QUIC）
在边缘代理终止。这样做的原因：

- 证书签发/续期、SNI、ALPN 等是边缘代理（Traefik / Nginx / Envoy）的强项；
- 网关专注协议转换与调度，避免重复实现 TLS；
- 内网回源明文，性能更好，也便于灰度与多副本负载均衡。

## 三类协议的回源规则

Gateway 是多协议网关，三类入站协议在 L7 上性质不同，**必须分别路由**：

| 协议 | 默认端口 | 配置项 | 回源 scheme | 说明 |
| --- | --- | --- | --- | --- |
| HTTP/REST + gRPC-Web | 8080 | `grpc_server.http_port` | `http` | REST 挂在 `/api` 前缀；gRPC-Web 是普通 POST |
| WebSocket | 8081 | `grpc_server.websocket_port` | `http` | HTTP/1.1 `Upgrade`，代理需透传 Upgrade 头 |
| 原生 gRPC | 50051 | `grpc_server.grpc_port` | **`h2c`** | 明文 HTTP/2，必须用 h2c |

> 最常见的坑：原生 gRPC 回源写成 `http://` 会被代理按 HTTP/1.1 处理，导致 gRPC 失败。
> 必须用 `h2c://`（明文 HTTP/2）。

## Traefik（推荐）

仓库提供了开箱即用的配置：[`deploy/traefik/`](../../../deploy/traefik/)

| 文件 | 作用 |
| --- | --- |
| `traefik.yml` | 静态配置：80/443 entrypoints、HTTP/3、ACME 自动证书、长连接超时 |
| `dynamic.yml` | 三类协议的 routers / services（含 `h2c://` 示例） |
| `acme.json` | Let's Encrypt 证书存储（初始 `{}`，签发后由 Traefik 写入） |
| `docker-compose.yml` | Traefik + gateway 一体化示例 |
| `README.md` | 使用步骤与生产清单 |

最小步骤：

```bash
cd deploy/traefik
# 改 dynamic.yml 的 Host() 与后端 url，改 traefik.yml 的 ACME email
docker compose up -d
```

HTTP/3 在 Traefik `websecure` 启用（`http3: {}`），需暴露 UDP 443；回源规则不变，gateway 无需改动。
详见 [`deploy/traefik/README.md`](../../../deploy/traefik/README.md#http3quic)。

## X-Forwarded-* 头

边缘代理终止 TLS 后，需注入 `X-Forwarded-For` / `X-Forwarded-Proto` / `X-Forwarded-Host`，
后端据此还原客户端真实 IP 与原始 scheme。Traefik 默认即注入，无需额外配置。

## 长连接超时

gRPC streaming 与 WebSocket 是长连接，代理的空闲超时（如 Traefik `idleConnTimeout`）不要设太小，
否则空闲的 bidi / WS 会被中途掐断。`deploy/traefik/traefik.yml` 默认 `300s`。

## 生产清单

- 关闭或鉴权保护代理 dashboard。
- `acme.json` 签发后含私钥，勿提交到 git。
- WebSocket 在 `grpc_server.websocket_origin_patterns` 配置允许来源，不要依赖开发期默认的跳过校验，
  详见 [WebSocket 文档](websocket.md)。
- 证书存储文件权限 `600` 并持久化。
