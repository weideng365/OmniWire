import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建 axios 实例
const request = axios.create({
    baseURL: '/api/v1',
    timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
    config => {
        const token = localStorage.getItem('token')
        if (token) {
            config.headers['Authorization'] = `Bearer ${token}`
        }
        return config
    },
    error => {
        return Promise.reject(error)
    }
)

// 响应拦截器
request.interceptors.response.use(
    response => {
        return response.data
    },
    error => {
        const message = error.response?.data?.message || error.message || '请求失败'
        ElMessage.error(message)

        if (error.response?.status === 401) {
            localStorage.removeItem('token')
            window.location.href = '/login'
        }

        return Promise.reject(error)
    }
)

// 系统 API
export const systemApi = {
    info: () => request.get('/system/info'),
    dashboard: () => request.get('/system/dashboard'),
    health: () => request.get('/system/health')
}

// WireGuard API
export const wireguardApi = {
    status: () => request.get('/wireguard/status'),
    start: () => request.post('/wireguard/start'),
    stop: () => request.post('/wireguard/stop'),
    restart: () => request.post('/wireguard/restart'),
    config: () => request.get('/wireguard/config'),
    updateConfig: (data) => request.put('/wireguard/config', data),
    peers: () => request.get('/wireguard/peers'),
    createPeer: (data) => request.post('/wireguard/peers', data),
    updatePeer: (id, data) => request.put(`/wireguard/peers/${id}`, data),
    deletePeer: (id) => request.delete(`/wireguard/peers/${id}`),
    peerConfig: (id) => request.get(`/wireguard/peers/${id}/config`),
    peerQRCode: (id) => request.get(`/wireguard/peers/${id}/qrcode`)
}

// 端口转发 API
export const forwardApi = {
    list: (params) => request.get('/forward', { params }),
    create: (data) => request.post('/forward', data),
    update: (id, data) => request.put(`/forward/${id}`, data),
    delete: (id) => request.delete(`/forward/${id}`),
    start: (id) => request.post(`/forward/${id}/start`),
    stop: (id) => request.post(`/forward/${id}/stop`),
    stats: (id) => request.get(`/forward/${id}/stats`)
}

// 端口管理 API
export const portApi = {
    scan: (data) => request.post('/port/scan', data),
    check: (port) => request.get(`/port/check/${port}`),
    listen: () => request.get('/port/listen'),
    connections: (port) => request.get(`/port/connections/${port}`),
    firewall: () => request.get('/port/firewall'),
    firewallOpen: (data) => request.post('/port/firewall/open', data),
    firewallClose: (data) => request.post('/port/firewall/close', data)
}

export default request
