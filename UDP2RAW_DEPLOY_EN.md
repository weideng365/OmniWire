# WireGuard Traffic Obfuscation Guide (udp2raw)

When WireGuard ports are blocked by DPI (Deep Packet Inspection), you can use udp2raw to disguise UDP traffic as TCP traffic to bypass the blocking.

## How It Works

```
Original: Client → UDP 51820 (WireGuard signature) → Blocked
Disguised: Client → udp2raw → TCP 443 (Normal TCP) → Server udp2raw → UDP 51820 → WireGuard
```

## Server Deployment (Linux)

### 1. Download udp2raw

```bash
cd /opt
wget https://github.com/wangyu-/udp2raw/releases/download/20230206.0/udp2raw_binaries.tar.gz
tar -xzf udp2raw_binaries.tar.gz
chmod +x udp2raw_amd64
```

### 2. Start udp2raw

```bash
# Forward TCP 443 traffic to WireGuard's UDP port
./udp2raw_amd64 -s -l 0.0.0.0:443 -r 127.0.0.1:51820 -k "your_password" --raw-mode faketcp -a
```

Parameters:
- `-s`: Server mode
- `-l 0.0.0.0:443`: Listen on port 443 (HTTPS port, rarely blocked)
- `-r 127.0.0.1:51820`: Forward to local WireGuard port
- `-k "your_password"`: Encryption password, must match client
- `--raw-mode faketcp`: Disguise as TCP traffic
- `-a`: Auto-add iptables rules

### 3. Configure as System Service

```bash
cat > /etc/systemd/system/udp2raw.service << 'EOF'
[Unit]
Description=udp2raw tunnel
After=network.target

[Service]
Type=simple
ExecStart=/opt/udp2raw_amd64 -s -l 0.0.0.0:443 -r 127.0.0.1:51820 -k "your_password" --raw-mode faketcp -a
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable udp2raw
systemctl start udp2raw
```

### 4. Firewall Configuration

```bash
# AWS Security Group: Allow TCP 443 inbound
# Server firewall
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

## Client Configuration

### Linux / macOS

```bash
# Download udp2raw
wget https://github.com/wangyu-/udp2raw/releases/download/20230206.0/udp2raw_binaries.tar.gz
tar -xzf udp2raw_binaries.tar.gz

# Start client (requires root)
sudo ./udp2raw_amd64 -c -l 127.0.0.1:51820 -r SERVER_IP:443 -k "your_password" --raw-mode faketcp -a
```

Then modify WireGuard config:
```ini
[Peer]
Endpoint = 127.0.0.1:51820  # Change to local address
```

### Windows

Windows requires WSL2 or third-party ports:

**Option 1: WSL2**
```bash
# Run udp2raw in WSL2
sudo ./udp2raw_amd64 -c -l 127.0.0.1:51820 -r SERVER_IP:443 -k "your_password" --raw-mode faketcp -a
```

**Option 2: udp2raw-multiplatform**

Download: https://github.com/nicholascw/udp2raw-multiplatform/releases

```powershell
.\udp2raw.exe -c -l 127.0.0.1:51820 -r SERVER_IP:443 -k "your_password" --raw-mode faketcp
```

## FAQ

### Q: Why use port 443?
Port 443 is the standard HTTPS port. Blocking it would break normal web browsing, so it's rarely blocked.

### Q: How much latency does it add?
Typically 10-20ms, acceptable for VPN use cases.

### Q: Will udp2raw auto-start after server reboot?
Yes, if you configured the systemd service and ran `systemctl enable udp2raw`.

### Q: Do I need to restart udp2raw after client disconnects?
No, udp2raw keeps running and clients can reconnect anytime.

## Alternatives

If udp2raw doesn't meet your needs:

| Tool | Features |
|------|----------|
| [Phantun](https://github.com/dndx/phantun) | Better performance, written in Rust |
| [udp2raw-tunnel](https://github.com/wangyu-/udp2raw-tunnel) | Original version, most features |
| [Gost](https://github.com/ginuerzh/gost) | Supports multiple tunnel protocols |
