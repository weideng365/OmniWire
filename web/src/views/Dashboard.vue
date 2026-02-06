<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="icon" style="background: linear-gradient(135deg, #22c55e, #16a34a);">
          <el-icon><Connection /></el-icon>
        </div>
        <div class="value">{{ stats.wireguardPeers }}</div>
        <div class="label">WireGuard 客户端</div>
      </div>
      
      <div class="stat-card">
        <div class="icon" style="background: linear-gradient(135deg, #3b82f6, #2563eb);">
          <el-icon><Switch /></el-icon>
        </div>
        <div class="value">{{ stats.forwardRules }}</div>
        <div class="label">转发规则</div>
      </div>
      
      <div class="stat-card">
        <div class="icon" style="background: linear-gradient(135deg, #f59e0b, #d97706);">
          <el-icon><Link /></el-icon>
        </div>
        <div class="value">{{ stats.activeConnections }}</div>
        <div class="label">活跃连接</div>
      </div>
      
      <div class="stat-card">
        <div class="icon" :style="{ background: stats.wireguardStatus === 'running' ? 'linear-gradient(135deg, #22c55e, #16a34a)' : 'linear-gradient(135deg, #64748b, #475569)' }">
          <el-icon><VideoPlay v-if="stats.wireguardStatus === 'running'" /><VideoPause v-else /></el-icon>
        </div>
        <div class="value">{{ stats.wireguardStatus === 'running' ? '运行中' : '已停止' }}</div>
        <div class="label">WireGuard 状态</div>
      </div>
    </div>
    
    <!-- 快捷操作 -->
    <el-card class="quick-actions">
      <template #header>
        <span>快捷操作</span>
      </template>
      <div class="action-buttons">
        <el-button type="primary" size="large" @click="$router.push('/wireguard')">
          <el-icon><Plus /></el-icon>
          添加 WireGuard 客户端
        </el-button>
        <el-button size="large" @click="$router.push('/forward')">
          <el-icon><Plus /></el-icon>
          添加转发规则
        </el-button>
        <el-button size="large" @click="$router.push('/port')">
          <el-icon><Search /></el-icon>
          扫描端口
        </el-button>
      </div>
    </el-card>
    
    <!-- 系统信息 -->
    <div class="info-grid">
      <el-card>
        <template #header>
          <span>系统信息</span>
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
            <el-tag :type="systemInfo.status === 'running' ? 'success' : 'danger'">
              {{ systemInfo.status === 'running' ? '运行中' : '异常' }}
            </el-tag>
          </div>
        </div>
      </el-card>
      
      <el-card>
        <template #header>
          <span>WireGuard 配置</span>
        </template>
        <div class="info-list">
          <div class="info-item">
            <span class="label">接口</span>
            <span class="value">{{ wgConfig.interface || 'wg0' }}</span>
          </div>
          <div class="info-item">
            <span class="label">监听端口</span>
            <span class="value">{{ wgConfig.listenPort || '51820' }}</span>
          </div>
          <div class="info-item">
            <span class="label">公钥</span>
            <span class="value key">{{ wgConfig.publicKey || '未配置' }}</span>
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
  interface: 'wg0',
  listenPort: 51820,
  publicKey: ''
})

const loadData = async () => {
  try {
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
  animation: fadeIn 0.3s ease-out;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.quick-actions {
  margin-bottom: 24px;
}

.action-buttons {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
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

.info-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-item .label {
  color: var(--text-secondary);
}

.info-item .value {
  color: var(--text-primary);
  font-weight: 500;
}

.info-item .value.key {
  font-family: monospace;
  font-size: 12px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
