<template>
  <div class="settings-page">
    <div class="page-title">系统设置</div>
    
    <div class="settings-container">
      <div class="settings-nav">
        <div 
          class="nav-item" 
          :class="{ active: activeTab === 'account' }"
          @click="activeTab = 'account'"
        >
          <div class="nav-icon"><el-icon><User /></el-icon></div>
          <div class="nav-text">账号安全</div>
          <el-icon class="nav-arrow"><ArrowRight /></el-icon>
        </div>
        
        <div 
          class="nav-item" 
          :class="{ active: activeTab === 'about' }"
          @click="activeTab = 'about'"
        >
          <div class="nav-icon"><el-icon><InfoFilled /></el-icon></div>
          <div class="nav-text">关于系统</div>
          <el-icon class="nav-arrow"><ArrowRight /></el-icon>
        </div>
      </div>
      
      <div class="settings-content">
        <!-- 账号安全 -->
        <transition name="fade-slide" mode="out-in">
          <div v-if="activeTab === 'account'" class="content-panel" key="account">
            <div class="panel-header">
              <h3>修改密码</h3>
              <p>为了您的账户安全，建议定期更换密码</p>
            </div>
            
            <el-form 
              ref="passwordFormRef"
              :model="settings" 
              label-position="top" 
              class="settings-form"
              size="large"
            >
              <el-form-item label="当前登录用户">
                <el-input v-model="settings.username" disabled prefix-icon="User">
                  <template #suffix>
                    <el-tag size="small" type="success" effect="plain">管理员</el-tag>
                  </template>
                </el-input>
              </el-form-item>
              
              <el-form-item label="旧密码" required>
                <el-input 
                  v-model="settings.oldPassword" 
                  type="password" 
                  placeholder="请输入当前使用的密码" 
                  show-password
                  prefix-icon="Lock"
                />
              </el-form-item>
              
              <el-form-item label="新密码" required>
                <el-input 
                  v-model="settings.newPassword" 
                  type="password" 
                  placeholder="设置一个新的强密码" 
                  show-password
                  prefix-icon="Key"
                />
                 <div class="password-strength">
                    <div class="strength-bar" :style="{ width: passwordStrength + '%', background: strengthColor }"></div>
                 </div>
                 <span class="strength-text">{{ strengthText }}</span>
              </el-form-item>
              
              <el-form-item label="确认新密码" required>
                <el-input 
                  v-model="settings.confirmPassword" 
                  type="password" 
                  placeholder="再次输入新密码" 
                  show-password
                  prefix-icon="CircleCheck"
                />
              </el-form-item>
              
              <div class="form-actions">
                <el-button type="primary" @click="handleSave" :loading="loading" class="save-btn">
                  确认修改
                </el-button>
              </div>
            </el-form>
          </div>
          
          <!-- 关于系统 -->
          <div v-else-if="activeTab === 'about'" class="content-panel" key="about">
            <div class="about-container">
              <div class="app-logo">
                <el-icon><Connection /></el-icon>
              </div>
              <h1 class="app-name">OmniWire</h1>
              <div class="app-version">Version 1.0.0</div>
              
              <div class="feature-list">
                <div class="feature-item">
                  <el-icon><Monitor /></el-icon>
                  <span>WireGuard 核心管理</span>
                </div>
                <div class="feature-item">
                  <el-icon><Switch /></el-icon>
                  <span>高效端口转发</span>
                </div>
                <div class="feature-item">
                  <el-icon><Lock /></el-icon>
                  <span>企业级安全防护</span>
                </div>
              </div>
              
              <div class="app-links">
                <a href="https://github.com" target="_blank" class="link-item">
                  <el-icon><Link /></el-icon>
                  GitHub Repository
                </a>
                <a href="https://wireguard.com" target="_blank" class="link-item">
                  <el-icon><Document /></el-icon>
                  WireGuard Docs
                </a>
              </div>
              
              <div class="copyright">
                &copy; 2024 OmniWire Team. All rights reserved.
              </div>
            </div>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { systemApi } from '@/api/index.js'
import { User, Lock, Key, CircleCheck, InfoFilled, ArrowRight, Connection, Monitor, Switch, Link, Document } from '@element-plus/icons-vue'

const router = useRouter()
const activeTab = ref('account')
const loading = ref(false)

