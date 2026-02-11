# WireGuard 流量伪装部署指南 (udp2raw)

当 WireGuard 端口被 DPI（深度包检测）识别并封锁时，可以使用 udp2raw 将 UDP 流量伪装成 TCP 流量来规避封锁。

## 原理

```
原始: 客户端 → UDP 51820 (WireGuard特征) → 被封锁
伪装: 客户端 → udp2raw → TCP 443 (普通TCP特征) → 服务端 udp2raw → UDP 51820 → WireGuard
```

## 服务端部署 (Linux)

### 1. 下载 udp2raw

```bash
cd /opt
wget https://github.com/wangyu-/udp2raw/releases/download/20230206.0/udp2raw_binaries.tar.gz
tar -xzf udp2raw_binaries.tar.gz
chmod +x udp2raw_amd64
```

### 2. 启动 udp2raw

```bash
# 将 TCP 443 端口的流量转发到 WireGuard 的 UDP 端口
./udp2raw_amd64 -s -l 0.0.0.0:443 -r 127.0.0.1:51820 -k "你的密码" --raw-mode faketcp -a
```

参数说明：
- `-s`: 服务端模式
- `-l 0.0.0.0:443`: 监听 443 端口（HTTPS端口，几乎不会被封）
- `-r 127.0.0.1:51820`: 转发到本地 WireGuard 端口
- `-k "你的密码"`: 加密密码，客户端需要一致
- `--raw-mode faketcp`: 伪装成 TCP 流量
- `-a`: 自动添加 iptables 规则

### 3. 配置为系统服务

```bash
cat > /etc/systemd/system/udp2raw.service << 'EOF'
[Unit]
Description=udp2raw tunnel
After=network.target

[Service]
Type=simple
ExecStart=/opt/udp2raw_amd64 -s -l 0.0.0.0:443 -r 127.0.0.1:51820 -k "你的密码" --raw-mode faketcp -a
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable udp2raw
systemctl start udp2raw
```

### 4. 防火墙配置

```bash
# AWS 安全组需要开放 TCP 443 入站
# 服务器防火墙
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

## 客户端配置

### Linux / macOS

```bash
# 下载 udp2raw
wget https://github.com/wangyu-/udp2raw/releases/download/20230206.0/udp2raw_binaries.tar.gz
tar -xzf udp2raw_binaries.tar.gz

# 启动客户端（需要 root 权限）
sudo ./udp2raw_amd64 -c -l 127.0.0.1:51820 -r 服务器IP:443 -k "你的密码" --raw-mode faketcp -a
```

然后修改 WireGuard 配置：
```ini
[Peer]
Endpoint = 127.0.0.1:51820  # 改为本地地址
```

### Windows

Windows 需要使用 WSL2 或第三方移植版本：

**方案一：WSL2**
```bash
# 在 WSL2 中运行 udp2raw
sudo ./udp2raw_amd64 -c -l 127.0.0.1:51820 -r 服务器IP:443 -k "你的密码" --raw-mode faketcp -a
```

**方案二：udp2raw-multiplatform**

下载地址：https://github.com/nicholascw/udp2raw-multiplatform/releases

```powershell
.\udp2raw.exe -c -l 127.0.0.1:51820 -r 服务器IP:443 -k "你的密码" --raw-mode faketcp
```

## 常见问题

### Q: 为什么选择 443 端口？
443 是 HTTPS 标准端口，封锁它会影响正常网页访问，因此几乎不会被封。

### Q: 延迟会增加多少？
通常增加 10-20ms，对于 VPN 使用场景可以接受。

### Q: 服务端重启后 udp2raw 会自动启动吗？
如果配置了 systemd 服务并执行了 `systemctl enable udp2raw`，会自动启动。

### Q: 客户端断开后需要重启 udp2raw 吗？
不需要，udp2raw 会保持运行，客户端可以随时重连。

## 替代方案

如果 udp2raw 不满足需求，可以考虑：

| 工具 | 特点 |
|------|------|
| [Phantun](https://github.com/dndx/phantun) | 性能更好，Rust 编写 |
| [udp2raw-tunnel](https://github.com/wangyu-/udp2raw-tunnel) | 原版，功能最全 |
| [Gost](https://github.com/ginuerzh/gost) | 支持多种隧道协议 |
