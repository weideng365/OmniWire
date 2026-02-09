<template>
  <div class="login-container">
    <div class="login-wrapper">
      <div class="background-shape shape-1"></div>
      <div class="background-shape shape-2"></div>
      
      <div class="login-left">
        <div class="brand">
          <div class="brand-logo">
            <el-icon><Connection /></el-icon>
          </div>
          <h1>OmniWire</h1>
          <p class="subtitle">新一代网络管理系统</p>
        </div>
        <div class="illustration">
           <!-- 简单的几何图形模拟小清新插画 -->
           <div class="circle c1"></div>
           <div class="circle c2"></div>
           <div class="card glass-effect">
             <div class="line l1"></div>
             <div class="line l2"></div>
             <div class="dot"></div>
           </div>
        </div>
        <div class="footer-text">
          &copy; 2024 OmniWire Team
        </div>
      </div>
      
      <div class="login-right">
        <div class="form-container">
          <div class="form-header">
            <h2>欢迎回来</h2>
            <p>请登录您的管理后台</p>
          </div>
          
          <el-form :model="form" @submit.prevent="handleLogin" size="large" class="login-form">
            <el-form-item>
              <el-input 
                v-model="form.username" 
                placeholder="用户名" 
                prefix-icon="User" 
                class="light-input"
              />
            </el-form-item>
            <el-form-item>
              <el-input 
                v-model="form.password" 
                type="password" 
                placeholder="密码" 
                prefix-icon="Lock" 
                show-password 
                class="light-input"
              />
            </el-form-item>
            
            <div class="form-options">
               <el-checkbox v-model="rememberMe">记住我</el-checkbox>
               <a href="#" class="forgot-link">忘记密码？</a>
            </div>

            <el-form-item>
              <el-button type="primary" class="login-btn" @click="handleLogin" :loading="loading">
                登 录
                <el-icon class="el-icon--right"><ArrowRight /></el-icon>
              </el-button>
            </el-form-item>
          </el-form>
           
          <div class="default-creds">
            <p>默认账号: <code>admin</code> / <code>admin123</code></p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { systemApi } from '@/api/index.js'
import { User, Lock, ArrowRight, Connection } from '@element-plus/icons-vue'

const router = useRouter()
const loading = ref(false)
const form = ref({ username: '', password: '' })
const rememberMe = ref(false)

const handleLogin = async () => {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const res = await systemApi.login({ username: form.value.username, password: form.value.password })
    localStorage.setItem('token', res.data?.token || res.token)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (e) {
    // 错误已由响应拦截器处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f7ff; /* 浅蓝背景 */
  position: relative;
  overflow: hidden;
}

/* 背景装饰 */
.background-shape {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  z-index: 0;
}

.shape-1 {
  width: 400px;
  height: 400px;
  background: #bfdbfe; /* Blue-200 */
  top: -100px;
  left: -100px;
  opacity: 0.6;
}

.shape-2 {
  width: 300px;
  height: 300px;
  background: #e9d5ff; /* Purple-200 */
  bottom: -50px;
  right: -50px;
  opacity: 0.6;
}

.login-wrapper {
  display: flex;
  width: 960px;
  height: 580px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(20px);
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.05), 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.5);
  z-index: 10;
}

.login-left {
  flex: 1;
  background: linear-gradient(135deg, #e0f2fe 0%, #f0f9ff 100%);
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 48px;
  overflow: hidden;
}

.brand {
  position: relative;
  z-index: 10;
}

.brand-logo {
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: white;
  margin-bottom: 24px;
  box-shadow: 0 10px 20px rgba(59, 130, 246, 0.2);
}

.brand h1 {
  font-size: 32px;
  font-weight: 800;
  color: #1e293b;
  margin-bottom: 8px;
  letter-spacing: -0.5px;
}

.subtitle {
  color: #64748b;
  font-size: 16px;
  font-weight: 500;
}

/* 插画风格几何图形 */
.illustration {
  flex: 1;
  position: relative;
  width: 100%;
}

.circle {
  position: absolute;
  border-radius: 50%;
}

.c1 {
  width: 180px;
  height: 180px;
  background: linear-gradient(135deg, #dbeafe, #eff6ff);
  top: 10%;
  right: 10%;
  box-shadow: inset 0 0 20px rgba(255,255,255,0.8);
}

.c2 {
  width: 100px;
  height: 100px;
  background: #fdf4ff; /* Pinkish-white */
  bottom: 20%;
  left: 10%;
}

.card.glass-effect {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-6deg);
  width: 220px;
  height: 140px;
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 8px 32px rgba(31, 38, 135, 0.1);
  padding: 24px;
}

.line {
  height: 10px;
  background: #e2e8f0;
  border-radius: 5px;
  margin-bottom: 12px;
}
.l1 { width: 80%; }
.l2 { width: 60%; }
.dot {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #bfdbfe;
  position: absolute;
  bottom: 20px;
  right: 20px;
}

.footer-text {
  font-size: 12px;
  color: #94a3b8;
}

.login-right {
  flex: 1;
  padding: 60px 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
}

.form-container {
  width: 100%;
  max-width: 340px;
}

.form-header {
  margin-bottom: 32px;
}

.form-header h2 {
  font-size: 26px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 8px;
}

.form-header p {
  color: #64748b;
  font-size: 15px;
}

/* 输入框样式覆盖 - 清新风 */
:deep(.light-input .el-input__wrapper) {
  background: #f8fafc !important;
  border: 1px solid #e2e8f0 !important;
  box-shadow: none !important;
  border-radius: 12px !important;
  padding: 4px 12px !important;
  transition: all 0.2s ease;
}

:deep(.light-input .el-input__wrapper.is-focus) {
  background: white !important;
  border-color: #3b82f6 !important;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1) !important;
}

:deep(.light-input .el-input__inner) {
  color: #1e293b !important;
  height: 48px;
  font-size: 15px;
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28px;
  font-size: 14px;
}

:deep(.el-checkbox__label) {
  color: #64748b;
}

.forgot-link {
  color: #3b82f6;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.forgot-link:hover {
  color: #2563eb;
}

.login-btn {
  width: 100%;
  height: 52px;
  font-size: 16px;
  border-radius: 12px;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
  transition: all 0.2s ease;
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.35);
}

.default-creds {
  margin-top: 32px;
  text-align: center;
  font-size: 13px;
  color: #64748b;
  background: #f8fafc;
  padding: 10px;
  border-radius: 12px;
  border: 1px solid #f1f5f9;
}

.default-creds code {
  color: #334155;
  font-family: monospace;
  background: #e2e8f0;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
}

/* Mobile responsive */
@media (max-width: 900px) {
  .login-wrapper {
    flex-direction: column;
    height: auto;
    width: 100%;
    max-width: 440px;
  }
  
  .login-left {
    padding: 32px;
    height: 160px;
    min-height: auto;
  }
  
  .illustration {
    display: none;
  }
  
  .login-right {
    padding: 40px 32px;
  }
}
</style>
