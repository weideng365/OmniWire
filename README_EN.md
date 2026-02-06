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

## Features

### WireGuard VPN Server
| Feature | Status | Description |
|---------|--------|-------------|
| Server Configuration | ✅ Done | Manage WireGuard interface, listen port, key pairs |
| Peer Management | ✅ Done | Add/Edit/Delete/Enable/Disable peers |
| Config File Generation | ✅ Done | Auto-generate client configuration files |
| QR Code Support | ✅ Done | Generate config QR codes for mobile scanning |
| Connection Monitoring | ✅ Done | Real-time client online status display |
| Traffic Statistics | ✅ Done | Track upload/download traffic per client |
| One-click WireGuard Install | ✅ Done | Auto-detect and install WireGuard kernel module |

### TCP/UDP Port Forwarding
| Feature | Status | Description |
|---------|--------|-------------|
| TCP Forwarding | ✅ Done | TCP protocol port forwarding |
| UDP Forwarding | ✅ Done | UDP protocol port forwarding |
| Multi-rule Management | ✅ Done | Run multiple forwarding rules simultaneously |
| Rule Start/Stop Control | ✅ Done | Dynamically enable/disable forwarding rules |
| Connection Limit | ✅ Done | Limit max connections per rule |
| Bandwidth Limit | 🔲 TODO | Limit bandwidth usage per rule |
| Forwarding Logs | 🔲 TODO | Record forwarding connection logs |

### Intelligent Port Management
| Feature | Status | Description |
|---------|--------|-------------|
| Port Scanning | ✅ Done | Scan local port usage |
| Port Monitoring | ✅ Done | Monitor connection status of specified ports |
| Port Access Control | ✅ Done | Firewall port open/close management |
| Port Usage Statistics | ✅ Done | Statistics on port usage and traffic |
| Security Policies | 🔲 TODO | Configure port access whitelist/blacklist |

### System Management
| Feature | Status | Description |
|---------|--------|-------------|
| User Authentication | ✅ Done | Login authentication and permission management |
| System Dashboard | ✅ Done | System status overview and statistics |
| Operation Logs | 🔲 TODO | Record user operation logs |
| System Configuration | ✅ Done | System parameter configuration |
| API Interface | ✅ Done | RESTful API interface |

## Tech Stack

### Backend Technologies

| Technology | Version | Description |
|------------|---------|-------------|
| [Go](https://go.dev/) | 1.21+ | Programming language |
| [GoFrame](https://goframe.org/) | 2.x | Enterprise-level web development framework |
| [gnet](https://github.com/panjf2000/gnet) | 2.x | High-performance, lightweight networking framework for TCP/UDP forwarding |
| [wireguard-go](https://github.com/WireGuard/wireguard-go) | - | Pure Go implementation of WireGuard, no kernel module required |
| [Wintun](https://www.wintun.net/) | 0.14+ | Windows TUN driver |
| [go-qrcode](https://github.com/skip2/go-qrcode) | - | QR code generation library |
| SQLite / MySQL | 3.x / 8.0+ | Data storage |

### Frontend Technologies

| Technology | Version | Description |
|------------|---------|-------------|
| [Vue.js](https://vuejs.org/) | 3.x | Progressive JavaScript framework |
| [Vite](https://vitejs.dev/) | 5.x | Next-generation frontend build tool |
| [Element Plus](https://element-plus.org/) | 2.x | Vue 3 UI component library |
| [Pinia](https://pinia.vuejs.org/) | 2.x | Vue state management library |
| [Vue Router](https://router.vuejs.org/) | 4.x | Official Vue router |
| [Axios](https://axios-http.com/) | 1.x | HTTP client |

### Network Protocols

| Technology | Description |
|------------|-------------|
| [WireGuard](https://www.wireguard.com/) | Modern, high-performance VPN protocol |
| TCP/UDP | Transport layer protocols supported for port forwarding |

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
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    GoFrame Backend                          │
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

- **OS**: Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+ recommended)
- **Go**: 1.21+
- **Node.js**: 18+
- **Database**: SQLite 3 / MySQL 8.0+

### Installation

#### 1. Clone the Repository

```bash
git clone https://github.com/weideng365/OmniWire.git
cd OmniWire
```

#### 2. Backend Service

```bash
cd server
go mod tidy
go run main.go
```

#### 3. Frontend Application

```bash
cd web
npm install
npm run dev
```

#### 4. Access the System

Open your browser and visit `http://localhost:4000`

### Docker Deployment

```bash
docker-compose up -d
```

## Configuration

Configuration file is located at `server/manifest/config/config.yaml`

```yaml
server:
  address: ":8080"
  serverRoot: "/resource/public"

database:
  default:
    type: "sqlite"
    path: "./data/omniwire.db"

wireguard:
  interface: "wg0"
  listenPort: 51820
  configPath: "/etc/wireguard"
```

## Notes

### Windows Platform

1. **WireGuard NIC IP Configuration**
   - Uses `winipcfg.LUID.SetIPAddressesForFamily` API on Windows (not PowerShell/netsh)
   - Requires **Administrator privileges** to run
   - If you see `169.254.x.x` (APIPA) address, IP configuration failed

2. **Wintun Driver**
   - Program depends on Wintun driver, auto-loaded on first run
   - If issues occur, download latest version from [Wintun Official Site](https://www.wintun.net/)

3. **Firewall**
   - Manually allow UDP port 51820 (WireGuard listen port)

### Database

1. **Data Directory**
   - Program auto-creates `./data` directory for SQLite database
   - To reset configuration, delete `./data/omniwire.db` file

2. **Key Generation**
   - WireGuard key pair auto-generated on first startup and stored in database
   - View public/private keys in Web UI under "WireGuard → Configuration"

## Roadmap

- [x] Project initialization
- [x] WireGuard server basic features
- [x] TCP/UDP port forwarding
- [x] Intelligent port management
- [x] Vue 3 frontend interface
- [x] Docker containerization
- [ ] Complete documentation
- [ ] Bandwidth limiting
- [ ] Operation logging
- [ ] Security policy configuration

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
