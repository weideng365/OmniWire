<template>
  <div class="dashboard">
    <!-- 汇总统计 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="icon primary">
          <el-icon><User /></el-icon>
        </div>
        <div class="value">{{ stats.wireguardPeers + (ovpnStatus.clientCount || 0) }}</div>
        <div class="label">VPN 总客户端</div>
        <div class="sub-label">WG {{ stats.wireguardPeers }} / OVPN {{ ovpnStatus.clientCount || 0 }}</div>
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
    </div>

    <!-- VPN 分组 -->
    <div class="vpn-grid">
      <!-- WireGuard -->
      <el-card class="vpn-card">
        <template #header>
          <div class="card-header">
            <div class="card-title">
              <el-icon><Connection /></el-icon>
              <span>WireGuard</span>
            </div>
            <el-tag :type="stats.wireguardStatus === 'running' ? 'success' : 'danger'" effect="dark" size="small">
              {{ stats.wireguardStatus === 'running' ? '运行中' : '已停止' }}
            </el-tag>
          </div>
        </template>
        <div class="info-list">
          <div class="info-item">
            <span class="label">客户端数</span>
            <span class="value highlight">{{ stats.wireguardPeers }}</span>
          </div>
          <div class="info-item">
            <span class="label">网络接口</span>
            <span class="value">{{ wgConfig.interface || 'omniwire' }}</span>
          </div>
          <div class="info-item">
            <span class="label">监听端口</span>
            <span class="value">{{ wgConfig.listenPort || '51820' }} / UDP</span>
          </div>
          <div class="info-item">
            <span class="label">公钥</span>
            <el-tooltip :content="wgConfig.publicKey || '未配置'" placement="top">
              <span class="value key">{{ wgConfig.publicKey ? wgConfig.publicKey.substring(0, 20) + '...' : '未配置' }}</span>
            </el-tooltip>
          </div>
        </div>
        <div class="card-action">
          <el-button text type="primary" @click="$router.push('/wireguard')">管理 WireGuard →</el-button>
        </div>
      </el-card>

      <!-- OpenVPN -->
      <el-card class="vpn-card">
        <template #header>
          <div class="card-header">
            <div class="card-title">
              <el-icon><Lock /></el-icon>
              <span>OpenVPN</span>
            </div>
            <el-tag :type="ovpnStatus.running ? 'success' : 'danger'" effect="dark" size="small">
              {{ ovpnStatus.running ? '运行中' : '已停止' }}
            </el-tag>
          </div>
        </template>
        <div class="info-list">
          <div class="info-item">
            <span class="label">在线用户</span>
            <span class="value highlight">{{ ovpnStatus.clientCount || 0 }}</span>
          </div>
          <div class="info-item">
            <span class="label">协议</span>
            <span class="value">{{ (ovpnConfig.protocol || 'udp').toUpperCase() }}</span>
          </div>
          <div class="info-item">
            <span class="label">监听端口</span>
            <span class="value">{{ ovpnConfig.port || '1194' }} / {{ (ovpnConfig.protocol || 'udp').toUpperCase() }}</span>
          </div>
          <div class="info-item">
            <span class="label">路由模式</span>
            <span class="value">{{ ovpnConfig.routeMode === 'split' ? '分流路由' : '全局路由' }}</span>
          </div>
        </div>
        <div class="card-action">
          <el-button text type="primary" @click="$router.push('/openvpn')">管理 OpenVPN →</el-button>
        </div>
      </el-card>
    </div>

    <!-- 快捷操作 -->
    <div class="section-title">快捷操作</div>
    <div class="quick-actions">
      <el-button class="action-btn" size="large" @click="$router.push('/wireguard')">
        <div class="btn-content">
          <div class="btn-icon primary"><el-icon><Plus /></el-icon></div>
          <div class="btn-text">
            <span class="main">添加 WG 客户端</span>
            <span class="sub">配置新的 WireGuard 连接</span>
          </div>
        </div>
      </el-button>

      <el-button class="action-btn" size="large" @click="$router.push('/openvpn')">
        <div class="btn-content">
          <div class="btn-icon success"><el-icon><Plus /></el-icon></div>
          <div class="btn-text">
            <span class="main">添加 OVPN 用户</span>
            <span class="sub">配置新的 OpenVPN 连接</span>
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { systemApi, wireguardApi, openvpnApi } from '@/api'
import { User, Switch, Link, Connection, Lock, Plus, Search } from '@element-plus/icons-vue'

