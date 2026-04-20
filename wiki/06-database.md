# 数据库设计

默认使用 SQLite（`./data/omniwire.db`），支持 MySQL 8.0+。

## 表结构

### user — 管理员账户

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增主键 |
| username | TEXT | 用户名 |
| password | TEXT | bcrypt 哈希密码 |
| created_at | DATETIME | 创建时间 |

### wireguard_config — 服务端配置

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| interface_name | TEXT | 接口名（默认 omniwire） |
| listen_port | INTEGER | 监听端口（默认 51820） |
| private_key | TEXT | 服务端私钥 |
| public_key | TEXT | 服务端公钥 |
| subnet | TEXT | VPN 子网（默认 10.66.66.0/24） |
| dns | TEXT | 推送给客户端的 DNS |
| endpoint | TEXT | 公网地址:端口 |

### wireguard_peer — VPN 客户端

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| name | TEXT | 客户端名称 |
| public_key | TEXT | 客户端公钥 |
| private_key | TEXT | 客户端私钥 |
| allowed_ips | TEXT | 分配的 IP |
| enabled | BOOLEAN | 是否启用 |
| rx_bytes | INTEGER | 接收流量 |
| tx_bytes | INTEGER | 发送流量 |
| last_handshake | DATETIME | 最后握手时间 |

### forward_rule — 端口转发规则

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| name | TEXT | 规则名称 |
| protocol | TEXT | tcp / udp |
| listen_port | INTEGER | 本地监听端口 |
| target_addr | TEXT | 目标地址:端口 |
| enabled | BOOLEAN | 是否启用 |
| rx_bytes | INTEGER | 接收流量 |
| tx_bytes | INTEGER | 发送流量 |

### operation_log — 操作审计日志

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| user | TEXT | 操作用户 |
| action | TEXT | 操作内容 |
| created_at | DATETIME | 操作时间 |

### wireguard_connection_log — 连接日志

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| peer_id | INTEGER | 关联 peer |
| event | TEXT | 连接/断开事件 |
| created_at | DATETIME | 时间 |

## 切换到 MySQL

修改 `server/manifest/config/config.yaml`：

```yaml
database:
  default:
    type: mysql
    host: 127.0.0.1
    port: 3306
    name: omniwire
    user: root
    pass: yourpassword
```
