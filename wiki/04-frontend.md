# 前端开发指南

## 目录结构

```
web/src/
├── main.js              # 入口：Vue + Pinia + Element Plus + Router
├── App.vue
├── api/index.js         # Axios 封装，自动注入 JWT
├── router/index.js      # 路由配置
├── layouts/
│   └── MainLayout.vue   # 主布局（侧边栏 + 导航）
├── stores/              # Pinia 状态
└── views/
    ├── Login.vue
    ├── Dashboard.vue
    ├── WireGuard.vue
    ├── Forward.vue
    ├── Port.vue
    └── Settings.vue
```

## 路由

| 路径 | 组件 | 说明 |
|------|------|------|
| `/login` | Login.vue | 登录页，无需认证 |
| `/dashboard` | Dashboard.vue | 仪表盘 |
| `/wireguard` | WireGuard.vue | VPN 客户端管理 |
| `/forward` | Forward.vue | 端口转发规则 |
| `/port` | Port.vue | 端口监控 |
| `/settings` | Settings.vue | 系统设置 |

## API 调用

`src/api/index.js` 封装了 Axios，自动从 `localStorage` 读取 token：

```js
import api from '@/api/index.js'

// GET
const res = await api.get('/wireguard/peers')

// POST
const res = await api.post('/wireguard/peers', { name: 'client1' })
```

401 响应自动跳转 `/login`。

## 开发命令

```bash
npm install       # 安装依赖
npm run dev       # 开发服务器 :4000
npm run build     # 构建到 ../server/resource/public
```

## 开发代理

`vite.config.js` 将 `/api` 代理到 `http://localhost:8110`，开发时无需跨域处理。
