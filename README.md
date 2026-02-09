# OmniWire

<div align="center">

![OmniWire Logo](https://img.shields.io/badge/OmniWire-Network%20Security%20Gateway-blue?style=for-the-badge&logo=wireguard)

**基于 GoFrame 开发的 WireGuard 服务端 & 网络端口转发管理系统**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![GoFrame](https://img.shields.io/badge/GoFrame-2.x-green?style=flat-square)](https://goframe.org/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)

[English](README_EN.md) | 简体中文

</div>

---

## 项目简介

**OmniWire** 是一个功能强大的网络安全网关系统，集成了 WireGuard VPN 服务端管理、TCP/UDP 端口转发、智能端口管理等核心功能。采用 GoFrame 作为后端框架，Vue 3 作为前端框架，提供现代化的 Web 管理界面。

## 核心技术栈

### 后端

| 技术 | 版本 | 说明 |
|------|------|------|
| [Go](https://go.dev/) | 1.21+ | 编程语言 |
| [GoFrame](https://goframe.org/) | 2.x | 企业级 Web 开发框架，提供路由、ORM、配置管理等 |
| [wireguard-go](https://github.com/WireGuard/wireguard-go) | - | WireGuard 纯 Go 实现，无需系统内核模块 |
| [gnet](https://github.com/panjf2000/gnet) | 2.x | 高性能、轻量级网络框架，用于 TCP/UDP 端口转发 |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | 5.x | JWT 令牌生成与验证 |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | - | bcrypt 密码哈希 |
| [go-qrcode](https://github.com/skip2/go-qrcode) | - | 二维码生成库 |
| [Wintun](https://www.wintun.net/) | 0.14+ | Windows TUN 网卡驱动 |
| SQLite / MySQL | 3.x / 8.0+ | 数据存储（默认 SQLite，零配置） |

### 前端

| 技术 | 版本 | 说明 |
|------|------|------|
| [Vue.js](https://vuejs.org/) | 3.x | 渐进式 JavaScript 框架 |
| [Vite](https://vitejs.dev/) | 5.x | 下一代前端构建工具 |
| [Element Plus](https://element-plus.org/) | 2.x | Vue 3 UI 组件库 |
| [Pinia](https://pinia.vuejs.org/) | 2.x | Vue 状态管理库 |
| [Vue Router](https://router.vuejs.org/) | 4.x | Vue 官方路由（含路由守卫鉴权） |
| [Axios](https://axios-http.com/) | 1.x | HTTP 客户端（含 Bearer Token 拦截器） |

### 网络协议

| 技术 | 说明 |
|------|------|
| [WireGuard](https://www.wireguard.com/) | 现代化、高性能 VPN 协议 |
| TCP/UDP | 端口转发支持的传输层协议 |

## 已完成功能清单

### WireGuard VPN 服务端
| 功能 | 描述 |
|------|------|
| ✅ 服务器配置管理 | 管理 WireGuard 接口、监听端口、密钥对、MTU、DNS 等 |
| ✅ 客户端（Peer）管理 | 添加/编辑/删除/启用/禁用客户端 |
| ✅ 配置文件生成 | 自动生成客户端配置文件，支持一键复制 |
| ✅ QR 码支持 | 生成配置二维码，方便移动端扫描导入 |
| ✅ 连接状态监控 | 实时显示客户端在线状态 |
| ✅ 流量统计 | 统计每个客户端的上传/下载流量 |
| ✅ 一键安装 WireGuard | 自动检测并安装 WireGuard 内核模块 |
| ✅ 服务启停控制 | 一键启动/停止/重启 WireGuard 服务 |

### TCP/UDP 端口转发
| 功能 | 描述 |
|------|------|
| ✅ TCP 转发 | 支持 TCP 协议端口转发 |
| ✅ UDP 转发 | 支持 UDP 协议端口转发 |
| ✅ 多规则管理 | 同时运行多个转发规则 |
| ✅ 规则启停控制 | 动态启用/禁用转发规则 |
| ✅ 连接数限制 | 限制单个规则的最大连接数 |
| ✅ 带宽限制 | 限制单个规则的上传/下载速率 |
| ✅ 实时速度监控 | 显示每条规则的实时上传/下载速度 |
| ✅ 流量统计 | 累计统计每条规则的总上传/下载流量 |

### 智能端口管理
| 功能 | 描述 |
|------|------|
| ✅ 端口扫描 | 扫描本机端口占用情况 |
| ✅ 端口监控 | 监控指定端口的连接状态 |
| ✅ 端口开放控制 | 防火墙端口开放/关闭管理 |
| ✅ 端口使用统计 | 统计端口的使用情况和流量 |

### 系统管理
| 功能 | 描述 |
|------|------|
| ✅ JWT 登录鉴权 | 基于 JWT 的用户登录认证，Bearer Token 鉴权 |
| ✅ bcrypt 密码加密 | 用户密码使用 bcrypt 哈希存储 |
| ✅ 路由守卫 | 前端路由守卫，未登录自动跳转登录页 |
| ✅ 密码修改 | 支持在线修改管理员密码（需验证旧密码） |
| ✅ 系统仪表盘 | 实时显示 WireGuard 状态、Peer 数、转发规则数、活跃连接数 |
| ✅ RESTful API | 完整的 RESTful API 接口 |
| ✅ Docker 部署 | 多阶段构建 Dockerfile，一键容器化部署 |

## 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                        Web Browser                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Vue 3 Frontend                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Dashboard  │  │  WireGuard  │  │  Port Forward       │  │
│  │             │  │  Management │  │  Management         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │ Axios + Bearer Token
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    GoFrame Backend                          │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  JWT Middleware (白名单: /login, /health)             │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  API Layer  │  │  Service    │  │  Data Access        │  │
│  │  Controller │  │  Layer      │  │  Layer (DAO)        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Core Services                           │    │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────────────┐  │    │
│  │  │ WireGuard │ │   TCP/UDP │ │   Port            │  │    │
│  │  │  Service  │ │  Forward  │ │   Manager         │  │    │
│  │  └───────────┘ └───────────┘ └───────────────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    System Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  WireGuard  │  │  iptables/  │  │  SQLite/MySQL       │  │
│  │  Kernel     │  │  nftables   │  │  Database           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 快速开始

### 环境要求

- **操作系统**: Linux (推荐 Ubuntu 20.04+, Debian 11+, CentOS 8+)，Windows 需管理员权限
- **Go**: 1.21+
- **Node.js**: 18+
- **数据库**: SQLite 3（默认，零配置）/ MySQL 8.0+（可选）

### 安装部署

#### 1. 克隆项目

```bash
git clone https://github.com/weideng365/OmniWire.git
cd OmniWire
```

#### 2. 启动后端服务

```bash
cd server
go mod tidy
go run main.go    # 默认监听 :8110
```

#### 3. 启动前端开发服务器

```bash
cd web
npm install
npm run dev       # 默认监听 :4000，自动代理 /api 到 :8110
```

#### 4. 访问系统

打开浏览器访问 `http://localhost:4000`

- 默认账号：`admin`
- 默认密码：`admin123`

> 首次登录后请立即在「系统设置」中修改默认密码。

### Docker 部署

```bash
docker-compose up -d
```

容器暴露端口：
- `8080` — Web 管理界面
- `51820/udp` — WireGuard VPN

### 生产构建

```bash
# 构建前端（输出到 server/resource/public）
cd web && npm run build

# 构建后端二进制（SQLite 需要 CGO_ENABLED=1）
cd server && CGO_ENABLED=1 go build -o omniwire main.go

# 运行
./omniwire
```

## 使用手册

### 登录与鉴权

1. 访问系统会自动跳转到登录页面
2. 输入用户名和密码登录，系统返回 JWT Token
3. Token 有效期默认 24 小时，过期后自动跳转登录页
4. 所有 API 请求（除 `/login` 和 `/health`）需携带 `Authorization: Bearer <token>` 头

### WireGuard 管理

1. 进入「WireGuard」页面，首次使用需点击「启动」开启服务
2. 在「配置」中设置服务器公网 IP（Endpoint）、监听端口、子网等参数
3. 点击「添加客户端」创建 Peer，系统自动分配 IP 和生成密钥对
4. 通过「下载配置」或「扫描二维码」将配置导入客户端设备

### 端口转发

1. 进入「端口转发」页面，点击「添加规则」
2. 填写规则名称、协议（TCP/UDP）、监听端口、目标地址和端口
3. 可选设置最大连接数和上传/下载速率限制
4. 创建后规则自动启动，可随时启停

### 端口管理

1. 进入「端口管理」页面，可扫描本机端口占用情况
2. 检查指定端口是否被占用
3. 管理防火墙端口开放/关闭

### 系统设置

1. 进入「系统设置」页面可修改管理员密码
2. 需输入旧密码验证身份，修改成功后自动跳转登录页重新登录

## 配置说明

配置文件位于 `server/manifest/config/config.yaml`

```yaml
# 服务器配置
server:
  address: ":8110"          # 监听地址（Docker 中为 :8080）

# 数据库配置
database:
  default:
    type: "sqlite"
    link: "sqlite::@file(./data/omniwire.db)"

# WireGuard 配置
wireguard:
  interface: "omniwire"     # 接口名称
  listenPort: 51820         # 监听端口
  addressRange: "10.66.66.0/24"  # VPN 子网

# 安全配置
security:
  jwtSecret: "omniwire-secret-key-change-in-production"  # 生产环境务必修改！
  tokenExpire: "24h"        # Token 过期时间
  authEnabled: true         # 是否启用登录验证
```

## API 接口

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | `/api/v1/system/login` | 用户登录 | 否 |
| POST | `/api/v1/system/change-password` | 修改密码 | 是 |
| GET | `/api/v1/system/info` | 系统信息 | 是 |
| GET | `/api/v1/system/dashboard` | 仪表盘数据 | 是 |
| GET | `/api/v1/system/health` | 健康检查 | 否 |
| GET | `/api/v1/wireguard/status` | WireGuard 状态 | 是 |
| POST | `/api/v1/wireguard/start` | 启动 WireGuard | 是 |
| POST | `/api/v1/wireguard/stop` | 停止 WireGuard | 是 |
| POST | `/api/v1/wireguard/restart` | 重启 WireGuard | 是 |
| GET | `/api/v1/wireguard/config` | 获取配置 | 是 |
| PUT | `/api/v1/wireguard/config` | 更新配置 | 是 |
| GET | `/api/v1/wireguard/peers` | 客户端列表 | 是 |
| POST | `/api/v1/wireguard/peers` | 创建客户端 | 是 |
| GET | `/api/v1/forward` | 转发规则列表 | 是 |
| POST | `/api/v1/forward` | 创建转发规则 | 是 |
| PUT | `/api/v1/forward/:id` | 更新转发规则 | 是 |
| DELETE | `/api/v1/forward/:id` | 删除转发规则 | 是 |
| POST | `/api/v1/port/scan` | 扫描端口 | 是 |
| GET | `/api/v1/port/check/:port` | 检查端口占用 | 是 |
| GET | `/api/v1/port/listen` | 获取监听端口 | 是 |

## 注意事项

### 安全

- **生产环境务必修改** `config.yaml` 中的 `jwtSecret`，使用随机强密钥
- 首次登录后立即修改默认密码 `admin123`
- 建议通过 Nginx 反向代理并启用 HTTPS 访问管理界面
- JWT Token 存储在浏览器 localStorage 中，注意 XSS 防护

### Windows 平台

1. **管理员权限** — 程序需要以管理员权限运行才能创建 WireGuard 虚拟网卡
2. **Wintun 驱动** — 程序依赖 Wintun 驱动，首次运行自动加载。如遇问题，从 [Wintun 官网](https://www.wintun.net/) 下载最新版本放到程序目录
3. **防火墙** — 需手动放行 UDP 51820 端口（WireGuard 监听端口）
4. **IP 配置** — Windows 下 WireGuard 虚拟网卡可能出现 `169.254.x.x`（APIPA）地址，表示 IP 配置失败

### Linux 平台

1. 生产部署的主要支持平台，推荐使用 Docker 部署
2. 容器需要 `NET_ADMIN` 和 `SYS_MODULE` 权限
3. 宿主机需开启 `net.ipv4.ip_forward=1`

### 数据库

1. 程序自动创建 `./data` 目录存放 SQLite 数据库
2. 如需重置所有配置，删除 `./data/omniwire.db` 文件后重启即可
3. 首次启动自动生成 WireGuard 密钥对并存入数据库
4. 用户密码使用 bcrypt 哈希存储，数据库泄露不会暴露明文密码

## 开发计划

### 待开发功能

- [ ] **SSL 证书一键申请** — 集成 Let's Encrypt / ACME 协议，通过 Web 界面一键申请和自动续期 SSL 证书
- [ ] **转发端口绑定域名和证书** — 为端口转发规则绑定自定义域名和 SSL 证书，支持 HTTPS 转发
- [ ] **反向代理** — 内置 HTTP/HTTPS 反向代理功能，支持负载均衡、Header 改写、WebSocket 代理
- [ ] **域名分流** — 基于域名（SNI/Host）的流量分流，不同域名路由到不同后端服务
- [ ] 操作日志记录 — 记录用户操作日志，支持审计追溯
- [ ] 安全策略配置 — 端口访问白名单/黑名单
- [ ] 多用户管理 — 支持多用户和角色权限
- [ ] 转发连接日志 — 记录转发连接详情

## 项目结构

```
OmniWire/
├── server/                   # 后端服务 (GoFrame)
│   ├── api/v1/              # API 接口定义（按领域划分）
│   ├── internal/
│   │   ├── cmd/             # 命令行入口、路由注册、JWT 中间件、数据库初始化
│   │   ├── controller/      # 控制器层（HTTP 处理器）
│   │   ├── service/         # 业务逻辑层
│   │   │   ├── wgserver/    # WireGuard 服务器生命周期管理
│   │   │   ├── wireguard/   # WireGuard 业务逻辑
│   │   │   ├── forward/     # TCP/UDP 端口转发引擎
│   │   │   └── port/        # 端口扫描和监控
│   │   ├── model/           # 数据模型
│   │   └── dao/             # 数据访问层
│   ├── manifest/config/     # 配置文件 (config.yaml)
│   ├── resource/public/     # 前端构建输出（生产环境静态文件）
│   ├── go.mod
│   └── main.go              # 程序入口
│
├── web/                      # 前端应用 (Vue 3)
│   ├── src/
│   │   ├── api/             # API 接口封装（Axios）
│   │   ├── layouts/         # 布局组件
│   │   ├── router/          # 路由配置（含路由守卫）
│   │   ├── stores/          # Pinia 状态管理
│   │   ├── views/           # 页面视图
│   │   ├── App.vue
│   │   └── main.js
│   ├── index.html
│   ├── package.json
│   └── vite.config.js       # Vite 配置（开发代理 /api → :8110）
│
├── Dockerfile               # 多阶段构建
├── docker-compose.yml
└── README.md
```

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## 开源协议

本项目采用 [MIT License](LICENSE) 开源协议。

## 致谢

- [WireGuard](https://www.wireguard.com/) - 现代化 VPN 协议
- [GoFrame](https://goframe.org/) - Go 企业级开发框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [gnet](https://github.com/panjf2000/gnet) - 高性能网络框架
- [wg-easy](https://github.com/wg-easy/wg-easy) - WireGuard Web UI 参考
- [wireguard-ui](https://github.com/ngoduykhanh/wireguard-ui) - Go 实现参考

---

<div align="center">

**如果这个项目对你有帮助，请给一个 Star！**

</div>