const settings = ref({
  username: 'admin',
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 简单的密码强度计算
const passwordStrength = computed(() => {
  const pwd = settings.value.newPassword
  if (!pwd) return 0
  let score = 0
  if (pwd.length > 6) score += 20
  if (pwd.length > 10) score += 20
  if (/[A-Z]/.test(pwd)) score += 20
  if (/[0-9]/.test(pwd)) score += 20
  if (/[^A-Za-z0-9]/.test(pwd)) score += 20
  return score
})

const strengthColor = computed(() => {
  const s = passwordStrength.value
  if (s < 40) return '#ef4444' // Red
  if (s < 80) return '#f59e0b' // Orange
  return '#10b981' // Green
})

const strengthText = computed(() => {
  const s = passwordStrength.value
  if (s === 0) return ''
  if (s < 40) return '弱'
  if (s < 80) return '中'
  return '强'
})

const handleSave = async () => {
  if (!settings.value.oldPassword) {
    ElMessage.warning('请输入旧密码')
    return
  }
  if (!settings.value.newPassword) {
    ElMessage.warning('请输入新密码')
    return
  }
  if (settings.value.newPassword !== settings.value.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  
  loading.value = true
  try {
    await systemApi.changePassword({
      oldPassword: settings.value.oldPassword,
      newPassword: settings.value.newPassword
    })
    ElMessage.success('密码修改成功，请重新登录')
    setTimeout(() => {
      localStorage.removeItem('token')
      router.push('/login')
    }, 1500)
  } catch (e) {
    // 错误已由响应拦截器处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.settings-page {
  animation: fadeIn 0.4s ease-out;
  max-width: 1000px;
  margin: 0 auto;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 32px;
}

.settings-container {
  display: flex;
  gap: 32px;
  align-items: flex-start;
}

@media (max-width: 800px) {
  .settings-container {
    flex-direction: column;
  }
}

/* 侧边导航 */
.settings-nav {
  width: 280px;
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  overflow: hidden;
  flex-shrink: 0;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  cursor: pointer;
  transition: all 0.2s;
  border-left: 3px solid transparent;
  color: var(--text-secondary);
}

.nav-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.nav-item.active {
  background: rgba(var(--primary-rgb), 0.05);
  color: var(--primary-color);
  border-left-color: var(--primary-color);
}

.nav-icon {
  font-size: 20px;
  margin-right: 12px;
  display: flex;
  align-items: center;
}

.nav-text {
  flex: 1;
  font-weight: 500;
}

.nav-arrow {
  font-size: 14px;
  opacity: 0.5;
}

/* 内容区域 */
.settings-content {
  flex: 1;
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  min-height: 500px;
  width: 100%;
}

.content-panel {
  padding: 40px;
}

.panel-header {
  margin-bottom: 32px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 20px;
}

.panel-header h3 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.panel-header p {
  color: var(--text-secondary);
  font-size: 14px;
}

.settings-form {
  max-width: 480px;
}

.password-strength {
  height: 4px;
  background: var(--bg-hover);
  border-radius: 2px;
  margin-top: 8px;
  width: 100%;
  overflow: hidden;
}

.strength-bar {
  height: 100%;
  transition: all 0.3s ease;
}

.strength-text {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
  display: block;
  text-align: right;
}

.form-actions {
  margin-top: 40px;
}

.save-btn {
  width: 100%;
}

/* 关于页面 */
.about-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  height: 100%;
  padding: 40px 0;
}

.app-logo {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, var(--primary-color), var(--primary-light));
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40px;
  color: white;
  margin-bottom: 24px;
  box-shadow: 0 10px 25px -5px rgba(var(--primary-rgb), 0.4);
}

.app-name {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.app-version {
  font-family: monospace;
  background: var(--bg-hover);
  padding: 4px 12px;
  border-radius: 20px;
  color: var(--text-secondary);
  font-size: 13px;
  margin-bottom: 40px;
}

.feature-list {
  display: flex;
  gap: 24px;
  margin-bottom: 48px;
}

.feature-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: var(--text-secondary);
  font-size: 14px;
}

.feature-item .el-icon {
  font-size: 24px;
  color: var(--primary-color);
  background: rgba(var(--primary-rgb), 0.1);
  padding: 12px;
  border-radius: 12px;
}

.app-links {
  display: flex;
  gap: 20px;
  margin-bottom: 40px;
}

.link-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary);
  text-decoration: none;
  font-weight: 500;
  padding: 10px 20px;
  background: var(--bg-hover);
  border-radius: 8px;
  transition: all 0.2s;
}

.link-item:hover {
  background: var(--bg-card);
  border: 1px solid var(--primary-color);
  color: var(--primary-color);
  transform: translateY(-2px);
}

.copyright {
  font-size: 12px;
  color: var(--text-muted);
}

/* 动画 */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s ease;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
</style>
