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

## 防火墙

确保安全组放行 **UDP/TCP 3478** 与 **UDP 10000-20000**。
