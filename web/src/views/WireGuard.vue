<template>
  <div class="wireguard-page">
    <!-- 状态卡片 -->
    <el-card class="status-card">
      <div class="status-header">
        <div class="status-info">
          <div class="status-indicator" :class="{ online: wgStatus.running }"></div>
          <span class="status-text">{{ wgStatus.running ? '服务运行中' : '服务已停止' }}</span>
        </div>
        <div class="status-actions">
          <el-button v-if="!wgStatus.running" type="success" @click="handleStart" :loading="loading">
            <el-icon><VideoPlay /></el-icon>
            启动
          </el-button>
          <el-button v-else type="danger" @click="handleStop" :loading="loading">
            <el-icon><VideoPause /></el-icon>
            停止
          </el-button>
          <el-button @click="handleRestart" :loading="loading">
            <el-icon><Refresh /></el-icon>
            重启
          </el-button>
          <el-button @click="showConfigDialog = true">
            <el-icon><Setting /></el-icon>
            配置
          </el-button>
        </div>
      </div>
      <div class="status-details">
        <div class="detail-item">
          <span class="label">接口</span>
          <span class="value">{{ wgStatus.interface || 'wg0' }}</span>
        </div>
        <div class="detail-item">
          <span class="label">监听端口</span>
          <span class="value">{{ wgStatus.listenPort || 51820 }}</span>
        </div>
        <div class="detail-item">
          <span class="label">公钥</span>
          <el-tooltip :content="wgStatus.publicKey" placement="top">
            <span class="value key">{{ truncateKey(wgStatus.publicKey) }}</span>
          </el-tooltip>
        </div>
        <div class="detail-item">
          <span class="label">客户端数量</span>
          <span class="value">{{ wgStatus.peerCount || 0 }}</span>
        </div>
      </div>
    </el-card>
    
    <!-- 客户端管理 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>客户端管理</span>
          <div class="header-actions">
            <el-select v-model="refreshInterval" @change="onRefreshIntervalChange" size="small" style="width: 100px;">
              <el-option :value="0" label="不刷新" />
              <el-option :value="3" label="3秒" />
              <el-option :value="5" label="5秒" />
              <el-option :value="10" label="10秒" />
              <el-option :value="30" label="30秒" />
            </el-select>
            <el-button type="primary" @click="showAddDialog = true">
              <el-icon><Plus /></el-icon>
              添加客户端
            </el-button>
          </div>
        </div>
      </template>
      
      <el-table :data="peers" style="width: 100%" v-loading="tableLoading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.online ? 'success' : 'info'" size="small">
              {{ row.online ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="IP 地址" min-width="140">
          <template #default="{ row }">
            <code>{{ row.allowedIPs }}</code>
          </template>
        </el-table-column>
        <el-table-column label="最后握手" min-width="120">
          <template #default="{ row }">
            {{ row.latestHandshake || '从未连接' }}
          </template>
        </el-table-column>
        <el-table-column label="流量" min-width="150">
          <template #default="{ row }">
            <span class="traffic">
              ↓ {{ formatBytes(row.transferRx) }} / ↑ {{ formatBytes(row.transferTx) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggle(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showQRCode(row)">
              <el-icon><Cellphone /></el-icon>
            </el-button>
            <el-button link type="primary" @click="downloadConfig(row)">
              <el-icon><Download /></el-icon>
            </el-button>
            <el-button link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 添加/编辑客户端对话框 -->
    <el-dialog v-model="showAddDialog" :title="editingPeer ? '编辑客户端' : '添加客户端'" width="500px">
      <el-form :model="peerForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="peerForm.name" placeholder="请输入客户端名称" />
        </el-form-item>
        <el-form-item label="IP 地址" v-if="!editingPeer">
          <el-input v-model="peerForm.allowedIPs" placeholder="留空自动分配" />
          <div class="form-tip">留空将自动从地址池分配 IP</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSavePeer" :loading="loading">
          {{ editingPeer ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
    
    <!-- 配置对话框 -->
    <el-dialog v-model="showConfigDialog" title="WireGuard 配置" width="600px">
      <el-form :model="configForm" label-width="120px">
        <!-- 密钥信息（只读） -->
        <el-divider content-position="left">密钥信息</el-divider>
        <el-form-item label="公钥">
          <el-input v-model="configForm.publicKey" readonly>
            <template #append>
              <el-button @click="copyToClipboard(configForm.publicKey, '公钥')">
                <el-icon><DocumentCopy /></el-icon>
              </el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="私钥">
          <el-input v-model="configForm.privateKey" :type="showPrivateKey ? 'text' : 'password'" readonly>
            <template #append>
              <el-button-group>
                <el-button @click="showPrivateKey = !showPrivateKey">
                  <el-icon><View v-if="!showPrivateKey" /><Hide v-else /></el-icon>
                </el-button>
                <el-button @click="copyToClipboard(configForm.privateKey, '私钥')">
                  <el-icon><DocumentCopy /></el-icon>
                </el-button>
              </el-button-group>
            </template>
          </el-input>
        </el-form-item>

        <el-divider content-position="left">服务配置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="监听端口">
              <el-input-number v-model="configForm.listenPort" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="MTU">
              <el-input-number v-model="configForm.mtu" :min="1280" :max="1500" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="公网地址">
          <el-input v-model="configForm.endpointAddress" placeholder="IP 或域名，客户端连接使用" />
        </el-form-item>

        <el-form-item label="网卡接口">
          <el-input v-model="configForm.ethDevice" placeholder="如 eth0, 留空自动检测" />
        </el-form-item>
        
        <el-form-item label="VPN 网段">
          <el-input v-model="configForm.address" placeholder="10.66.66.1/24" />
        </el-form-item>
        
        <el-form-item label="DNS">
          <el-input v-model="configForm.dns" placeholder="1.1.1.1, 8.8.8.8" />
        </el-form-item>
        
        <el-form-item label="客户端路由">
          <el-input v-model="configForm.clientAllowedIPs" placeholder="0.0.0.0/0, ::/0" />
          <div class="form-tip">客户端默认转发到服务端的 IP 范围</div>
        </el-form-item>
        
        <el-form-item label="存活间隔">
          <el-input-number v-model="configForm.persistentKeepalive" :min="0" :max="3600" />
          <span class="unit">秒</span>
        </el-form-item>
        
        <el-form-item label="TCP 中转">
          <el-input v-model="configForm.proxyAddress" placeholder=":50122 (可选)" />
          <div class="form-tip">用于防止 UDP QoS，留空不开启</div>
        </el-form-item>

        <el-form-item label="日志等级">
          <el-select v-model="configForm.logLevel">
            <el-option label="错误 (Error)" value="error" />
            <el-option label="警告 (Warning)" value="warning" />
            <el-option label="信息 (Info)" value="info" />
            <el-option label="调试 (Debug)" value="debug" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveConfig" :loading="loading">保存</el-button>
      </template>
    </el-dialog>
    
    <!-- 二维码对话框 -->
    <el-dialog v-model="showQRDialog" title="扫描二维码" width="350px" center>
      <div class="qrcode-container">
        <img :src="qrcodeData" alt="QR Code" v-if="qrcodeData" />
        <p>使用 WireGuard 客户端扫描此二维码</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { wireguardApi } from '@/api'
import { DocumentCopy, View, Hide } from '@element-plus/icons-vue'

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
  dns: '',
  mtu: 1420,
  endpointAddress: '',
  ethDevice: 'eth0',
  persistentKeepalive: 25,
  clientAllowedIPs: '0.0.0.0/0, ::/0',
  proxyAddress: '',
  logLevel: 'error'
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
    const blob = new Blob([res.data?.config || ''], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${row.name}.conf`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err) { console.error(err) }
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
.wireguard-page { animation: fadeIn 0.3s ease-out; }

.status-card { margin-bottom: 24px; }

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--text-muted);
}

.status-indicator.online {
  background: var(--success);
  box-shadow: 0 0 8px var(--success);
}

.status-text {
  font-size: 18px;
  font-weight: 600;
}

.status-actions {
  display: flex;
  gap: 12px;
}

.status-details {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item .label {
  color: var(--text-muted);
  font-size: 12px;
}

.detail-item .value {
  color: var(--text-primary);
  font-weight: 500;
}

.detail-item .value.key {
  font-family: monospace;
  font-size: 13px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.traffic {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-secondary);
}

.form-tip {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.qrcode-container {
  text-align: center;
}

.qrcode-container img {
  width: 256px;
  height: 256px;
  border-radius: 8px;
}

.qrcode-container p {
  margin-top: 16px;
  color: var(--text-secondary);
}
</style>
