# lava gateway + Traefik

框架本身**不处理 HTTPS/TLS**，全程明文（`http` / `h2c`）。TLS 在边缘的 Traefik 终止，
回源到 lava gateway 用内网明文。本目录给出可直接套用的 Traefik 配置。

## 文件

| 文件 | 作用 |
| --- | --- |
| `traefik.yml` | 静态配置：entrypoints（80/443）、ACME 自动证书、长连接超时 |
| `dynamic.yml` | 动态配置：三类协议的 routers / services |
| `acme.json` | Let's Encrypt 证书存储（初始 `{}`，Traefik 签发后自动写入） |
| `docker-compose.yml` | Traefik + gateway 一体化示例 |

## 为什么要分三个 service

lava gateway 是多协议网关，三类入站协议在 L7 上性质不同，**必须分别路由**：

| 协议 | 默认端口 | 配置项 | 回源 scheme | 备注 |
| --- | --- | --- | --- | --- |
| HTTP/REST + gRPC-Web | 8080 | `grpc_server.http_port` | `http` | REST 挂在 `/api` 前缀；gRPC-Web 是普通 POST |
| WebSocket | 8081 | `grpc_server.websocket_port` | `http` | HTTP/1.1 `Upgrade`，Traefik 自动透传 |
| 原生 gRPC | 50051 | `grpc_server.grpc_port` | **`h2c`** | 明文 HTTP/2，必须用 h2c |

> 关键点：原生 gRPC 回源**必须**用 `h2c://`（明文 HTTP/2）。若写成 `http://`，
> Traefik 会按 HTTP/1.1 回源，gRPC 直接失败。

## 使用

1. 改 `dynamic.yml` 里的 `Host(...)` 与后端 `url`（域名、内网地址、端口）。
2. 改 `traefik.yml` 里的 ACME `email`。
3. 启动（`acme.json` 已在仓库中，Traefik 首次签发后会写入证书）：

```bash
cd deploy/traefik && docker compose up -d
```

首次部署建议确认 `acme.json` 权限为 `600`（Traefik 要求）：

```bash
chmod 600 deploy/traefik/acme.json
```

## 不用 Docker（独立 Traefik 进程）

```bash
traefik --configFile=deploy/traefik/traefik.yml
```

把 `dynamic.yml` 里的 service `url` 指向 gateway 实际监听地址即可。

## X-Forwarded-* 头

Traefik 默认会注入 `X-Forwarded-For` / `X-Forwarded-Proto` / `X-Forwarded-Host`。
gateway 后端据此还原客户端真实 IP 与原始 scheme，无需额外配置。

## 长连接超时

gRPC streaming 与 WebSocket 是长连接，`traefik.yml` 中 `idleConnTimeout` 不要设太小，
否则空闲的 bidi / WS 会被中途掐断。本目录默认 `300s`。

## 生产清单

- 关闭或鉴权保护 Traefik dashboard（`api.dashboard`）。
- `acme.json` 含真实证书后勿提交到 git（本地/服务器持久化即可）。
- WebSocket 务必在 `grpc_server.websocket_origin_patterns` 配置允许的来源，
  不要依赖开发期默认的「跳过校验」。
- `acme.json` 权限 `600`，并纳入持久化卷。
