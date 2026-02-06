import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'layout',
            component: () => import('@/layouts/MainLayout.vue'),
            redirect: '/dashboard',
            children: [
                {
                    path: 'dashboard',
                    name: 'dashboard',
                    component: () => import('@/views/Dashboard.vue'),
                    meta: { title: '仪表盘', icon: 'Odometer' }
                },
                {
                    path: 'wireguard',
                    name: 'wireguard',
                    component: () => import('@/views/WireGuard.vue'),
                    meta: { title: 'WireGuard', icon: 'Lock' }
                },
                {
                    path: 'forward',
                    name: 'forward',
                    component: () => import('@/views/Forward.vue'),
                    meta: { title: '端口转发', icon: 'Switch' }
                },
                {
                    path: 'port',
                    name: 'port',
                    component: () => import('@/views/Port.vue'),
                    meta: { title: '端口管理', icon: 'Monitor' }
                },
                {
                    path: 'settings',
                    name: 'settings',
                    component: () => import('@/views/Settings.vue'),
                    meta: { title: '系统设置', icon: 'Setting' }
                }
            ]
        },
        {
            path: '/login',
            name: 'login',
            component: () => import('@/views/Login.vue')
        }
    ]
})

export default router
