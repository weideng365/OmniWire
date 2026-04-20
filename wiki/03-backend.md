# 后端开发指南

## 目录结构

```
server/
├── main.go                          # 入口，导入 SQLite 驱动
├── go.mod
├── manifest/config/config.yaml      # 主配置
├── api/v1/                          # DTO 定义
│   ├── forward/forward.go
│   ├── port/port.go
│   ├── system/system.go
│   └── wireguard/wireguard.go
└── internal/
    ├── cmd/
    │   ├── cmd.go                   # 路由注册、服务启动
    │   └── database.go              # 建表逻辑
    ├── controller/                  # HTTP 处理器
    ├── service/                     # 业务逻辑
    ├── model/                       # 实体定义
    └── dao/                         # 数据访问
```

## 添加新接口的步骤

1. **定义 DTO** — 在 `api/v1/<domain>/` 添加请求/响应结构体
2. **实现 Controller** — 在 `internal/controller/<domain>/` 添加处理方法
3. **实现 Service** — 在 `internal/service/<domain>/` 添加业务逻辑
4. **注册路由** — 在 `internal/cmd/cmd.go` 的路由组中绑定

## 核心依赖

| 包 | 用途 |
|----|------|
| `github.com/gogf/gf/v2` | Web 框架、ORM、配置、日志 |
| `github.com/panjf2000/gnet/v2` | 高性能 TCP/UDP 转发 |
| `golang.zx2c4.com/wireguard` | WireGuard 纯 Go 实现 |
| `github.com/golang-jwt/jwt/v5` | JWT 鉴权 |
| `github.com/skip2/go-qrcode` | 客户端配置二维码 |

## 构建命令

```bash
# 开发运行
go run main.go

# 生产构建（内嵌前端，需 CGO）
CGO_ENABLED=1 go build -tags embed -o omniwire main.go

# Windows 构建
set CGO_ENABLED=1
go build -tags embed -o omniwire.exe main.go
```

## 平台注意事项

- **Linux**：主要支持平台，WireGuard 功能完整
- **Windows**：需要 `wintun.dll`，WireGuard 网卡可能回退到 APIPA 地址（169.254.x.x）
- SQLite 需要 CGO，交叉编译时需配置对应平台的 C 编译器
