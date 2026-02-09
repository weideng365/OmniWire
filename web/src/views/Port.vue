<template>
  <div class="port-page">
    <div class="page-header">
      <div class="header-content">
        <h1>端口管理</h1>
        <p>扫描并分析服务器端口占用情况</p>
      </div>
      <div class="header-actions">
        <!-- 可以在这里添加全局操作按钮 -->
      </div>
    </div>
    
    <div class="port-dashboard">
      <!-- 端口扫描卡片 -->
      <div class="scan-section">
        <el-card class="scan-card">
          <template #header>
            <div class="card-title">
              <el-icon><Search /></el-icon>
              <span>快速扫描</span>
            </div>
          </template>
          
          <div class="scan-wrapper">
            <div class="scan-inputs">
              <el-input-number v-model="scanForm.startPort" :min="1" :max="65535" placeholder="起始" controls-position="right" class="port-input"/>
              <span class="range-separator"><el-icon><Right /></el-icon></span>
              <el-input-number v-model="scanForm.endPort" :min="1" :max="65535" placeholder="结束" controls-position="right" class="port-input"/>
            </div>
            
            <el-button type="primary" size="large" @click="handleScan" :loading="scanning" class="scan-btn">
              <el-icon v-if="!scanning"><Aim /></el-icon>
              {{ scanning ? '正在扫描...' : '开始扫描' }}
            </el-button>
          </div>
          
          <div class="scan-status" v-if="scanning">
             <el-progress :percentage="100" status="success" :indeterminate="true" :duration="2" :stroke-width="4" :show-text="false"/>
             <span class="scanning-text">正在扫描端口范围 {{ scanForm.startPort }} - {{ scanForm.endPort }}...</span>
          </div>
          
          <div class="scan-results-area" v-if="scanned">
            <div class="result-header">
              <span class="result-count" v-if="scanResults.length > 0">发现 {{ scanResults.length }} 个开放端口</span>
              <span class="result-empty" v-else>未发现开放端口</span>
            </div>
            
            <div class="result-tags" v-if="scanResults.length > 0">
              <div v-for="port in scanResults" :key="port.port" class="result-tag" @click="checkPort(port.port)">
                <span class="tag-port">{{ port.port }}</span>
                <span class="tag-proto">{{ port.protocol }}</span>
                <el-icon class="arrow"><ArrowRight /></el-icon>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 监听端口列表 -->
      <div class="list-section">
        <el-card class="list-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon><Headset /></el-icon>
                <span>正在监听</span>
                <el-tag size="small" effect="plain" round>{{ listeningPorts.length }}</el-tag>
              </div>
              <el-button circle @click="loadListeningPorts" :loading="loading" :disabled="loading">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </div>
          </template>
          
          <el-table :data="listeningPorts" style="width: 100%" v-loading="loading" :row-style="{ height: '56px' }">
            <el-table-column prop="port" label="端口" width="90" align="center">
              <template #default="{ row }">
                <span class="port-badge">{{ row.port }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="protocol" label="协议" width="80" align="center">
              <template #default="{ row }">
                <el-tag size="small" effect="light" :type="row.protocol === 'tcp' ? '' : 'warning'">
                  {{ row.protocol?.toUpperCase() }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="process" label="进程服务" min-width="140">
              <template #default="{ row }">
                <div class="process-info">
                  <el-icon><Operation /></el-icon>
                  <span>{{ row.process || 'Unknown' }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="pid" label="PID" width="80" align="center">
              <template #default="{ row }">
                <span class="pid-text">{{ row.pid || '-' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="地址绑定" min-width="140">
              <template #default="{ row }">
                <code class="addr-text">{{ row.address }}</code>
              </template>
            </el-table-column>
            <el-table-column width="60" align="center">
              <template #default="{ row }">
                <el-button circle size="small" @click="checkPort(row.port)">
                  <el-icon><ArrowRight /></el-icon>
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>
    </div>
    
    <!-- 端口详情弹窗 -->
    <el-dialog v-model="showCheckDialog" title="端口详情分析" width="600px" center destroy-on-close>
      <div class="detail-container" v-if="portDetail">
        <div class="port-header-banner" :class="{ active: portDetail.inUse }">
          <div class="banner-icon">
            <el-icon><component :is="portDetail.inUse ? 'CircleCheck' : 'CircleClose'" /></el-icon>
          </div>
          <div class="banner-content">
            <div class="port-number">Port {{ checkingPort }}</div>
            <div class="port-status">{{ portDetail.inUse ? '端口正在被使用' : '端口空闲' }}</div>
          </div>
        </div>

        <div class="detail-grid" v-if="portDetail.inUse">
           <div class="detail-box">
             <div class="label">进程名称</div>
             <div class="value process">{{ portDetail.process || 'Unknown' }}</div>
           </div>
           <div class="detail-box">
             <div class="label">进程 PID</div>
             <div class="value pid">{{ portDetail.pid || '-' }}</div>
           </div>
           <div class="detail-box full">
             <div class="label">监听地址</div>
             <div class="value addr">{{ portDetail.address || '0.0.0.0' }}</div>
           </div>
        </div>

        <div class="connections-list" v-if="connections.length">
          <div class="list-title">活跃连接 ({{ connections.length }})</div>
          <el-table :data="connections" size="small" max-height="200px">
            <el-table-column prop="remoteAddr" label="远程地址" min-width="140">
              <template #default="{ row }">
                <code class="conn-addr">{{ row.remoteAddr }}</code>
              </template>
            </el-table-column>
            <el-table-column prop="remotePort" label="端口" width="80" />
            <el-table-column prop="state" label="状态" width="100">
               <template #default="{ row }">
                 <el-tag size="small" effect="plain">{{ row.state }}</el-tag>
               </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { portApi } from '@/api'
import { Search, Right, Aim, ArrowRight, Refresh, Headset, Operation, CircleCheck, CircleClose } from '@element-plus/icons-vue'

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
  if (scanForm.value.startPort > scanForm.value.endPort) {
    ElMessage.warning('起始端口不能大于结束端口')
    return
  }
  scanning.value = true
  scanned.value = false
  scanResults.value = []
  
  try {
    const res = await portApi.scan(scanForm.value)
    scanResults.value = res.data?.ports || []
    scanned.value = true
    if (scanResults.value.length > 0) {
      ElMessage.success(`扫描完成，发现 ${scanResults.value.length} 个开放端口`)
    } else {
      ElMessage.info('扫描完成，未发现开放端口')
    }
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
  showCheckDialog.value = true // 先显示弹窗，内容加载中
  portDetail.value = null
  connections.value = []
  
  try {
    const [checkRes, connRes] = await Promise.all([
      portApi.check(port),
      portApi.connections(port)
    ])
    portDetail.value = checkRes.data || { port, inUse: false }
    connections.value = connRes.data?.connections || []
  } catch (err) { 
    console.error(err) 
    portDetail.value = { port, inUse: false }
  }
}

onMounted(() => { loadListeningPorts() })
</script>

<style scoped>
.port-page {
  animation: fadeIn 0.4s ease-out;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 32px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h1 {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.header-content p {
  color: var(--text-muted);
  font-size: 14px;
}

.port-dashboard {
  display: grid;
  grid-template-columns: 350px 1fr;
  gap: 24px;
}

@media (max-width: 900px) {
  .port-dashboard {
    grid-template-columns: 1fr;
  }
}

/* 扫描卡片 */
.scan-card, .list-card {
  height: 100%;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-card);
}

.scan-card :deep(.el-card__header), .list-card :deep(.el-card__header) {
  border-bottom: 1px solid var(--border-color);
  padding: 16px 20px;
}

.card-title, .header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  color: var(--text-primary);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.scan-wrapper {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 24px;
}

.scan-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.port-input {
  width: 100%;
}

.range-separator {
  color: var(--text-muted);
  flex-shrink: 0;
}

.scan-btn {
  width: 100%;
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(var(--primary-rgb), 0.2);
}

.scan-status {
  margin: 16px 0;
  text-align: center;
}

.scanning-text {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 8px;
  display: block;
}

.result-header {
  margin-bottom: 12px;
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--text-secondary);
}

.result-tags {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.result-tag {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--bg-hover);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.result-tag:hover {
  background: var(--bg-card);
  border-color: var(--primary-color);
  transform: translateX(4px);
}

.tag-port {
  font-weight: 700;
  color: var(--primary-color);
  width: 60px;
}

.tag-proto {
  font-size: 12px;
  background: rgba(var(--primary-rgb), 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  color: var(--primary-color);
  text-transform: uppercase;
}

.arrow {
  color: var(--text-muted);
  font-size: 12px;
}

/* 列表样式 */
.port-badge {
  font-weight: 700;
  font-family: monospace;
  background: var(--bg-hover);
  padding: 4px 8px;
  border-radius: 6px;
  color: var(--text-primary);
}

.process-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary);
}

.process-info .el-icon {
  color: var(--text-muted);
}

.pid-text {
  font-family: monospace;
  color: var(--text-secondary);
}

.addr-text {
  font-family: monospace;
  color: var(--text-secondary);
  font-size: 12px;
}

/* 详情弹窗 */
.port-header-banner {
  background: var(--bg-hover);
  border-radius: var(--radius-lg);
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.port-header-banner.active {
  background: rgba(16, 185, 129, 0.1); /* Success green bg */
}

.banner-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--bg-card);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: var(--text-muted);
}

.port-header-banner.active .banner-icon {
  color: var(--success);
}

.banner-content {
  flex: 1;
}

.port-number {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.port-status {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.detail-box {
  background: var(--bg-hover);
  padding: 16px;
  border-radius: var(--radius-md);
  text-align: center;
}

.detail-box.full {
  grid-column: span 2;
  text-align: left;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-box .label {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.detail-box.full .label { margin-bottom: 0; }

.detail-box .value {
  font-weight: 600;
  color: var(--text-primary);
}

.detail-box .value.process { color: var(--primary-color); }
.detail-box .value.pid { font-family: monospace; }
.detail-box .value.addr { font-family: monospace; }

.list-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  padding-left: 8px;
  border-left: 3px solid var(--primary-color);
}

.conn-addr {
  font-family: monospace;
}
</style>
