# 项目概述

## 简介

OmniWire 是一个开源的网络安全网关系统，通过 Web 界面统一管理：

- **WireGuard VPN 服务器** — 客户端创建、配置下发、流量统计
- **TCP/UDP 端口转发** — 基于 gnet 的高性能转发引擎
- **端口监控** — 扫描和检查本机端口占用状态

## 技术栈

| 层次 | 技术 |
|------|------|
| 后端框架 | Go 1.21+ / GoFrame v2 |
| 前端框架 | Vue 3 + Element Plus |
| 数据库 | SQLite（默认）/ MySQL 8.0+ |
| VPN 实现 | wireguard-go（纯 Go，无需 wg 命令行） |
| 转发引擎 | gnet v2（事件驱动，高性能） |
| 认证 | JWT（Bearer Token） |
| 构建工具 | Vite 7 |
| 容器化 | Docker + docker-compose |

## 版本信息

- 当前版本：v1.0.0
- 许可证：见 LICENSE 文件

## 快速开始

### 开发模式

```bash
# 后端
cd server
go run main.go        # 监听 :8110

# 前端（另开终端）
cd web
npm install
npm run dev           # 监听 :4000，/api 代理到 :8110
```

访问 `http://localhost:4000`，默认账号 `admin` / `admin123`。

### Docker 模式

```bash
docker-compose up -d
```

访问 `http://localhost:8080`。
