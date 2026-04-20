# 配置参考

配置文件路径：`server/manifest/config/config.yaml`

## 服务器

```yaml
server:
  address: ":8110"      # 监听地址（Docker 用 :8080）
  serverRoot: ""        # 静态文件目录（embed 模式留空）
```

## 数据库

```yaml
database:
  default:
    type: sqlite        # sqlite 或 mysql
    path: "./data/omniwire.db"
```

## JWT 认证

```yaml
jwt:
  secret: "omniwire-secret-key-change-in-production"  # 生产必须修改
  expire: "24h"
```

## WireGuard

```yaml
wireguard:
  interface: "omniwire"
  port: 51820
  subnet: "10.66.66.0/24"
  dns: "1.1.1.1"
  auto_start: true
```

## 默认管理员

```yaml
admin:
  username: "admin"
  password: "admin123"   # 首次启动后请立即修改
```

## 端口转发

```yaml
forward:
  auto_start: true       # 启动时自动恢复已启用的规则
```
