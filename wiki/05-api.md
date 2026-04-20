# API 参考

所有接口基础路径：`/api/v1`

除 `/system/login` 和 `/system/health` 外，均需请求头：
```
Authorization: Bearer <JWT Token>
```

---

## 系统 `/system`

### POST /system/login
登录获取 Token。

**请求**
```json
{ "username": "admin", "password": "admin123" }
```
**响应**
```json
{ "token": "<jwt>", "user": { "id": 1, "username": "admin" } }
```

### GET /system/health
健康检查，无需认证。

### GET /system/info
系统信息（版本、运行时间等）。

### GET /system/dashboard
仪表盘统计数据（在线 peer 数、转发规则数、流量汇总）。

### POST /system/change-password
```json
{ "old_password": "admin123", "new_password": "newpass" }
```

---

## WireGuard `/wireguard`

### GET /wireguard/status
WireGuard 服务运行状态。

### POST /wireguard/start | /wireguard/stop | /wireguard/restart
控制 WireGuard 服务。

### GET /wireguard/config
获取服务端配置。

### PUT /wireguard/config
更新服务端配置（子网、端口、DNS 等）。

### GET /wireguard/peers
获取所有客户端列表（含流量统计）。

### POST /wireguard/peers
创建新客户端。
```json
{ "name": "client1", "allowed_ips": "10.66.66.2/32" }
```

---

## 端口转发 `/forward`

### GET /forward
获取所有转发规则。

### POST /forward
创建转发规则。
```json
{
  "name": "rule1",
  "protocol": "tcp",
  "listen_port": 8080,
  "target_addr": "192.168.1.100:80"
}
```

### PUT /forward/:id
更新规则。

### DELETE /forward/:id
删除规则。

---

## 端口管理 `/port`

### POST /port/scan
扫描端口范围。
```json
{ "start": 1, "end": 1024 }
```

### GET /port/check/:port
检查指定端口是否被占用。

### GET /port/listen
获取当前所有监听端口列表。
