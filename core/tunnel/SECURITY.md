# Tunnel 安全说明

本文描述 `core/tunnel` 的信任边界与生产部署建议。开发环境可放宽部分约束，**生产环境请逐项核对**。

## 信任边界

| 路径 | 鉴权 | 说明 |
| --- | --- | --- |
| Agent → Gateway **注册** | 可选 `TUNNEL_AUTH_TOKEN` | 未配置 token 时任意 agent 可注册 |
| P2P **信令** register/signal | 同 `TUNNEL_AUTH_TOKEN` | 见 `core/p2p` |
| 外部 → Gateway **HTTP 代理** | ✅ 需 token（配置 `TUNNEL_AUTH_TOKEN` 后） | `Authorization: Bearer` 或 `X-Tunnel-Token` |
| 外部 → Gateway **gRPC 代理** | ✅ 需 token | 路由行后 `Authorization: Bearer <token>\n` |
| 外部 → Gateway **Debug 代理** | ✅ 需 token | 同 HTTP |
| Admin UI (`TUNNEL_ADMIN_ADDR`) | ✅ 可选 `TUNNEL_ADMIN_TOKEN` | 未设置且非 localhost 绑定时会 Warn |
| Agent ↔ Gateway **传输** | 可选 TLS | 默认 yamux 明文 TCP |

**结论**：配置 `TUNNEL_AUTH_TOKEN` 后，Gateway 注册与代理面均受 token 保护；未配置时仍为开发模式（任意 agent 可注册、代理无鉴权）。

## 生产 Checklist

```text
[ ] 设置 TUNNEL_AUTH_TOKEN，Gateway 与 Agent 一致
[ ] Agent ↔ Gateway 启用 TLS（TUNNEL_TLS_ENABLED=true + 证书）
[ ] Debug 端口不对外（TUNNEL_DEBUG_PORT=0 或防火墙拒绝）
[ ] 设置 TUNNEL_ADMIN_TOKEN（Admin UI 非 localhost 时必设）
[ ] 启用 P2P 时使用 P2P_TURN_SECRET（HMAC 临时凭证）
[ ] 监控 gateway 429（限流）与 agent 注册失败日志
[ ] 定期轮换 TUNNEL_AUTH_TOKEN
```

## 已实现的安全机制

- **注册 token**：`TokenAuthProvider` + `subtle.ConstantTimeCompare`
- **代理鉴权**：配置 `TUNNEL_AUTH_TOKEN` 后，HTTP/Debug/gRPC 代理要求客户端 token
  - HTTP：`Authorization: Bearer <token>` 或 `X-Tunnel-Token`
  - gRPC：路由行后发送 `Authorization: Bearer <token>\n`
  - 客户端：`tunnel.GetService(..., token)`、`GRPCDialOptions{Token: ...}`
- **Admin UI**：`TUNNEL_ADMIN_TOKEN` 保护 `:6067` 管理界面
- **Deregister 绑定**：仅注册该服务的 agent 可注销
- **消息大小上限**：4 MiB（防 length-prefix OOM）
- **代理限流**：默认 100 req/s/服务名（HTTP/gRPC/debug）
- **P2P 信令限流**：默认 10/s 注册、60/s 信令（可配置）
- **心跳超时**：默认 90s 无心跳则剔除服务

## 已知缺口（后续可加固）

1. ~~`AuthProvider.Authorize()` 未在 HTTP/gRPC/debug 代理中调用~~ ✅ 已接入
2. ~~`Deregister` 无 auth~~ ✅ 已绑定 agent
3. ~~Admin UI 与 `/p2p/peers` 无鉴权~~ ✅ 代理面与 admin 已支持 token
4. ~~`TLSConfig.MinVersion` / `CipherSuites` / `ClientAuth` 字段尚未接入传输层~~ ✅ 经 `BuildClientTLSConfig` / `BuildServerTLSConfig` 接入 yamux/kcp

## 相关文档

- 使用与环境变量：`README.md`
- P2P 安全与 coturn：`docs/design-p2p.md`
- coturn 部署：`deploy/coturn/README.md`
