# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此仓库中工作时提供指导，所有对话必须是中文回答。

## 项目概述

OmniWire 是一个网络安全网关系统，提供 WireGuard VPN 服务器管理、TCP/UDP 端口转发和端口监控功能，通过 Web 界面进行操作。

## 必须遵守的开发协议规范 (Development Protocols)
1. **先设计后编码**：在编写或修改任何代码之前，必须先用自然语言描述你的方案细节。如果原始需求存在歧义，严禁猜测，必须先提出澄清问题并获得我的确认。
2. **任务原子化**：如果一个任务涉及修改超过 3 个文件，请立即停止。将任务分解为更小、独立的子任务，并逐个提请批准后再执行。
3. **风险自评**：代码编写完成后，必须列出至少 3 个潜在的边界情况或可能出现的 Bug，并针对这些风险建议相应的测试用例（如 Unit Test 或 Integration Test）。
4. **TDD 故障排查**：发现 Bug 时，遵循“测试驱动修复”流程：
  - 首先编写一个能稳定重现该 Bug 的测试脚本/用例。
  - 运行并确认测试失败。
  - 修改代码直到该测试通过，且不破坏现有功能。
5. **动态规则进化**：如果我在对话中纠正了你的逻辑错误、代码风格或理解偏误，请在任务结束时自动更新 `CLAUDE.md`，增加一条针对性的新规则，以防止同类错误再次发生。
6. **Vue 模板安全**：在 Vue 模板中使用 `v-show` 时，务必注意其内部元素仍在求值。对于可能为 `null/undefined` 的对象属性访问（如 `user.name`），必须优先使用 `v-if` 控制渲染，或使用可选链 `?.`，防止空指针异常导致页面白屏。


## 开发命令

### 后端 (Go/GoFrame)

```bash
cd server
go mod tidy          # 安装/同步依赖
go run main.go       # 启动开发服务器，监听 :8110
go build -o omniwire main.go  # 构建二进制文件（SQLite 需要 CGO_ENABLED=1）
```

### 前端 (Vue 3/Vite)

```bash
cd web
npm install          # 安装依赖
npm run dev          # 启动开发服务器，监听 :4000（代理 /api 到 :8110）
npm run build        # 构建到 ../server/resource/public
```

### Docker

```bash
docker-compose up -d          # 构建并运行（暴露 :8080 Web 界面，:51820/udp WireGuard）
docker-compose up -d --build  # 修改后重新构建
```

### 完整开发流程

分别运行后端和前端：在 `server/` 目录执行 `go run main.go`，然后在 `web/` 目录执行 `npm run dev`。通过 `http://localhost:4000` 访问应用。

## 架构

### 后端 — GoFrame 2.x (server/)

遵循 GoFrame 分层模式：**API 定义 → Controller → Service → DAO/Model**。

- `main.go` — 入口文件，导入 SQLite 驱动
- `internal/cmd/cmd.go` — 服务启动：初始化数据库，在 `/api/v1` 下注册路由组，设置静态文件服务，启动 HTTP 服务器
- `internal/cmd/database.go` — 数据库表结构初始化（自动建表）
- `api/v1/` — 按领域划分的请求/响应结构体定义（wireguard、forward、port、system）
- `internal/controller/` — HTTP 处理器，通过 GoFrame 的 `group.Bind()` 绑定到 API 结构体。每个控制器有 `NewV1()` 构造函数
- `internal/service/` — 业务逻辑：
  - `wireguard/` — 纯 Go 实现的 WireGuard（不调用 `wg` 命令行工具）
  - `forward/` — 使用 `gnet` 实现的高性能 TCP/UDP 转发引擎
  - `port/` — 端口扫描和监控
  - `wgserver/` — WireGuard 服务器生命周期管理
- `internal/model/` — 数据模型和实体定义
- `internal/dao/` — 数据访问层
- `manifest/config/config.yaml` — 所有配置（服务器、数据库、wireguard、转发、端口、安全）

**路由**：所有 API 路由在 `/api/v1` 下分组，使用 `ghttp.MiddlewareHandlerResponse` 中间件统一 JSON 响应格式。子分组：`/system`、`/wireguard`、`/forward`、`/port`。

**数据库**：默认使用 SQLite（`./data/omniwire.db`），支持 MySQL 8.0+。通过 config.yaml 中的 `database.default` 配置。

### 前端 — Vue 3 + Element Plus (web/)

- `src/main.js` — 应用入口：挂载 Vue，配置 Pinia、Element Plus、Vue Router
- `src/router/index.js` — 路由：`/dashboard`、`/wireguard`、`/forward`、`/port`、`/settings`、`/login`
- `src/api/index.js` — Axios 客户端，基础 URL `/api/v1`，从 localStorage 读取 Bearer token 认证，401 时跳转登录页
- `src/views/` — 按路由划分的页面组件
- `src/stores/` — Pinia 状态管理
- `src/layouts/MainLayout.vue` — 需认证的布局包装器
- `vite.config.js` — 开发代理：`/api` → `http://localhost:8110`，路径别名 `@` → `./src`

**构建输出**到 `../server/resource/public`，使 Go 服务器在生产环境可以作为静态文件服务 SPA。

### Docker 部署

多阶段 Dockerfile：Node 20 构建前端 → Go 1.21 构建后端（启用 CGO）→ Alpine 3.19 运行时（包含 wireguard-tools、iptables、iproute2）。容器需要 `NET_ADMIN` 和 `SYS_MODULE` 权限，以及 `net.ipv4.ip_forward=1` 内核参数。

## 关键配置 (server/manifest/config/config.yaml)

- 服务器监听 `:8110`（开发）或 `:8080`（Docker）
- WireGuard：接口 `wg0`，端口 `51820/udp`，子网 `10.66.66.0/24`
- 默认凭据：`admin` / `admin123`
- 默认启用 JWT 认证

## 已知平台问题

- **Windows**：WireGuard 虚拟网卡无法获取静态 IP（回退到 APIPA 169.254.x.x）。UDP 端口 51820 需要手动添加防火墙规则。Wintun 驱动需要外部下载。
- **Linux**：生产部署的主要支持平台。
