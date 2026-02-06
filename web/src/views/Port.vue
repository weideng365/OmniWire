<template>
  <div class="port-page">
    <!-- 端口扫描 -->
    <el-card class="scan-card">
      <template #header><span>端口扫描</span></template>
      <div class="scan-form">
        <el-input-number v-model="scanForm.startPort" :min="1" :max="65535" placeholder="起始端口" />
        <span class="separator">-</span>
        <el-input-number v-model="scanForm.endPort" :min="1" :max="65535" placeholder="结束端口" />
        <el-button type="primary" @click="handleScan" :loading="scanning">
          <el-icon><Search /></el-icon>
          扫描
        </el-button>
      </div>
      <div class="scan-results" v-if="scanResults.length">
        <el-tag v-for="port in scanResults" :key="port.port" class="port-tag">
          {{ port.port }}/{{ port.protocol }}
        </el-tag>
      </div>
      <el-empty v-else-if="scanned" description="未发现开放端口" />
    </el-card>
    
    <!-- 监听端口列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>监听端口</span>
          <el-button @click="loadListeningPorts" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <el-table :data="listeningPorts" style="width: 100%" v-loading="loading">
        <el-table-column prop="port" label="端口" width="100" />
        <el-table-column prop="protocol" label="协议" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ row.protocol?.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="140" />
        <el-table-column prop="process" label="进程" min-width="120" />
        <el-table-column prop="pid" label="PID" width="100" />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button link type="primary" @click="checkPort(row.port)">
              <el-icon><View /></el-icon>
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 端口检查对话框 -->
    <el-dialog v-model="showCheckDialog" title="端口详情" width="500px">
      <div class="port-detail" v-if="portDetail">
        <div class="detail-row">
          <span class="label">端口</span>
          <span class="value">{{ checkingPort }}</span>
        </div>
        <div class="detail-row">
          <span class="label">状态</span>
          <el-tag :type="portDetail.inUse ? 'danger' : 'success'">
            {{ portDetail.inUse ? '已占用' : '空闲' }}
          </el-tag>
        </div>
        <div class="detail-row" v-if="portDetail.process">
          <span class="label">进程</span>
          <span class="value">{{ portDetail.process }}</span>
        </div>
        <div class="detail-row" v-if="portDetail.pid">
          <span class="label">PID</span>
          <span class="value">{{ portDetail.pid }}</span>
        </div>
      </div>
      <div class="connections" v-if="connections.length">
        <h4>活跃连接</h4>
        <el-table :data="connections" size="small">
          <el-table-column prop="remoteAddr" label="远程地址" />
          <el-table-column prop="remotePort" label="远程端口" width="100" />
          <el-table-column prop="state" label="状态" width="120" />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { portApi } from '@/api'

const loading = ref(false)
const scanning = ref(false)
const scanned = ref(false)
const scanForm = ref({ startPort: 1, endPort: 1024 })
const scanResults = ref([])
const listeningPorts = ref([])
const showCheckDialog = ref(false)
const checkingPort = ref(0)
const portDetail = ref(null)
const connections = ref([])

const handleScan = async () => {
  scanning.value = true
  scanned.value = false
  try {
    const res = await portApi.scan(scanForm.value)
    scanResults.value = res.data?.ports || []
    scanned.value = true
    ElMessage.success(`扫描完成，发现 ${scanResults.value.length} 个开放端口`)
  } catch (err) { console.error(err) }
  scanning.value = false
}

const loadListeningPorts = async () => {
  loading.value = true
  try {
    const res = await portApi.listen()
    listeningPorts.value = res.data?.ports || []
  } catch (err) { console.error(err) }
  loading.value = false
}

const checkPort = async (port) => {
  checkingPort.value = port
  try {
    const [checkRes, connRes] = await Promise.all([
      portApi.check(port),
      portApi.connections(port)
    ])
    portDetail.value = checkRes.data || {}
    connections.value = connRes.data?.connections || []
    showCheckDialog.value = true
  } catch (err) { console.error(err) }
}

onMounted(() => { loadListeningPorts() })
</script>

<style scoped>
.port-page { animation: fadeIn 0.3s ease-out; }
.scan-card { margin-bottom: 24px; }
.scan-form { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; }
.separator { color: var(--text-muted); }
.scan-results { display: flex; flex-wrap: wrap; gap: 8px; }
.port-tag { font-family: monospace; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.port-detail { margin-bottom: 20px; }
.detail-row { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid var(--border-color); }
.detail-row .label { color: var(--text-muted); }
.detail-row .value { color: var(--text-primary); font-weight: 500; }
.connections h4 { margin-bottom: 12px; color: var(--text-primary); }
</style>
