<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="icon success">
          <el-icon><Connection /></el-icon>
        </div>
        <div class="value">{{ stats.wireguardPeers }}</div>
        <div class="label">WireGuard 客户端</div>
      </div>
      
      <div class="stat-card">
        <div class="icon info">
          <el-icon><Switch /></el-icon>
        </div>
        <div class="value">{{ stats.forwardRules }}</div>
        <div class="label">转发规则</div>
      </div>
      
      <div class="stat-card">
        <div class="icon warning">
          <el-icon><Link /></el-icon>
        </div>
        <div class="value">{{ stats.activeConnections }}</div>
        <div class="label">活跃连接</div>
      </div>
      
      <div class="stat-card">
        <div class="icon" :class="stats.wireguardStatus === 'running' ? 'success' : 'offline'">
          <el-icon><VideoPlay v-if="stats.wireguardStatus === 'running'" /><VideoPause v-else /></el-icon>
        </div>
        <div class="value">{{ stats.wireguardStatus === 'running' ? '运行中' : '已停止' }}</div>
        <div class="label">WireGuard 状态</div>
      </div>
    </div>
    
    <!-- 快捷操作 -->
    <div class="section-title">快捷操作</div>
    <div class="quick-actions">
      <el-button class="action-btn" size="large" @click="$router.push('/wireguard')">
        <div class="btn-content">
          <div class="btn-icon primary"><el-icon><Plus /></el-icon></div>
          <div class="btn-text">
            <span class="main">添加客户端</span>
            <span class="sub">配置新的 WireGuard 连接</span>
          </div>
        </div>
      </el-button>
      
      <el-button class="action-btn" size="large" @click="$router.push('/forward')">
        <div class="btn-content">
          <div class="btn-icon info"><el-icon><Switch /></el-icon></div>
          <div class="btn-text">
            <span class="main">添加转发</span>
            <span class="sub">设置新的端口转发规则</span>
          </div>
        </div>
      </el-button>
      
      <el-button class="action-btn" size="large" @click="$router.push('/port')">
        <div class="btn-content">
          <div class="btn-icon warning"><el-icon><Search /></el-icon></div>
          <div class="btn-text">
            <span class="main">端口扫描</span>
            <span class="sub">检查服务器端口状态</span>
          </div>
        </div>
      </el-button>
    </div>
    
    <!-- 系统信息 -->
    <div class="info-grid">
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <el-icon><Monitor /></el-icon>
            <span>系统信息</span>
          </div>
        </template>
        <div class="info-list">
          <div class="info-item">
            <span class="label">系统名称</span>
            <span class="value">{{ systemInfo.name }}</span>
          </div>
          <div class="info-item">
            <span class="label">版本</span>
            <span class="value">{{ systemInfo.version }}</span>
          </div>
          <div class="info-item">
            <span class="label">状态</span>
            <el-tag effect="dark" :type="systemInfo.status === 'running' ? 'success' : 'danger'" size="small">
              {{ systemInfo.status === 'running' ? '运行中' : '异常' }}
            </el-tag>
          </div>
        </div>
      </el-card>
      
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <el-icon><Setting /></el-icon>
            <span>WireGuard 配置</span>
          </div>
        </template>
        <div class="info-list">
          <div class="info-item">
            <span class="label">接口</span>
            <span class="value">{{ wgConfig.interface || 'omniwire' }}</span>
          </div>
          <div class="info-item">
            <span class="label">监听端口</span>
            <span class="value">{{ wgConfig.listenPort || '51820' }}</span>
          </div>
          <div class="info-item">
            <span class="label">公钥</span>
            <el-tooltip :content="wgConfig.publicKey || '未配置'" placement="top">
              <span class="value key">{{ wgConfig.publicKey ? wgConfig.publicKey.substring(0, 20) + '...' : '未配置' }}</span>
            </el-tooltip>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { systemApi, wireguardApi } from '@/api'

const stats = ref({
  wireguardStatus: 'stopped',
  wireguardPeers: 0,
  forwardRules: 0,
  activeConnections: 0
})

const systemInfo = ref({
  name: 'OmniWire',
  version: '1.0.0',
  status: 'running'
})

const wgConfig = ref({
  interface: 'omniwire',
  listenPort: 51820,
  publicKey: ''
})

const loadData = async () => {
  try {
    // 模拟数据加载，实际应调用API
    // const [dashboardRes, wgStatusRes] = await Promise.all([
    //   systemApi.dashboard(),
    //   wireguardApi.status()
    // ])
    
    // if (dashboardRes.data) stats.value = dashboardRes.data
    // if (wgStatusRes.data) {
    //   wgConfig.value = wgStatusRes.data
    //   stats.value.wireguardStatus = wgStatusRes.data.running ? 'running' : 'stopped'
    // }
    
    // 临时保持原有逻辑
    const [dashboardRes, wgStatusRes] = await Promise.all([
       systemApi.dashboard(),
       wireguardApi.status()
    ])
    
    if (dashboardRes.data) {
      stats.value = dashboardRes.data
    }
    
    if (wgStatusRes.data) {
       wgConfig.value = wgStatusRes.data
       stats.value.wireguardStatus = wgStatusRes.data.running ? 'running' : 'stopped'
    }

  } catch (err) {
    console.error('加载数据失败:', err)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  animation: fadeIn 0.4s ease-out;
  max-width: 1200px;
  margin: 0 auto;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 32px;
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}

/* 统计卡片图标变体 */
.stat-card .icon.success {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.stat-card .icon.info {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.stat-card .icon.warning {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.stat-card .icon.offline {
  background: var(--bg-hover);
  color: var(--text-muted);
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 32px;
}

@media (max-width: 900px) {
  .quick-actions {
    grid-template-columns: 1fr;
  }
}

.action-btn {
  height: auto;
  padding: 24px;
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  justify-content: flex-start;
  transition: all 0.3s ease;
}

.action-btn:hover {
  transform: translateY(-2px);
  border-color: var(--primary-color);
  box-shadow: var(--shadow-md);
  background: var(--bg-card);
}

.btn-content {
  display: flex;
  align-items: center;
  gap: 16px;
  text-align: left;
}

.btn-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}

.btn-icon.primary { background: rgba(99, 102, 241, 0.1); color: var(--primary-color); }
.btn-icon.info { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.btn-icon.warning { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }

.btn-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.btn-text .main {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.btn-text .sub {
  font-size: 12px;
  color: var(--text-secondary);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

@media (max-width: 900px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
}

.info-card .card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-list {
  display: flex;
  flex-direction: column;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.info-item:last-child {
  border-bottom: none;
}

.info-item .label {
  color: var(--text-secondary);
  font-size: 14px;
}

.info-item .value {
  color: var(--text-primary);
  font-weight: 500;
}

.info-item .value.key {
  font-family: monospace;
  font-size: 12px;
  background: var(--bg-hover);
  padding: 4px 8px;
  border-radius: 4px;
}
</style>
