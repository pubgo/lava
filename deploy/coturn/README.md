# coturn on dev

部署到 `~/.ssh/config` 中的 **dev** 主机（`118.178.168.253`）。

## 安装

```bash
scp -r deploy/coturn dev:/tmp/lava-coturn
ssh dev bash /tmp/lava-coturn/install.sh
```

## 默认端点

| 服务 | 地址 |
| --- | --- |
| STUN/TURN | `stun:118.178.168.253:3478` / `turn:118.178.168.253:3478` |
| Relay 端口范围 | `10000-20000` |

## 凭证（开发环境）

- 用户名：`lava`
- 密码：`lava-p2p-dev`
- Realm：`lava.dev`

与 `core/p2p/config.go` 中 `DefaultConfig()` 一致。

## 生产：HMAC 临时凭证

1. 编辑 `/etc/turnserver.conf`：注释 `user=`，启用：

```ini
use-auth-secret
static-auth-secret=YOUR_STRONG_SECRET
```

2. 客户端设置环境变量（与 secret 一致）：

```bash
export P2P_TURN_SECRET=YOUR_STRONG_SECRET
export P2P_TURN_CRED_TTL=24h   # 可选，默认 24h
```

P2P 模块会在每次 ICE 建连时生成临时 username/password（`core/p2p/turncred`），
格式为 `expiry_unix:peer_id` + `base64(HMAC-SHA1(secret, username))`。

## 防火墙

确保安全组放行 **UDP/TCP 3478** 与 **UDP 10000-20000**。

> 注意：STUN/TURN 走 **UDP**；仅 TCP 3478 通而 UDP 不通会导致 ICE 候选收集超时。

## 阿里云 ECS 配置要点

ECS 公网 IP 通常**不在网卡上**，`turnserver.conf` 必须：

```ini
external-ip=118.178.168.253/172.27.55.124   # 公网/私网 映射
relay-ip=172.27.55.124                       # 私网 IP（eth0）
```

若 `relay-ip` 误设为公网 IP，TURN Allocate 会返回 **508 Cannot create socket**。

更新配置后：

```bash
scp deploy/coturn/turnserver.conf dev:/etc/turnserver.conf
ssh dev systemctl restart coturn
```

本地验证 STUN（在 dev 上）：

```bash
ssh dev turnutils_stunclient -p 3478 127.0.0.1
```
