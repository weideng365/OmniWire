<template>
  <div class="forward-page">
    <el-card class="rules-card">
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <el-icon><Switch /></el-icon>
            <span>转发规则管理</span>
          </div>
          <div class="header-actions">
            <div class="refresh-control">
              <span>自动刷新: </span>
              <el-select v-model="refreshInterval" @change="handleIntervalChange" size="small" style="width: 100px;">
                <el-option label="关闭" :value="0" />
                <el-option label="1秒" :value="1000" />
                <el-option label="3秒" :value="3000" />
                <el-option label="5秒" :value="5000" />
                <el-option label="10秒" :value="10000" />
              </el-select>
            </div>
            <el-button type="primary" @click="showAddDialog = true">
              <el-icon><Plus /></el-icon>
              添加规则
            </el-button>
          </div>
        </div>
      </template>
      
      <el-table :data="rules" style="width: 100%" v-loading="loading" :row-style="{ height: '64px' }">
        <el-table-column prop="name" label="名称" min-width="140">
          <template #default="{ row }">
            <span class="rule-name">{{ row.name }}</span>
            <div class="rule-desc" v-if="row.description">{{ row.description }}</div>
          </template>
        </el-table-column>
        <el-table-column label="协议" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.protocol === 'tcp' ? 'primary' : 'warning'" effect="light" round>
              {{ row.protocol.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="监听端口" width="100">
          <template #default="{ row }">
            <code class="port-tag">{{ row.listenPort }}</code>
          </template>
        </el-table-column>
        <el-table-column label="转发目标" min-width="180">
          <template #default="{ row }">
            <div class="target-info">
              <el-icon><Right /></el-icon>
              <code>{{ row.targetAddr }}:{{ row.targetPort }}</code>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <div class="status-badge" :class="{ running: row.running }">
              <span class="dot"></span>
              {{ row.running ? '运行中' : '已停止' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="实时速率" width="160">
          <template #default="{ row }">
            <div class="rate-info">
              <div class="rate-item up">
                <el-icon><Upload /></el-icon>
                <span>{{ formatSpeed(row.uploadSpeed) }}</span>
              </div>
              <div class="rate-item down">
                <el-icon><Download /></el-icon>
                <span>{{ formatSpeed(row.downloadSpeed) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="连接数" width="100" align="center">
          <template #default="{ row }">
            <div class="conn-count">
              <span class="current">{{ row.currentConn || 0 }}</span>
              <span class="max">/{{ row.maxConn }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="80" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggle(row)" 
              style="--el-switch-on-color: var(--success); --el-switch-off-color: var(--text-muted)"/>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-tooltip :content="row.running ? '停止' : '启动'" placement="top">
              <el-button circle size="small" :type="row.running ? 'warning' : 'success'" plain @click="row.running ? handleStop(row) : handleStart(row)">
                <el-icon><component :is="row.running ? 'VideoPause' : 'VideoPlay'" /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="统计详情" placement="top">
              <el-button circle size="small" type="info" plain @click="showStats(row)">
                <el-icon><DataLine /></el-icon>
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
    
    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="showAddDialog" :title="editingRule ? '编辑转发规则' : '添加转发规则'" width="600px" destroy-on-close>
      <el-form :model="ruleForm" label-width="100px" label-position="right">
        <el-form-item label="名称" required>
          <el-input v-model="ruleForm.name" placeholder="给规则起个名字，如：Web Server" />
        </el-form-item>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="协议" required>
              <el-radio-group v-model="ruleForm.protocol">
                <el-radio-button label="tcp">TCP</el-radio-button>
                <el-radio-button label="udp">UDP</el-radio-button>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="监听端口" required>
              <el-input-number v-model="ruleForm.listenPort" :min="1" :max="65535" style="width: 100%" controls-position="right" placeholder="例如: 8080"/>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-divider content-position="left">转发目标</el-divider>
        
        <el-row :gutter="20">
          <el-col :span="16">
            <el-form-item label="目标地址" required>
              <el-input v-model="ruleForm.targetAddr" placeholder="例如: 192.168.1.100" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="目标端口" required label-width="70px">
              <el-input-number v-model="ruleForm.targetPort" :min="1" :max="65535" style="width: 100%" controls-position="right" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-divider content-position="left">高级选项</el-divider>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="最大连接数">
              <el-input-number v-model="ruleForm.maxConn" :min="1" :max="10000" style="width: 100%" controls-position="right"/>
            </el-form-item>
          </el-col>
          <el-col :span="12">
             <el-form-item label="立即启用">
              <el-switch v-model="ruleForm.enabled" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="上传限速">
              <el-input v-model="ruleForm.uploadLimit" placeholder="0 为不限速">
                <template #append>B/s</template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="下载限速">
              <el-input v-model="ruleForm.downloadLimit" placeholder="0 为不限速">
                <template #append>B/s</template>
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="备注">
          <el-input v-model="ruleForm.description" type="textarea" rows="2" placeholder="可选备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存配置</el-button>
      </template>
    </el-dialog>
    
    <!-- 统计对话框 -->
    <el-dialog v-model="showStatsDialog" title="流量统计详情" width="600px" center destroy-on-close>
      <div class="stats-container" v-if="currentStats">
        <div class="stats-header">
          <div class="uptime-badge">
            <el-icon><Timer /></el-icon>
            已运行: {{ formatUptime(currentStats.uptime) }}
          </div>
        </div>
        
        <div class="stats-cards">
          <div class="stat-box primary">
            <div class="stat-icon"><el-icon><Connection /></el-icon></div>
            <div class="stat-p">当前连接</div>
            <div class="stat-v">{{ currentStats.currentConn }}</div>
            <div class="stat-l">总连接: {{ currentStats.totalConn }}</div>
          </div>
          
          <div class="stat-box success">
            <div class="stat-icon"><el-icon><Download /></el-icon></div>
            <div class="stat-p">接收流量</div>
            <div class="stat-v">{{ formatBytes(currentStats.bytesReceived) }}</div>
            <div class="stat-l">速率: {{ formatSpeed(currentStats.downloadSpeed || 0) }}</div>
          </div>
          
           <div class="stat-box warning">
            <div class="stat-icon"><el-icon><Upload /></el-icon></div>
            <div class="stat-p">发送流量</div>
            <div class="stat-v">{{ formatBytes(currentStats.bytesSent) }}</div>
             <div class="stat-l">速率: {{ formatSpeed(currentStats.uploadSpeed || 0) }}</div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { forwardApi } from '@/api'
import { Switch, Plus, Right, VideoPlay, VideoPause, DataLine, Edit, Delete, Upload, Download, Timer, Connection } from '@element-plus/icons-vue'

const loading = ref(false)
const saving = ref(false)
const rules = ref([])
const showAddDialog = ref(false)
const showStatsDialog = ref(false)
const editingRule = ref(null)
const currentStats = ref(null)
const refreshInterval = ref(3000)
let timer = null

const ruleForm = ref({
  name: '', protocol: 'tcp', listenPort: 8080,
  targetAddr: '', targetPort: 80, maxConn: 1000,
  uploadLimit: 0, downloadLimit: 0,
  enabled: true, description: ''
})

const loadRules = async () => {
  // loading.value = true // 定时刷新时不显示 loading，避免闪烁
  try {
    const res = await forwardApi.list({ page: 1, pageSize: 100 })
    rules.value = res.data?.rules || []
  } catch (err) { console.error(err) }
  // loading.value = false
}

// 首次加载显示 loading
const initRules = async () => {
    loading.value = true
    await loadRules()
    loading.value = false
}

const handleSave = async () => {
  if (!ruleForm.value.name || !ruleForm.value.targetAddr) {
    ElMessage.warning('请填写完整信息')
    return
  }
  saving.value = true
  try {
    if (editingRule.value) {
      await forwardApi.update(editingRule.value.id, ruleForm.value)
    } else {
      await forwardApi.create(ruleForm.value)
    }
    ElMessage.success('保存成功')
    showAddDialog.value = false
    resetForm()
    await loadRules()
  } catch (err) { console.error(err) }
  saving.value = false
}

const handleEdit = (row) => {
  editingRule.value = row
  ruleForm.value = { ...row }
  showAddDialog.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除规则 "${row.name}" 吗？`, '确认删除', { type: 'warning' })
    await forwardApi.delete(row.id)
    ElMessage.success('删除成功')
    await loadRules()
  } catch (err) { /* cancelled */ }
}

const handleStart = async (row) => {
  try {
    await forwardApi.start(row.id)
    ElMessage.success('规则已启动')
    await loadRules()
  } catch (err) { console.error(err) }
}

const handleStop = async (row) => {
  try {
    await forwardApi.stop(row.id)
    ElMessage.success('规则已停止')
    await loadRules()
  } catch (err) { console.error(err) }
}

const handleToggle = async (row) => {
  try {
    await forwardApi.update(row.id, { enabled: row.enabled })
    if (row.enabled) await forwardApi.start(row.id)
    else await forwardApi.stop(row.id)
    ElMessage.success(row.enabled ? '已启用' : '已禁用')
    await loadRules()
  } catch (err) { row.enabled = !row.enabled }
}

const showStats = async (row) => {
  try {
    const res = await forwardApi.stats(row.id)
    currentStats.value = res.data?.stats || {}
    showStatsDialog.value = true
  } catch (err) { console.error(err) }
}

const resetForm = () => {
  editingRule.value = null
  ruleForm.value = {
    name: '', protocol: 'tcp', listenPort: 8080,
    targetAddr: '', targetPort: 80, maxConn: 1000,
    uploadLimit: 0, downloadLimit: 0,
    enabled: true, description: ''
  }
}

const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) { bytes /= 1024; i++ }
  return `${bytes.toFixed(1)} ${units[i]}`
}

const formatSpeed = (bytesPerSec) => {
  if (!bytesPerSec) return '0 B/s'
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  let i = 0
  while (bytesPerSec >= 1024 && i < units.length - 1) { bytesPerSec /= 1024; i++ }
  return `${bytesPerSec.toFixed(0)} ${units[i]}`
}

const formatUptime = (seconds) => {
  if (!seconds) return '0秒'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (d > 0) return `${d}天${h}时${m}分`
  if (h > 0) return `${h}时${m}分${s}秒`
  if (m > 0) return `${m}分${s}秒`
  return `${s}秒`
}

const handleIntervalChange = () => {
  if (timer) clearInterval(timer)
  if (refreshInterval.value > 0) {
    // 立即刷新一次
    loadRules()
    timer = setInterval(loadRules, refreshInterval.value)
  }
}

onMounted(() => {
  initRules()
  if (refreshInterval.value > 0) {
    timer = setInterval(loadRules, refreshInterval.value)
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.forward-page { 
  animation: fadeIn 0.4s ease-out; 
}

/* 顶部卡片 */
.rules-card {
  border: none;
  background: transparent;
  box-shadow: none;
}

.rules-card :deep(.el-card__body) {
  padding: 0;
  background: transparent;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.header-title .el-icon {
  background: rgba(var(--primary-rgb), 0.1);
  padding: 8px;
  border-radius: 8px;
  color: var(--primary-color);
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

/* 表格样式 */
.rule-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--text-primary);
}

.rule-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.port-tag {
  font-family: monospace;
  background: var(--bg-hover);
  padding: 2px 6px;
  border-radius: 4px;
  color: var(--primary-color);
  font-weight: 600;
}

.target-info {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-secondary);
  font-family: monospace;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-muted);
}

.status-badge.running {
  color: var(--success);
}

.status-badge .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
}

.rate-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  font-family: monospace;
}

.rate-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.rate-item.up { color: #f59e0b; }
.rate-item.down { color: #10b981; }

.conn-count {
  font-family: monospace;
}

.conn-count .current { font-weight: 600; color: var(--text-primary); }
.conn-count .max { color: var(--text-muted); font-size: 12px; }

/* 统计弹窗 */
.stats-container {
  padding: 10px 0;
}

.stats-header {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

.uptime-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-hover);
  padding: 6px 16px;
  border-radius: 20px;
  color: var(--text-primary);
  font-size: 14px;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.stat-box {
  background: var(--bg-hover);
  padding: 20px;
  border-radius: 12px;
  text-align: center;
  position: relative;
  overflow: hidden;
}

.stat-box::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: currentColor;
  opacity: 0.5;
}

.stat-box.primary { color: var(--primary-color); }
.stat-box.success { color: var(--success); }
.stat-box.warning { color: var(--warning); }

.stat-icon {
  font-size: 24px;
  margin-bottom: 12px;
  opacity: 0.8;
}

.stat-p {
  color: var(--text-secondary);
  font-size: 13px;
  margin-bottom: 8px;
}

.stat-v {
  color: var(--text-primary);
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 8px;
}

.stat-l {
  color: var(--text-muted);
  font-size: 12px;
}
</style>
