# OmniWire

<div align="center">

![OmniWire Logo](https://img.shields.io/badge/OmniWire-Network%20Security%20Gateway-blue?style=for-the-badge&logo=wireguard)

**WireGuard Server & Network Port Forwarding Management System Built with GoFrame**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![GoFrame](https://img.shields.io/badge/GoFrame-2.x-green?style=flat-square)](https://goframe.org/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)

English | [简体中文](README.md)

</div>

---

## Overview

**OmniWire** is a powerful network security gateway system that integrates WireGuard VPN server management, TCP/UDP port forwarding, and intelligent port management. Built with GoFrame as the backend framework and Vue 3 as the frontend framework, it provides a modern web management interface.

## Core Tech Stack

### Backend

| Technology | Version | Description |
|------------|---------|-------------|
| [Go](https://go.dev/) | 1.21+ | Programming language |
| [GoFrame](https://goframe.org/) | 2.x | Enterprise web framework with routing, ORM, config management |
| [wireguard-go](https://github.com/WireGuard/wireguard-go) | - | Pure Go WireGuard implementation, no kernel module required |
| [gnet](https://github.com/panjf2000/gnet) | 2.x | High-performance networking framework for TCP/UDP forwarding |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | 5.x | JWT token generation and validation |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | - | bcrypt password hashing |
| [go-qrcode](https://github.com/skip2/go-qrcode) | - | QR code generation |
| [Wintun](https://www.wintun.net/) | 0.14+ | Windows TUN driver |
| SQLite / MySQL | 3.x / 8.0+ | Data storage (SQLite default, zero config) |

### Frontend

| Technology | Version | Description |
|------------|---------|-------------|
| [Vue.js](https://vuejs.org/) | 3.x | Progressive JavaScript framework |
| [Vite](https://vitejs.dev/) | 5.x | Next-generation frontend build tool |
| [Element Plus](https://element-plus.org/) | 2.x | Vue 3 UI component library |
| [Pinia](https://pinia.vuejs.org/) | 2.x | Vue state management |
| [Vue Router](https://router.vuejs.org/) | 4.x | Official Vue router (with navigation guards) |
| [Axios](https://axios-http.com/) | 1.x | HTTP client (with Bearer Token interceptor) |

### Network Protocols

| Technology | Description |
|------------|-------------|
| [WireGuard](https://www.wireguard.com/) | Modern, high-performance VPN protocol |
| TCP/UDP | Transport layer protocols for port forwarding |

## Completed Features

### WireGuard VPN Server
| Feature | Description |
|---------|-------------|
| ✅ Server Configuration | Manage WireGuard interface, listen port, key pairs, MTU, DNS |
| ✅ Peer Management | Add/Edit/Delete/Enable/Disable peers |
| ✅ Config File Generation | Auto-generate client config files with one-click copy |
| ✅ QR Code Support | Generate config QR codes for mobile scanning |
| ✅ Connection Monitoring | Real-time client online status display |
| ✅ Traffic Statistics | Track upload/download traffic per client |
| ✅ One-click WireGuard Install | Auto-detect and install WireGuard kernel module |
| ✅ Service Start/Stop Control | One-click start/stop/restart WireGuard service |

### TCP/UDP Port Forwarding
| Feature | Description |
|---------|-------------|
| ✅ TCP Forwarding | TCP protocol port forwarding |
| ✅ UDP Forwarding | UDP protocol port forwarding |
| ✅ Multi-rule Management | Run multiple forwarding rules simultaneously |
| ✅ Rule Start/Stop Control | Dynamically enable/disable forwarding rules |
| ✅ Connection Limit | Limit max connections per rule |
| ✅ Bandwidth Limit | Limit upload/download rate per rule |
| ✅ Real-time Speed Monitoring | Display real-time upload/download speed per rule |
| ✅ Traffic Statistics | Cumulative upload/download traffic per rule |

### Intelligent Port Management
| Feature | Description |
|---------|-------------|
| ✅ Port Scanning | Scan local port usage |
| ✅ Port Monitoring | Monitor connection status of specified ports |
| ✅ Port Access Control | Firewall port open/close management |
| ✅ Port Usage Statistics | Port usage and traffic statistics |

### System Management
| Feature | Description |
|---------|-------------|
| ✅ JWT Authentication | JWT-based login with Bearer Token auth |
| ✅ bcrypt Password Encryption | User passwords stored with bcrypt hashing |
| ✅ Route Guards | Frontend navigation guards, auto-redirect to login |
| ✅ Password Change | Online password change (requires old password verification) |
| ✅ System Dashboard | Real-time WireGuard status, peer count, forward rules, active connections |
| ✅ RESTful API | Complete RESTful API interface |
| ✅ Docker Deployment | Multi-stage Dockerfile for containerized deployment |

## Architecture

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
│  │  JWT Middleware (whitelist: /login, /health)          │   │
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

## Quick Start

### Requirements

- **OS**: Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+ recommended), Windows requires admin privileges
- **Go**: 1.21+
- **Node.js**: 18+
- **Database**: SQLite 3 (default, zero config) / MySQL 8.0+ (optional)

### Installation

#### 1. Clone the Repository

```bash
git clone https://github.com/weideng365/OmniWire.git
cd OmniWire
```

#### 2. Start Backend Service

```bash
cd server
go mod tidy
go run main.go    # Listens on :8110 by default
```

#### 3. Start Frontend Dev Server

```bash
cd web
npm install
npm run dev       # Listens on :4000, proxies /api to :8110
```

#### 4. Access the System

Open your browser and visit `http://localhost:4000`

- Default username: `admin`
- Default password: `admin123`

> Change the default password immediately after first login via "Settings".

### Docker Deployment

```bash
docker-compose up -d
```

Exposed ports:
- `8080` — Web management interface
- `51820/udp` — WireGuard VPN

### Production Build

```bash
# Build frontend (outputs to server/resource/public)
cd web && npm run build

# Build backend binary (SQLite requires CGO_ENABLED=1)
cd server && CGO_ENABLED=1 go build -o omniwire main.go

# Run
./omniwire
```

## User Guide

### Login & Authentication

1. Accessing the system automatically redirects to the login page
2. Enter username and password to login; the system returns a JWT Token
3. Token expires after 24 hours by default; auto-redirects to login on expiry
4. All API requests (except `/login` and `/health`) require `Authorization: Bearer <token>` header

### WireGuard Management

1. Go to the "WireGuard" page; click "Start" to enable the service on first use
2. Configure server public IP (Endpoint), listen port, subnet, etc. in "Configuration"
3. Click "Add Peer" to create a peer; the system auto-assigns IP and generates key pairs
4. Import config to client devices via "Download Config" or "Scan QR Code"

### Port Forwarding

1. Go to the "Port Forwarding" page, click "Add Rule"
2. Fill in rule name, protocol (TCP/UDP), listen port, target address and port
3. Optionally set max connections and upload/download rate limits
4. Rules auto-start after creation; can be started/stopped at any time

### Port Management

1. Go to the "Port Management" page to scan local port usage
2. Check if a specific port is in use
3. Manage firewall port open/close

### System Settings

1. Go to "Settings" to change the admin password
2. Old password verification required; auto-redirects to login after successful change

## Configuration

Configuration file is located at `server/manifest/config/config.yaml`

```yaml
# Server config
server:
  address: ":8110"          # Listen address (":8080" in Docker)

# Database config
database:
  default:
    type: "sqlite"
    link: "sqlite::@file(./data/omniwire.db)"

# WireGuard config
wireguard:
  interface: "omniwire"     # Interface name
  listenPort: 51820         # Listen port
  addressRange: "10.66.66.0/24"  # VPN subnet

# Security config
security:
  jwtSecret: "omniwire-secret-key-change-in-production"  # MUST change in production!
  tokenExpire: "24h"        # Token expiration
  authEnabled: true         # Enable login authentication
```

## API Reference

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/api/v1/system/login` | User login | No |
| POST | `/api/v1/system/change-password` | Change password | Yes |
| GET | `/api/v1/system/info` | System info | Yes |
| GET | `/api/v1/system/dashboard` | Dashboard data | Yes |
| GET | `/api/v1/system/health` | Health check | No |
| GET | `/api/v1/wireguard/status` | WireGuard status | Yes |
| POST | `/api/v1/wireguard/start` | Start WireGuard | Yes |
| POST | `/api/v1/wireguard/stop` | Stop WireGuard | Yes |
| POST | `/api/v1/wireguard/restart` | Restart WireGuard | Yes |
| GET | `/api/v1/wireguard/config` | Get config | Yes |
| PUT | `/api/v1/wireguard/config` | Update config | Yes |
| GET | `/api/v1/wireguard/peers` | List peers | Yes |
| POST | `/api/v1/wireguard/peers` | Create peer | Yes |
| GET | `/api/v1/forward` | List forward rules | Yes |
| POST | `/api/v1/forward` | Create forward rule | Yes |
| PUT | `/api/v1/forward/:id` | Update forward rule | Yes |
| DELETE | `/api/v1/forward/:id` | Delete forward rule | Yes |
| POST | `/api/v1/port/scan` | Scan ports | Yes |
| GET | `/api/v1/port/check/:port` | Check port usage | Yes |
| GET | `/api/v1/port/listen` | Get listening ports | Yes |

## Important Notes

### Security

- **MUST change** `jwtSecret` in `config.yaml` for production — use a strong random key
- Change the default password `admin123` immediately after first login
- Use Nginx reverse proxy with HTTPS for the management interface in production
- JWT Token is stored in browser localStorage — ensure XSS protection

### Windows

1. **Admin privileges** — Required to create WireGuard virtual network adapters
2. **Wintun driver** — Auto-loaded on first run. If issues occur, download from [Wintun Official Site](https://www.wintun.net/)
3. **Firewall** — Manually allow UDP port 51820 (WireGuard listen port)
4. **IP config** — WireGuard adapter may show `169.254.x.x` (APIPA) address, indicating IP config failure

### Linux

1. Primary supported platform for production deployment; Docker recommended
2. Container requires `NET_ADMIN` and `SYS_MODULE` capabilities
3. Host must enable `net.ipv4.ip_forward=1`

### Database

1. Program auto-creates `./data` directory for SQLite database
2. To reset all config, delete `./data/omniwire.db` and restart
3. WireGuard key pair auto-generated on first startup and stored in database
4. User passwords stored with bcrypt hashing — no plaintext exposure on database leak

## Roadmap

### Planned Features

- [ ] **One-click SSL Certificate** — Integrate Let's Encrypt / ACME protocol for one-click SSL certificate issuance and auto-renewal via Web UI
- [ ] **Domain & Certificate Binding for Forwarding** — Bind custom domains and SSL certificates to port forwarding rules, enabling HTTPS forwarding
- [ ] **Reverse Proxy** — Built-in HTTP/HTTPS reverse proxy with load balancing, header rewriting, WebSocket proxy support
- [ ] **Domain-based Routing** — SNI/Host-based traffic routing, directing different domains to different backend services
- [ ] Operation Logging — Record user operation logs for audit trails
- [ ] Security Policies — Port access whitelist/blacklist configuration
- [ ] Multi-user Management — Multiple users with role-based permissions
- [ ] Forwarding Connection Logs — Detailed forwarding connection records

## Project Structure

```
OmniWire/
├── server/                   # Backend service (GoFrame)
│   ├── api/v1/              # API definitions (domain-based)
│   ├── internal/
│   │   ├── cmd/             # Entry point, routing, JWT middleware, DB init
│   │   ├── controller/      # Controller layer (HTTP handlers)
│   │   ├── service/         # Business logic layer
│   │   │   ├── wgserver/    # WireGuard server lifecycle management
│   │   │   ├── wireguard/   # WireGuard business logic
│   │   │   ├── forward/     # TCP/UDP port forwarding engine
│   │   │   └── port/        # Port scanning and monitoring
│   │   ├── model/           # Data models
│   │   └── dao/             # Data access layer
│   ├── manifest/config/     # Config files (config.yaml)
│   ├── resource/public/     # Frontend build output (production static files)
│   ├── go.mod
│   └── main.go              # Entry point
│
├── web/                      # Frontend app (Vue 3)
│   ├── src/
│   │   ├── api/             # API wrappers (Axios)
│   │   ├── layouts/         # Layout components
│   │   ├── router/          # Route config (with navigation guards)
│   │   ├── stores/          # Pinia state management
│   │   ├── views/           # Page views
│   │   ├── App.vue
│   │   └── main.js
│   ├── index.html
│   ├── package.json
│   └── vite.config.js       # Vite config (dev proxy /api → :8110)
│
├── Dockerfile               # Multi-stage build
├── docker-compose.yml
└── README.md
```

## Contributing

Issues and Pull Requests are welcome!

## License

This project is licensed under the [MIT License](LICENSE).

## Acknowledgments

- [WireGuard](https://www.wireguard.com/) - Modern VPN protocol
- [GoFrame](https://goframe.org/) - Go enterprise development framework
- [Vue.js](https://vuejs.org/) - Progressive JavaScript framework
- [gnet](https://github.com/panjf2000/gnet) - High-performance networking framework
- [wg-easy](https://github.com/wg-easy/wg-easy) - WireGuard Web UI reference
- [wireguard-ui](https://github.com/ngoduykhanh/wireguard-ui) - Go implementation reference

---

<div align="center">

**If this project helps you, please give it a Star!**

</div>
