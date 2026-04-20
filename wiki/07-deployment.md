# 部署指南

## Docker 部署（推荐）

### 前置要求
- Linux 主机（推荐 Ubuntu 20.04+）
- Docker 20.10+
- docker-compose 1.29+
- 内核支持 WireGuard（Linux 5.6+ 内置）

### 启动

```bash
git clone <repo>
cd OmniWire
docker-compose up -d
```

访问 `http://<host>:8080`，默认账号 `admin` / `admin123`。

### docker-compose.yml 关键配置

```yaml
services:
  omniwire:
    build: .
    network_mode: host          # 使用宿主机网络
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    sysctls:
      - net.ipv4.ip_forward=1
    volumes:
      - ./data:/app/data
      - /etc/wireguard:/etc/wireguard
      - /lib/modules:/lib/modules:ro
```

### 多阶段 Dockerfile

1. `node:20-alpine` — 构建前端
2. `golang:1.24-alpine` — 构建后端（`-tags embed` 内嵌前端）
3. `alpine:3.19` — 运行时（含 wireguard-tools、iptables、iproute2）

---

## 手动部署（Linux）

```bash
# 构建前端
cd web && npm install && npm run build

# 构建后端（内嵌前端）
cd server
CGO_ENABLED=1 go build -tags embed -o omniwire main.go

# 运行
./omniwire
```

---

## 防火墙配置

```bash
# WireGuard UDP 端口
ufw allow 51820/udp

# Web 界面
ufw allow 8080/tcp
```

---

## 生产安全建议

1. 修改默认密码（`admin123`）
2. 修改 `config.yaml` 中的 `jwt.secret`
3. 使用反向代理（Nginx）并启用 HTTPS
4. 限制 Web 界面访问 IP
