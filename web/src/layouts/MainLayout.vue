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
          <el-tooltip content="切换主题" placement="bottom">
            <el-button circle @click="toggleTheme">
              <el-icon><component :is="isDark ? 'Moon' : 'Sunny'" /></el-icon>
            </el-button>
          </el-tooltip>
          
          <el-tooltip content="GitHub 源码" placement="bottom">
            <el-button circle tag="a" href="https://github.com/weideng365/OmniWire" target="_blank" class="github-link">
              <el-icon><Link /></el-icon>
            </el-button>
          </el-tooltip>

          <el-tooltip content="退出登录" placement="bottom">
            <el-button circle type="danger" plain @click="handleLogout">
              <el-icon><SwitchButton /></el-icon>
            </el-button>
          </el-tooltip>
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
import { SwitchButton, Link, Moon, Sunny, Odometer, Lock, Switch, Monitor, Setting, Connection, User } from '@element-plus/icons-vue'

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

const openGithub = () => {
  window.open('https://github.com/weideng365/OmniWire', '_blank')
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
  width: 260px;
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
  transition: all 0.3s ease;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 72px;
  padding: 0 24px;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-color);
  background: rgba(var(--primary-rgb), 0.05); /* Very subtle primary tint */
}

.logo :deep(.el-icon) {
  /* color: var(--primary-color);
  background: var(--bg-card); */
  padding: 4px;
  border-radius: 8px;
  /* box-shadow: 0 4px 10px rgba(var(--primary-rgb), 0.2); */ 
  /* Resetting styles to prevent conflict */
  color: inherit;
  background: transparent;
  box-shadow: none;
  font-size: 28px;
}
/* Re-apply specifically */
.logo :deep(.el-icon) {
    color: var(--primary-color);
}

.sidebar-menu {
  flex: 1;
  padding: 24px 12px;
  overflow-y: auto;
  border-right: none; /* Remove element-plus default border */
}

/* Override Element Plus Menu Item Styles */
:deep(.el-menu-item) {
  margin-bottom: 4px;
}

.sidebar-footer {
  padding: 20px;
  border-top: 1px solid var(--border-color);
  text-align: center;
  background: var(--bg-hover);
}

.version {
  color: var(--text-muted);
  font-size: 12px;
  font-family: monospace;
  background: var(--bg-card);
  padding: 2px 8px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  display: inline-block;
}

.main-content {
  flex: 1;
  margin-left: 260px;
  display: flex;
  flex-direction: column;
  min-width: 0; /* Prevent flex overflow */
  transition: margin-left 0.3s ease;
}

.header {
  height: 72px;
  /* Using glass effect variable or fallback */
  background: var(--bg-header-glass, rgba(30, 41, 59, 0.8));
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32px;
  position: sticky;
  top: 0;
  z-index: 50;
}

.page-title h1 {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.5px;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.content {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
  max-width: 1600px; /* Limit max width for ultra-wide screens */
  margin: 0 auto;
  width: 100%;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