const stats = ref({
  wireguardStatus: 'stopped',
  wireguardPeers: 0,
  forwardRules: 0,
  activeConnections: 0
})

const wgConfig = ref({ interface: 'omniwire', listenPort: 51820, publicKey: '' })
const ovpnStatus = ref({ running: false, clientCount: 0 })
const ovpnConfig = ref({ protocol: 'udp', port: 1194, routeMode: 'full' })

const loadData = async () => {
  try {
    const [dashboardRes, wgStatusRes, ovpnStatusRes, ovpnConfigRes] = await Promise.all([
      systemApi.dashboard(),
      wireguardApi.status(),
      openvpnApi.status(),
      openvpnApi.config()
    ])
    if (dashboardRes.data) stats.value = dashboardRes.data
    if (wgStatusRes.data) {
      wgConfig.value = wgStatusRes.data
      stats.value.wireguardStatus = wgStatusRes.data.running ? 'running' : 'stopped'
    }
    if (ovpnStatusRes.data) ovpnStatus.value = ovpnStatusRes.data
    if (ovpnConfigRes.data) ovpnConfig.value = ovpnConfigRes.data
  } catch (err) {
    console.error('加载数据失败:', err)
  }
}

onMounted(() => { loadData() })
</script>

<style scoped>
.dashboard {
  animation: fadeIn 0.4s ease-out;
  max-width: 1200px;
  margin: 0 auto;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 28px;
}

.stat-card .icon.primary { background: rgba(99, 102, 241, 0.1); color: var(--primary-color); }
.stat-card .icon.info { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.stat-card .icon.warning { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }

.sub-label {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

/* VPN 分组卡片 */
.vpn-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 28px;
}

.vpn-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.vpn-card .card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.info-list { display: flex; flex-direction: column; }

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color);
}
.info-item:last-child { border-bottom: none; }
.info-item .label { color: var(--text-secondary); font-size: 14px; }
.info-item .value { color: var(--text-primary); font-weight: 500; }
.info-item .value.highlight { color: var(--primary-color); font-size: 18px; font-weight: 700; }
.info-item .value.key {
  font-family: monospace;
  font-size: 12px;
  background: var(--bg-hover);
  padding: 4px 8px;
  border-radius: 4px;
}

.card-action {
  margin-top: 12px;
  text-align: right;
}

/* 快捷操作 */
.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 32px;
}

.action-btn {
  height: auto;
  padding: 20px;
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

.btn-content { display: flex; align-items: center; gap: 14px; text-align: left; }
.btn-icon {
  width: 44px; height: 44px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  font-size: 22px; flex-shrink: 0;
}
.btn-icon.primary { background: rgba(99, 102, 241, 0.1); color: var(--primary-color); }
.btn-icon.success { background: rgba(16, 185, 129, 0.1); color: #10b981; }
.btn-icon.info { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.btn-icon.warning { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }

.btn-text { display: flex; flex-direction: column; gap: 3px; }
.btn-text .main { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.btn-text .sub { font-size: 12px; color: var(--text-secondary); }

@media (max-width: 1200px) {
  .stats-grid { grid-template-columns: repeat(3, 1fr); }
  .quick-actions { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: 1fr; }
  .vpn-grid { grid-template-columns: 1fr; }
  .quick-actions { grid-template-columns: 1fr; }
}
</style>
