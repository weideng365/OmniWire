# 架构设计

## 整体架构

```
┌─────────────────────────────────────────┐
│              浏览器 (Vue 3)              │
│  Dashboard / WireGuard / Forward / Port  │
└──────────────────┬──────────────────────┘
                   │ HTTP/REST (JWT)
┌──────────────────▼──────────────────────┐
│           GoFrame HTTP Server            │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐  │
│  │ /system │ │/wireguard│ │/forward  │  │
│  └────┬────┘ └────┬─────┘ └────┬─────┘  │
│       │           │            │         │
│  ┌────▼───────────▼────────────▼──────┐  │
│  │           Service Layer             │  │
│  │  wireguard/ | forward/ | port/      │  │
│  └────────────────┬────────────────── ┘  │
│                   │                      │
│  ┌────────────────▼──────────────────┐   │
│  │         SQLite / MySQL             │   │
│  └───────────────────────────────────┘   │
└─────────────────────────────────────────┘
         │                    │
    WireGuard TUN         gnet TCP/UDP
    (wg0 interface)       (port forward)
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
       ├─ 注册路由 + JWT 中间件
       ├─ SPA fallback（非 /api 路径返回 index.html）
       └─ s.Run()
```

## 静态资源内嵌

通过构建标签控制：

- `go build -tags embed` — 将 `packed/public/` 内嵌到二进制（生产/Docker）
- 普通 `go run` — 从本地 `resource/public/` 读取（开发）

## 认证流程

1. `POST /api/v1/system/login` 返回 JWT Token
2. 前端存入 `localStorage`，后续请求携带 `Authorization: Bearer <token>`
3. GoFrame 中间件验证 Token，白名单路径（`/login`、`/health`）跳过验证
