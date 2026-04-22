<template>
  <div class="openvpn-page">
    <!-- 状态卡片 -->
    <el-card class="status-card">
      <div class="status-header">
        <div class="status-info">
          <div class="status-indicator" :class="{ online: status.running }">
            <div class="status-pulse" v-if="status.running"></div>
          </div>
          <div class="status-text-group">
            <span class="status-title">{{ status.running ? 'OpenVPN 服务运行中' : 'OpenVPN 服务已停止' }}</span>
            <span class="status-subtitle">{{ status.running ? '正在监听并处理连接请求' : '服务已关闭，无法建立连接' }}</span>
          </div>
        </div>
        <div class="status-actions">
          <el-button v-if="!status.running" type="success" @click="handleStart" :loading="loading" plain>
            <el-icon><VideoPlay /></el-icon>启动服务
          </el-button>
          <el-button v-else type="danger" @click="handleStop" :loading="loading" plain>
            <el-icon><VideoPause /></el-icon>停止服务
          </el-button>
          <el-button @click="handleRestart" :loading="loading">
            <el-icon><Refresh /></el-icon>重启
          </el-button>
          <el-button type="primary" @click="showConfigDialog = true">
            <el-icon><Setting /></el-icon>服务配置
          </el-button>
          <el-button @click="$router.push('/openvpn-guide')">
            <el-icon><Document /></el-icon>使用说明
          </el-button>
        </div>
      </div>
      <div class="status-details">
        <div class="detail-item">
          <span class="label">协议</span>
          <span class="value highlight">{{ status.protocol?.toUpperCase() || 'UDP' }}</span>
        </div>
        <div class="detail-item">
          <span class="label">监听端口</span>
          <span class="value highlight">{{ status.port || 1194 }}</span>
        </div>
        <div class="detail-item">
          <span class="label">在线用户</span>
          <span class="value highlight">{{ status.clientCount || 0 }}</span>
        </div>
        <div class="detail-item">
          <span class="label">总上传</span>
          <span class="value highlight">{{ formatBytes(status.txBytes) }}</span>
        </div>
        <div class="detail-item">
          <span class="label">总下载</span>
          <span class="value highlight">{{ formatBytes(status.rxBytes) }}</span>
        </div>
      </div>
    </el-card>

    <!-- 用户管理 -->
    <el-card class="peers-card">
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </div>
          <div class="header-actions">
            <el-select v-model="refreshInterval" size="small" style="width: 120px" @change="onRefreshIntervalChange">
              <el-option :value="0" label="关闭刷新" />
              <el-option :value="3" label="3 秒" />
              <el-option :value="5" label="5 秒" />
              <el-option :value="10" label="10 秒" />
            </el-select>
            <el-button type="primary" @click="showAddDialog = true">
              <el-icon><Plus /></el-icon>添加用户
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="users" style="width: 100%" v-loading="tableLoading" :row-style="{ height: '60px' }">
        <el-table-column prop="id" label="ID" width="60" align="center" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <div class="peer-status" :class="{ online: row.online }">
              <span class="dot"></span>{{ row.online ? '在线' : '离线' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="IP 地址" min-width="130">
          <template #default="{ row }">
            <code class="ip-tag">{{ row.ip || '-' }}</code>
          </template>
        </el-table-column>
        <el-table-column label="上传" min-width="100">
          <template #default="{ row }">
            <span class="time-text">{{ row.online ? formatBytes(row.txBytes) : '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="下载" min-width="100">
          <template #default="{ row }">
            <span class="time-text">{{ row.online ? formatBytes(row.rxBytes) : '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="最后连接" min-width="160">
          <template #default="{ row }">
            <span class="time-text">{{ row.connectedAt || '从未连接' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="90" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggle(row)"
              style="--el-switch-on-color: var(--success); --el-switch-off-color: var(--text-muted)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right" align="center">
          <template #default="{ row }">
            <el-tooltip content="下载配置" placement="top">
              <el-button circle size="small" @click="downloadConfig(row)">
                <el-icon><Download /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="修改密码" placement="top">
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

    <!-- 添加/编辑用户对话框 -->
    <el-dialog v-model="showAddDialog" :title="editingUser ? '修改密码' : '添加用户'" width="420px" destroy-on-close>
      <el-form :model="userForm" label-position="top">
        <el-form-item label="用户名" v-if="!editingUser" required>
          <el-input v-model="userForm.username" placeholder="输入用户名" size="large">
            <template #prefix><el-icon><User /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="密码" required>
          <el-input v-model="userForm.password" type="password" placeholder="输入密码（6-32位）" size="large" show-password>
            <template #prefix><el-icon><Lock /></el-icon></template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveUser" :loading="loading">
          {{ editingUser ? '保存修改' : '立即添加' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 服务配置对话框 -->
    <el-dialog v-model="showConfigDialog" title="OpenVPN 服务配置" width="480px" destroy-on-close>
      <el-form :model="configForm" label-position="top">
        <el-form-item label="协议">
          <el-radio-group v-model="configForm.protocol" size="large">
            <el-radio-button value="udp">UDP（推荐）</el-radio-button>
            <el-radio-button value="tcp">TCP</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="监听端口">
          <el-input-number v-model="configForm.port" :min="1" :max="65535" style="width: 100%" size="large" />
        </el-form-item>
        <el-form-item label="服务器地址">
          <el-input v-model="configForm.endpoint" placeholder="填写服务器外网IP或域名" size="large" />
          <div class="form-tip">客户端连接时使用的服务器地址</div>
        </el-form-item>
        <el-form-item label="VPN 子网">
          <el-input v-model="configForm.subnet" placeholder="10.8.0.0/24" size="large" />
        </el-form-item>
        <el-form-item label="DNS 服务器">
          <el-input v-model="configForm.dns" placeholder="223.5.5.5" size="large" />
        </el-form-item>
        <el-form-item label="路由模式">
          <el-radio-group v-model="configForm.routeMode" size="large">
            <el-radio-button value="full">全流量</el-radio-button>
            <el-radio-button value="split">分流</el-radio-button>
          </el-radio-group>
          <div class="form-tip">全流量：所有流量走VPN；分流：仅指定IP段走VPN</div>
        </el-form-item>
        <el-form-item label="分流路由" v-if="configForm.routeMode === 'split'">
          <el-input v-model="configForm.splitRoutes" type="textarea" :rows="3"
            placeholder="每行或逗号分隔，如：10.0.0.0/8,192.168.1.0/24" size="large" />
          <div class="form-tip">只有这些IP段的流量会走VPN</div>
        </el-form-item>
        <el-form-item label="自动启动">
          <el-switch v-model="configForm.autoStart" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveConfig" :loading="loading">保存配置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { openvpnApi } from '@/api'

const loading = ref(false)
const tableLoading = ref(false)
const status = ref({ running: false, protocol: 'udp', port: 1194, clientCount: 0 })
const users = ref([])
const showAddDialog = ref(false)
const showConfigDialog = ref(false)
const editingUser = ref(null)
const userForm = ref({ username: '', password: '' })
const configForm = ref({ protocol: 'udp', port: 1194, endpoint: '', subnet: '10.8.0.0/24', dns: '223.5.5.5', autoStart: false, routeMode: 'split', splitRoutes: '' })

const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

let refreshTimer = null
let toggling = false
const refreshInterval = ref(parseInt(localStorage.getItem('ovpn_refresh_interval') || '3'))

const fetchStatus = async () => {
  try {
    const res = await openvpnApi.status()
    status.value = res.data
  } catch (e) {}
}

const fetchUsers = async (showLoading = false) => {
  if (toggling) return
  if (showLoading) tableLoading.value = true
  try {
    const res = await openvpnApi.users()
    users.value = res.data.users || []
  } finally {
    if (showLoading) tableLoading.value = false
  }
}

const fetchConfig = async () => {
  try {
    const res = await openvpnApi.config()
    configForm.value = res.data
  } catch (e) {}
}

const handleStart = async () => {
  loading.value = true
  try {
    await openvpnApi.start()
    ElMessage.success('服务已启动')
    await fetchStatus()
  } catch (e) {
    // 响应拦截器已显示错误
  } finally {
    loading.value = false
  }
}

const handleStop = async () => {
  loading.value = true
  try {
    await openvpnApi.stop()
    ElMessage.success('服务已停止')
    await fetchStatus()
  } catch (e) {
    // 响应拦截器已显示错误
  } finally {
    loading.value = false
  }
}

const handleRestart = async () => {
  loading.value = true
  try {
    await openvpnApi.restart()
    ElMessage.success('服务已重启')
    await fetchStatus()
  } catch (e) {
    // 响应拦截器已显示错误
  } finally {
    loading.value = false
  }
}

const handleSaveConfig = async () => {
  loading.value = true
  try {
    await openvpnApi.updateConfig(configForm.value)
    ElMessage.success('配置已保存')
    showConfigDialog.value = false
    await fetchStatus()
  } finally {
    loading.value = false
  }
}

const handleSaveUser = async () => {
  if (!userForm.value.password || userForm.value.password.length < 6) {
    return ElMessage.warning('密码长度至少6位')
  }
  loading.value = true
  try {
    if (editingUser.value) {
      await openvpnApi.updateUser(editingUser.value.id, { password: userForm.value.password, enabled: true })
      ElMessage.success('密码已修改')
    } else {
      if (!userForm.value.username) return ElMessage.warning('请输入用户名')
      await openvpnApi.createUser(userForm.value)
      ElMessage.success('用户已创建')
    }
    showAddDialog.value = false
    await fetchUsers()
  } finally {
    loading.value = false
  }
}

const handleToggle = async (row) => {
  toggling = true
  try {
    await openvpnApi.updateUser(row.id, { enabled: row.enabled })
    ElMessage.success(row.enabled ? '已启用' : '已禁用')
  } catch (e) {
    row.enabled = !row.enabled
  } finally {
    toggling = false
  }
}

const handleEdit = (row) => {
  editingUser.value = row
  userForm.value = { username: row.username, password: '' }
  showAddDialog.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除用户 ${row.username}？`, '提示', { type: 'warning' })
    await openvpnApi.deleteUser(row.id)
    ElMessage.success('已删除')
    await fetchUsers()
  } catch (e) {}
}

const downloadConfig = async (row) => {
  try {
    const res = await openvpnApi.userConfig(row.id)
    const blob = new Blob([res.data.config], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${row.username}.ovpn`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {}
}

const startAutoRefresh = () => {
  stopAutoRefresh()
  if (refreshInterval.value > 0) {
    refreshTimer = setInterval(() => { fetchStatus(); fetchUsers() }, refreshInterval.value * 1000)
  }
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

const onRefreshIntervalChange = () => {
  localStorage.setItem('ovpn_refresh_interval', String(refreshInterval.value))
  startAutoRefresh()
}

onMounted(() => {
  fetchStatus()
  fetchUsers(true)
  fetchConfig()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.openvpn-page {
  padding: 20px;
}

.status-card {
  margin-bottom: 20px;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.status-indicator {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--el-color-danger-light-9);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.status-indicator.online {
  background: var(--el-color-success-light-9);
}

.status-pulse {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--el-color-success);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.9); }
}

.status-text-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-title {
  font-size: 18px;
  font-weight: 600;
}

.status-subtitle {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.status-actions {
  display: flex;
  gap: 10px;
}

.status-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  padding: 20px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-item .label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.detail-item .value {
  font-size: 16px;
  font-weight: 500;
}

.detail-item .value.highlight {
  color: var(--el-color-primary);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.peer-status {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--el-text-color-secondary);
}

.peer-status.online {
  color: var(--el-color-success);
}

.peer-status .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
}

.ip-tag {
  padding: 4px 8px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 13px;
}

.time-text {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
