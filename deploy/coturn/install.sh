#!/usr/bin/env bash
# 在 dev 服务器安装并启动 coturn（需 root）
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y coturn

cp "${SCRIPT_DIR}/turnserver.conf" /etc/turnserver.conf
sed -i 's/^#TURNSERVER_ENABLED=1/TURNSERVER_ENABLED=1/' /etc/default/coturn 2>/dev/null || true
grep -q 'TURNSERVER_ENABLED=1' /etc/default/coturn 2>/dev/null || echo 'TURNSERVER_ENABLED=1' >> /etc/default/coturn

systemctl enable coturn
systemctl restart coturn
systemctl --no-pager status coturn || true

echo "coturn listening on UDP/TCP 3478, relay 10000-20000"
