<template>
  <div class="wireguard-page">
    <!-- 状态卡片 -->
    <el-card class="status-card">
      <div class="status-header">
        <div class="status-info">
          <div class="status-indicator" :class="{ online: wgStatus.running }">
            <div class="status-pulse" v-if="wgStatus.running"></div>
          </div>
          <div class="status-text-group">
            <span class="status-title">{{ wgStatus.running ? 'WireGuard 服务运行中' : 'WireGuard 服务已停止' }}</span>
            <span class="status-subtitle">{{ wgStatus.running ? '正在监听并处理连接请求' : '服务已关闭，无法建立连接' }}</span>
          </div>
        </div>
        <div class="status-actions">
          <el-button v-if="!wgStatus.running" type="success" @click="handleStart" :loading="loading" plain>
            <el-icon><VideoPlay /></el-icon>
            启动服务
          </el-button>
          <el-button v-else type="danger" @click="handleStop" :loading="loading" plain>
            <el-icon><VideoPause /></el-icon>
            停止服务
          </el-button>
          <el-button @click="handleRestart" :loading="loading">
            <el-icon><Refresh /></el-icon>
            重启
          </el-button>
          <el-button type="primary" @click="showConfigDialog = true">
            <el-icon><Setting /></el-icon>
            服务配置
          </el-button>
        </div>
      </div>
      <div class="status-details">
        <div class="detail-item">
          <span class="label">网络接口</span>
          <span class="value">{{ wgStatus.interface || 'omniwire' }}</span>
        </div>
        <div class="detail-item">
          <span class="label">监听端口</span>
          <span class="value highlight">{{ wgStatus.listenPort || 51820 }}</span>
        </div>
        <div class="detail-item">
          <span class="label">服务器公钥</span>
          <el-tooltip :content="wgStatus.publicKey" placement="top">
            <span class="value key">{{ truncateKey(wgStatus.publicKey) }}</span>
          </el-tooltip>
        </div>
        <div class="detail-item">
          <span class="label">连接客户端</span>
          <span class="value highlight">{{ wgStatus.peerCount || 0 }}</span>
        </div>
      </div>
    </el-card>
    
    <!-- 客户端管理 -->
    <el-card class="peers-card">
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <el-icon><User /></el-icon>
            <span>客户端管理</span>
          </div>
          <div class="header-actions">
            <div class="refresh-control">
              <span>自动刷新: </span>
              <el-select v-model="refreshInterval" @change="onRefreshIntervalChange" size="small" style="width: 100px;">
                <el-option :value="0" label="关闭" />
                <el-option :value="3" label="3秒" />
                <el-option :value="5" label="5秒" />
                <el-option :value="10" label="10秒" />
                <el-option :value="30" label="30秒" />
              </el-select>
            </div>
            <el-button type="primary" @click="showAddDialog = true">
              <el-icon><Plus /></el-icon>
              添加客户端
            </el-button>
          </div>
        </div>
      </template>
      
      <el-table :data="peers" style="width: 100%" v-loading="tableLoading" :row-style="{ height: '60px' }">
        <el-table-column prop="id" label="ID" width="60" align="center" />
        <el-table-column prop="name" label="名称" min-width="120">
          <template #default="{ row }">
            <span class="peer-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <div class="peer-status" :class="{ online: row.online }">
              <span class="dot"></span>
              {{ row.online ? '在线' : '离线' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="IP 地址" min-width="140">
          <template #default="{ row }">
            <code class="ip-tag">{{ row.allowedIPs }}</code>
          </template>
        </el-table-column>
        <el-table-column label="最后握手" min-width="140">
          <template #default="{ row }">
            <span class="time-text">{{ row.latestHandshake || '从未连接' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="实时流量" min-width="160">
          <template #default="{ row }">
            <div class="traffic-stats">
              <span class="rx"><el-icon><Download /></el-icon> {{ formatBytes(row.transferRx) }}</span>
              <span class="tx"><el-icon><Upload /></el-icon> {{ formatBytes(row.transferTx) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="启用/禁用" width="100" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggle(row)" 
              style="--el-switch-on-color: var(--success); --el-switch-off-color: var(--text-muted)"/>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-tooltip content="扫码连接" placement="top">
              <el-button circle size="small" @click="showQRCode(row)">
                <el-icon><Cellphone /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="下载配置" placement="top">
              <el-button circle size="small" @click="downloadConfig(row)">
                <el-icon><Download /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="编辑" placement="top">
              <el-button circle size="small" @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <el-button circle size="small" type="danger" plain @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 添加/编辑客户端对话框 -->
    <el-dialog v-model="showAddDialog" :title="editingPeer ? '编辑客户端' : '添加客户端'" width="480px" destroy-on-close>
      <el-form :model="peerForm" label-width="80px" label-position="top">
        <el-form-item label="名称" required>
          <el-input v-model="peerForm.name" placeholder="例如：iPhone, Macbook Pro" size="large">
            <template #prefix><el-icon><User /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="IP 地址" v-if="!editingPeer">
          <el-input v-model="peerForm.allowedIPs" placeholder="留空自动分配" size="large">
            <template #prefix><el-icon><Connection /></el-icon></template>
          </el-input>
          <div class="form-tip">留空将自动从地址池分配下一个可用 IP</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSavePeer" :loading="loading">
          {{ editingPeer ? '保存修改' : '立即添加' }}
        </el-button>
      </template>
    </el-dialog>
    
    <!-- 配置对话框 -->
    <el-dialog v-model="showConfigDialog" title="WireGuard 服务配置" width="800px" custom-class="config-dialog">
      <el-form :model="configForm" label-width="100px" size="default">
        <!-- 密钥信息（只读） -->
        <div class="config-section">
          <div class="section-header">密钥信息</div>
          <div class="section-tip">密钥由系统自动生成，无需手动管理。公钥用于客户端配置，私钥请妥善保管。</div>
          <el-row :gutter="24">
            <el-col :span="12">
              <el-form-item label="公钥">
                <div class="key-display">
                  <code class="key-text">{{ configForm.publicKey || '未生成' }}</code>
                  <el-button link @click="copyToClipboard(configForm.publicKey, '公钥')">
                    <el-icon><DocumentCopy /></el-icon>
                  </el-button>
                </div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="私钥">
                <div class="key-display">
                  <code class="key-text">{{ showPrivateKey ? (configForm.privateKey || '未生成') : '••••••••••••••••••••' }}</code>
                  <el-button link @click="showPrivateKey = !showPrivateKey">
                    <el-icon><View v-if="!showPrivateKey" /><Hide v-else /></el-icon>
                  </el-button>
                  <el-button link @click="copyToClipboard(configForm.privateKey, '私钥')">
                    <el-icon><DocumentCopy /></el-icon>
                  </el-button>
                </div>
              </el-form-item>
            </el-col>
          </el-row>
        </div>

        <div class="config-section">
          <div class="section-header">网络参数</div>
          <el-row :gutter="24">
            <el-col :span="12">
              <el-form-item label="公网地址">
                <el-input v-model="configForm.endpointAddress" placeholder="IP 或域名" />
                <div class="form-tip">服务器公网 IP 或域名</div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="VPN 网段">
                <el-input v-model="configForm.address" placeholder="10.66.66.0/24" />
                <div class="form-tip">VPN 内网地址段</div>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="24">
            <el-col :span="12">
              <el-form-item label="监听端口">
                <el-input-number v-model="configForm.listenPort" :min="1" :max="65535" style="width: 100%" controls-position="right"/>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="MTU">
                <el-input-number v-model="configForm.mtu" :min="1280" :max="1500" style="width: 100%" controls-position="right"/>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="24">
            <el-col :span="12">
              <el-form-item label="DNS">
                <el-input v-model="configForm.dns" placeholder="223.5.5.5" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="客户端路由">
                <el-input v-model="configForm.clientAllowedIPs" placeholder="自动生成" />
              </el-form-item>
            </el-col>
          </el-row>
        </div>
        
        <div class="config-section">
          <div class="section-header">高级设置</div>
          <el-row :gutter="24">
            <el-col :span="12">
              <el-form-item label="网卡接口">
                <el-input v-model="configForm.ethDevice" placeholder="如 eth0（可选）" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="存活间隔">
                <el-input-number v-model="configForm.persistentKeepalive" :min="0" :max="3600" style="width: 100%" controls-position="right"/>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="24">
            <el-col :span="12">
              <el-form-item label="TCP 中转">
                <el-input v-model="configForm.proxyAddress" placeholder=":50122 (可选)" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="日志等级">
                <el-select v-model="configForm.logLevel" style="width: 100%">
                  <el-option label="错误 (Error)" value="error" />
                  <el-option label="警告 (Warning)" value="warning" />
                  <el-option label="信息 (Info)" value="info" />
                  <el-option label="调试 (Debug)" value="debug" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="开机自启">
            <el-switch v-model="configForm.autoStart" />
            <span style="margin-left: 12px; font-size: 13px; color: var(--text-muted);">服务启动时自动启动 WireGuard</span>
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="showConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveConfig" :loading="loading">保存配置</el-button>
      </template>
    </el-dialog>
    
    <!-- 二维码对话框 -->
    <el-dialog v-model="showQRDialog" title="扫描二维码" width="380px" center destroy-on-close>
      <div class="qrcode-container">
        <div class="qr-wrapper">
          <img :src="qrcodeData" alt="QR Code" v-if="qrcodeData" />
          <div v-else class="qr-placeholder" v-loading="true"></div>
        </div>
        <p class="qr-tip">请使用 WireGuard 客户端扫描此二维码</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { wireguardApi } from '@/api'
import { DocumentCopy, View, Hide, VideoPlay, VideoPause, Refresh, Setting, Plus, User, Download, Edit, Delete, Cellphone, Upload, Connection } from '@element-plus/icons-vue'

const loading = ref(false)
const tableLoading = ref(false)
const showPrivateKey = ref(false)
const wgStatus = ref({})
const peers = ref([])
const showAddDialog = ref(false)
const showConfigDialog = ref(false)
const showQRDialog = ref(false)
const editingPeer = ref(null)
const qrcodeData = ref('')

// 自动刷新配置
const refreshInterval = ref(parseInt(localStorage.getItem('wg_refresh_interval') || '5'))
let refreshTimer = null

const peerForm = ref({ name: '', allowedIPs: '' })
const configForm = ref({
  listenPort: 51820,
  address: '',
  dns: '223.5.5.5',
  mtu: 1420,
  endpointAddress: '',
  ethDevice: '',
  persistentKeepalive: 25,
  clientAllowedIPs: '',
  proxyAddress: '',
  logLevel: 'error',
  autoStart: false
})

// VPN 网段变化时自动同步客户端路由
watch(() => configForm.value.address, (val) => {
  if (!val) return
  const match = val.match(/^(\d+\.\d+\.\d+\.\d+\/\d+)$/)
  if (match) {
    // 解析 CIDR，计算网络地址
    const parts = val.split('/')
    const mask = parseInt(parts[1])
    const ipParts = parts[0].split('.').map(Number)
    // 计算网络地址
    const maskBits = mask >= 32 ? 0xFFFFFFFF : (0xFFFFFFFF << (32 - mask)) >>> 0
    const ipNum = ((ipParts[0] << 24) | (ipParts[1] << 16) | (ipParts[2] << 8) | ipParts[3]) >>> 0
    const netNum = (ipNum & maskBits) >>> 0
    const netAddr = [(netNum >>> 24) & 0xFF, (netNum >>> 16) & 0xFF, (netNum >>> 8) & 0xFF, netNum & 0xFF].join('.')
    configForm.value.clientAllowedIPs = netAddr + '/' + mask
  }
})

// 启动自动刷新
const startAutoRefresh = () => {
  stopAutoRefresh()
  if (refreshInterval.value > 0) {
    refreshTimer = setInterval(() => {
      if (!tableLoading.value) {
        loadPeers()
        loadStatus()
      }
    }, refreshInterval.value * 1000)
  }
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 刷新间隔变化
const onRefreshIntervalChange = (val) => {
  localStorage.setItem('wg_refresh_interval', val.toString())
  startAutoRefresh()
}

const loadStatus = async () => {
  try {
    const res = await wireguardApi.status()
    wgStatus.value = res.data || {}
  } catch (err) { console.error(err) }
}

const loadPeers = async () => {
  tableLoading.value = true
  try {
    const res = await wireguardApi.peers()
    peers.value = res.data?.peers || []
  } catch (err) { console.error(err) }
  tableLoading.value = false
}

const loadConfig = async () => {
  try {
    const res = await wireguardApi.config()
    if (res.data) {
      configForm.value = res.data
    }
  } catch (err) { console.error(err) }
}

const handleStart = async () => {
  loading.value = true
  try {
    await wireguardApi.start()
    ElMessage.success('服务已启动')
    await loadStatus()
    await loadConfig()
  } catch (err) { console.error(err) }
  loading.value = false
}

const handleStop = async () => {
  loading.value = true
  try {
    await wireguardApi.stop()
    ElMessage.success('服务已停止')
    await loadStatus()
  } catch (err) { console.error(err) }
  loading.value = false
}

const handleRestart = async () => {
  loading.value = true
  try {
    await wireguardApi.restart()
    ElMessage.success('服务已重启')
    await loadStatus()
  } catch (err) { console.error(err) }
  loading.value = false
}

const handleSavePeer = async () => {
  if (!peerForm.value.name) {
    ElMessage.warning('请输入客户端名称')
    return
  }
  loading.value = true
  try {
    if (editingPeer.value) {
      await wireguardApi.updatePeer(editingPeer.value.id, peerForm.value)
    } else {
      await wireguardApi.createPeer(peerForm.value)
    }
    ElMessage.success(editingPeer.value ? '保存成功' : '添加成功')
    showAddDialog.value = false
    peerForm.value = { name: '', allowedIPs: '' }
    editingPeer.value = null
    await loadPeers()
  } catch (err) { console.error(err) }
  loading.value = false
}

const handleEdit = (row) => {
  editingPeer.value = row
  peerForm.value = { name: row.name, allowedIPs: row.allowedIPs }
  showAddDialog.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除客户端 "${row.name}" 吗？`, '确认删除', { type: 'warning' })
    await wireguardApi.deletePeer(row.id)
    ElMessage.success('删除成功')
    await loadPeers()
  } catch (err) { /* cancelled */ }
}

const handleToggle = async (row) => {
  try {
    await wireguardApi.updatePeer(row.id, { enabled: row.enabled })
    ElMessage.success(row.enabled ? '已启用' : '已禁用')
  } catch (err) {
    row.enabled = !row.enabled
  }
}

const showQRCode = async (row) => {
  try {
    const res = await wireguardApi.peerQRCode(row.id)
    qrcodeData.value = res.data?.qrcode || ''
    showQRDialog.value = true
  } catch (err) { console.error(err) }
}

const downloadConfig = async (row) => {
  try {
    const res = await wireguardApi.peerConfig(row.id)
    const config = res.data?.config || res.config || ''
    if (!config) {
      ElMessage.error(res.message || '获取配置失败，请检查是否已配置公网地址')
      return
    }
    const blob = new Blob([config], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${row.name}.conf`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err) {
    console.error(err)
    ElMessage.error('下载配置失败，请检查是否已配置公网地址')
  }
}

const handleSaveConfig = async () => {
  if (!configForm.value.endpointAddress?.trim()) {
    ElMessage.warning('请填写公网地址，客户端需要此地址连接服务器')
    return
  }
  loading.value = true
  try {
    await wireguardApi.updateConfig(configForm.value)
    ElMessage.success('配置已保存')
    showConfigDialog.value = false
    await loadStatus()
    await loadConfig()  // 重新加载配置确保数据同步
  } catch (err) { console.error(err) }
  loading.value = false
}

const truncateKey = (key) => key ? `${key.slice(0, 10)}...${key.slice(-6)}` : '未配置'

const copyToClipboard = async (text, name) => {
  if (!text) {
    ElMessage.warning(`${name}为空`)
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(`${name}已复制`)
  } catch (err) {
    ElMessage.error('复制失败')
  }
}

const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) { bytes /= 1024; i++ }
  return `${bytes.toFixed(1)} ${units[i]}`
}

onMounted(() => {
  loadStatus()
  loadPeers()
  loadConfig()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.wireguard-page {
  animation: fadeIn 0.4s ease-out;
  padding-bottom: 24px;
}

/* 状态卡片 */
.status-card {
  margin-bottom: 24px;
  background: linear-gradient(135deg, var(--bg-card) 0%, rgba(var(--primary-rgb), 0.05) 100%);
  border: 1px solid var(--border-color);
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
  flex-wrap: wrap;
  gap: 20px;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.status-indicator {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--text-muted);
  position: relative;
  transition: all 0.3s ease;
}

.status-indicator.online {
  background: var(--success);
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.4);
}

.status-pulse {
  position: absolute;
  top: -4px;
  left: -4px;
  right: -4px;
  bottom: -4px;
  border-radius: 50%;
  border: 2px solid var(--success);
  opacity: 0.5;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(1); opacity: 0.5; }
  100% { transform: scale(2); opacity: 0; }
}

.status-text-group {
  display: flex;
  flex-direction: column;
}

.status-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.status-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.status-actions {
  display: flex;
  gap: 12px;
}

.status-details {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
  padding: 24px;
  background: var(--bg-hover);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
}

@media (max-width: 1000px) {
  .status-details {
    grid-template-columns: repeat(2, 1fr);
  }
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-item .label {
  color: var(--text-secondary);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-item .value {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 15px;
}

.detail-item .value.highlight {
  color: var(--primary-color);
  font-size: 18px;
}

.detail-item .value.key {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
  color: var(--text-primary);
}

/* 客户端管理 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.refresh-control {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}

/* 表格样式优化 */
.peer-name {
  font-weight: 600;
  color: var(--text-primary);
}

.peer-status {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-muted);
}

.peer-status.online {
  color: var(--success);
}

.peer-status .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.ip-tag {
  font-family: monospace;
  background: var(--bg-hover);
  padding: 2px 6px;
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 12px;
}

.traffic-stats {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--text-secondary);
}

.traffic-stats span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.traffic-stats .rx { color: var(--success); }
.traffic-stats .tx { color: var(--info); }

/* 配置对话框 */
.config-section {
  margin-bottom: 24px;
  padding: 20px;
  background: var(--bg-hover);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
}

.section-header {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
  padding-left: 10px;
  border-left: 3px solid var(--primary-color);
}

.section-tip {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 16px;
  padding: 8px 12px;
  background: var(--bg-card);
  border-radius: var(--radius-sm);
  border: 1px dashed var(--border-color);
}

.key-display {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 0 12px;
  height: 36px;
  width: 100%;
}

.key-text {
  flex: 1;
  font-size: 12px;
  font-family: monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-primary);
}

.form-tip {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
  line-height: 1.4;
}

/* 二维码 */
.qrcode-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 0;
}

.qr-wrapper {
  background: white;
  padding: 12px;
  border-radius: 12px;
  box-shadow: var(--shadow-md);
}

.qr-wrapper img {
  width: 240px;
  height: 240px;
  display: block;
}

.qr-tip {
  margin-top: 20px;
  font-size: 14px;
  color: var(--text-secondary);
}
</style>
