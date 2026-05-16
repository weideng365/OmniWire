# 架构设计

## 整体架构

```
┌─────────────────────────────────────────┐
│              浏览器 (Vue 3)              │
│ Dashboard / WireGuard / OpenVPN / Forward│
└──────────────────┬──────────────────────┘
                   │ HTTP/REST (JWT)
┌──────────────────▼──────────────────────┐
│           GoFrame HTTP Server            │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐  │
│  │ /system │ │/wireguard│ │ /openvpn │  │
│  └────┬────┘ └────┬─────┘ └────┬─────┘  │
│       │           │            │         │
│  ┌────▼───────────▼────────────▼──────┐  │
│  │           Service Layer             │  │
│  │ wireguard/ | openvpn/ |             │  │
│  │ forward/   | port/                  │  │
│  └────────────────┬──────────────────── ┘  │
│                   │                      │
│  ┌────────────────▼──────────────────┐   │
│  │         SQLite / MySQL             │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────┘
         │              │              │
    WireGuard TUN    OpenVPN tun    gnet TCP/UDP
  (omniwire 默认)   (1194 默认)   (port forward)
```

## 后端分层

遵循 GoFrame 标准分层：

```
api/v1/          → 请求/响应结构体（DTO）
controller/      → HTTP 处理器，参数绑定与校验
service/         → 业务逻辑
dao/             → 数据访问
model/entity/    → 数据库实体
```

## 启动流程

```
main.go
  └─ cmd.go: Run()
       ├─ 加载静态资源（embed FS 或本地 resource/public）
       ├─ InitDatabase()     建表 + 默认数据
       ├─ InitForwardRules() 恢复已启用的转发规则
       ├─ InitWireGuard()    自动启动 WireGuard
       ├─ InitOpenVPN()      根据配置自动启动 OpenVPN
       ├─ 注册路由 + JWT 中间件（OpenVPN 回调路径限 loopback）
       ├─ SPA fallback（非 /api 路径返回 index.html）
       └─ s.Run()
```

## 静态资源内嵌

通过构建标签控制（仅使用 Go 原生 `embed`，已废弃 `gf pack`）：

- `go build -tags embed` — `internal/packed/embed.go` 生效，`//go:embed all:public` 嵌入静态资源（生产/Docker）
- 普通 `go run` / `go build` — 使用 `embed_stub.go`，运行时从本地 `resource/public/` 读取（开发）

构建脚本 `build.sh` / `build.bat` 会先清空 `internal/packed/public`，再完整复制 `resource/public` 内容，并自动加 `-tags embed`。

## 认证流程

1. `POST /api/v1/system/login` 返回 JWT Token
2. 前端存入 `localStorage`，后续请求携带 `Authorization: Bearer <token>`
3. GoFrame 中间件验证 Token，白名单路径（`/login`、`/health`）跳过验证
4. OpenVPN 回调路径（`/openvpn/auth|connect|disconnect`）不需 JWT，但校验 `RemoteAddr` 必须为 loopback，防止外网伪造身份认证
