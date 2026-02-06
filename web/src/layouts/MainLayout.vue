<template>
  <div class="layout">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="logo">
        <el-icon size="32"><Connection /></el-icon>
        <span>OmniWire</span>
      </div>
      
      <el-menu
        :default-active="currentRoute"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/wireguard">
          <el-icon><Lock /></el-icon>
          <span>WireGuard</span>
        </el-menu-item>
        <el-menu-item index="/forward">
          <el-icon><Switch /></el-icon>
          <span>端口转发</span>
        </el-menu-item>
        <el-menu-item index="/port">
          <el-icon><Monitor /></el-icon>
          <span>端口管理</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
      
      <div class="sidebar-footer">
        <div class="version">v1.0.0</div>
      </div>
    </aside>
    
    <!-- 主内容区 -->
    <main class="main-content">
      <header class="header">
        <div class="page-title">
          <h1>{{ currentTitle }}</h1>
        </div>
        <div class="header-actions">
          <el-button circle @click="toggleTheme" style="margin-right: 12px">
            <el-icon><component :is="isDark ? 'Moon' : 'Sunny'" /></el-icon>
          </el-button>
          <el-dropdown>
            <el-button circle>
              <el-icon><User /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人设置</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>
      
      <div class="content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const currentRoute = computed(() => route.path)
const currentTitle = computed(() => route.meta?.title || 'OmniWire')

const isDark = ref(true)

const toggleTheme = () => {
  isDark.value = !isDark.value
  updateTheme()
}

const updateTheme = () => {
  const root = document.documentElement
  if (isDark.value) {
    root.classList.remove('light-theme')
    localStorage.setItem('theme', 'dark')
  } else {
    root.classList.add('light-theme')
    localStorage.setItem('theme', 'light')
  }
}

const handleLogout = () => {
  localStorage.removeItem('token')
  router.push('/login')
}

onMounted(() => {
  // 初始化主题（优先读取本地存储，否则随系统）
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme) {
    isDark.value = savedTheme === 'dark'
  } else {
    isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  updateTheme()
})
</script>

<style scoped>
.layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg-dark);
}

.sidebar {
  width: 240px;
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px 20px;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-color);
}

.logo .el-icon {
  color: var(--primary-color);
}

.sidebar-menu {
  flex: 1;
  padding: 16px 8px;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid var(--border-color);
  text-align: center;
}

.version {
  color: var(--text-muted);
  font-size: 12px;
}

.main-content {
  flex: 1;
  margin-left: 240px;
  display: flex;
  flex-direction: column;
}

.header {
  height: 64px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 50;
}

.page-title h1 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
