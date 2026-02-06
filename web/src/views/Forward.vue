<template>
  <div class="forward-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>转发规则</span>
          <div class="header-actions">
            <el-select v-model="refreshInterval" placeholder="刷新频率" style="width: 120px; margin-right: 12px" @change="handleIntervalChange">
              <el-option label="暂停刷新" :value="0" />
              <el-option label="1秒刷新" :value="1000" />
              <el-option label="3秒刷新" :value="3000" />
              <el-option label="5秒刷新" :value="5000" />
              <el-option label="10秒刷新" :value="10000" />
            </el-select>
            <el-button type="primary" @click="showAddDialog = true">
              <el-icon><Plus /></el-icon>
              添加规则
            </el-button>
          </div>
        </div>
      </template>
      
      <el-table :data="rules" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column label="协议" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.protocol === 'tcp' ? 'primary' : 'warning'">
              {{ row.protocol.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="监听端口" width="100">
          <template #default="{ row }">
            <code>{{ row.listenPort }}</code>
          </template>
        </el-table-column>
        <el-table-column label="目标地址" min-width="160">
          <template #default="{ row }">
            <code>{{ row.targetAddr }}:{{ row.targetPort }}</code>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.running ? 'success' : 'info'" size="small">
              {{ row.running ? '运行中' : '已停止' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="实时速率" width="140">
          <template #default="{ row }">
            <div class="rate-info">
              <span>↑{{ formatSpeed(row.uploadSpeed) }}</span>
              <span>↓{{ formatSpeed(row.downloadSpeed) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="速率限制" width="140">
          <template #default="{ row }">
            <div class="rate-info">
              <span>↑{{ formatLimit(row.uploadLimit) }}</span>
              <span>↓{{ formatLimit(row.downloadLimit) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="历史流量" width="160">
          <template #default="{ row }">
            <div class="traffic-info">
              <span>↑{{ formatBytes(row.totalUpload) }}</span>
              <span>↓{{ formatBytes(row.totalDownload) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="连接数" width="100">
          <template #default="{ row }">
            {{ row.currentConn || 0 }} / {{ row.maxConn }}
          </template>
        </el-table-column>
        <el-table-column label="启用" width="70">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggle(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" v-if="!row.running" @click="handleStart(row)">
              <el-icon><VideoPlay /></el-icon>
            </el-button>
            <el-button link type="warning" v-else @click="handleStop(row)">
              <el-icon><VideoPause /></el-icon>
            </el-button>
            <el-button link type="primary" @click="showStats(row)">
              <el-icon><DataLine /></el-icon>
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
    
    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="showAddDialog" :title="editingRule ? '编辑规则' : '添加规则'" width="550px">
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="ruleForm.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="协议" required>
          <el-radio-group v-model="ruleForm.protocol">
            <el-radio value="tcp">TCP</el-radio>
            <el-radio value="udp">UDP</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="监听端口" required>
          <el-input-number v-model="ruleForm.listenPort" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="目标地址" required>
          <el-input v-model="ruleForm.targetAddr" placeholder="192.168.1.100" />
        </el-form-item>
        <el-form-item label="目标端口" required>
          <el-input-number v-model="ruleForm.targetPort" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="最大连接数">
          <el-input-number v-model="ruleForm.maxConn" :min="1" :max="10000" />
        </el-form-item>
        <el-form-item label="上传速率">
          <el-input-number v-model="ruleForm.uploadLimit" :min="0" :step="1024" />
          <span class="unit">bytes/s (0=无限制)</span>
        </el-form-item>
        <el-form-item label="下载速率">
          <el-input-number v-model="ruleForm.downloadLimit" :min="0" :step="1024" />
          <span class="unit">bytes/s (0=无限制)</span>
        </el-form-item>
        <el-form-item label="立即启用">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="ruleForm.description" type="textarea" rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
    
    <!-- 统计对话框 -->
    <el-dialog v-model="showStatsDialog" title="转发统计" width="500px">
      <div class="stats-grid" v-if="currentStats">
        <div class="stat-item">
          <div class="label">总连接数</div>
          <div class="value">{{ currentStats.totalConn }}</div>
        </div>
        <div class="stat-item">
          <div class="label">当前连接</div>
          <div class="value">{{ currentStats.currentConn }}</div>
        </div>
        <div class="stat-item">
          <div class="label">接收流量</div>
          <div class="value">{{ formatBytes(currentStats.bytesReceived) }}</div>
        </div>
        <div class="stat-item">
          <div class="label">发送流量</div>
          <div class="value">{{ formatBytes(currentStats.bytesSent) }}</div>
        </div>
        <div class="stat-item full-width">
          <div class="label">运行时间</div>
          <div class="value">{{ formatUptime(currentStats.uptime) }}</div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { forwardApi } from '@/api'

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

const formatLimit = (bytesPerSec) => {
  if (!bytesPerSec) return '无限'
  return formatSpeed(bytesPerSec)
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
.forward-page { animation: fadeIn 0.3s ease-out; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; align-items: center; }
.stats-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; }
.stat-item { text-align: center; padding: 16px; background: var(--bg-dark); border-radius: 8px; }
.stat-item .label { color: var(--text-muted); font-size: 12px; margin-bottom: 8px; }
.stat-item .value { color: var(--text-primary); font-size: 24px; font-weight: 600; }
.stat-item.full-width { grid-column: span 2; }
.rate-info, .traffic-info { display: flex; flex-direction: column; gap: 2px; font-size: 12px; }
.rate-info span:first-child, .traffic-info span:first-child { color: #f56c6c; }
.rate-info span:last-child, .traffic-info span:last-child { color: #67c23a; }
.unit { margin-left: 8px; color: var(--text-muted); font-size: 12px; }
</style>

