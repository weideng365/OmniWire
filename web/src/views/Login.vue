<template>
  <div class="login-page">
    <div class="login-card">
      <div class="logo">
        <el-icon size="48"><Connection /></el-icon>
        <h1>OmniWire</h1>
        <p>WireGuard 服务端管理系统</p>
      </div>
      
      <el-form :model="form" @submit.prevent="handleLogin">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" prefix-icon="User" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" style="width: 100%;" @click="handleLogin" :loading="loading">
            登 录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const form = ref({ username: '', password: '' })

const handleLogin = async () => {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  // 模拟登录
  setTimeout(() => {
    if (form.value.username === 'admin' && form.value.password === 'admin123') {
      localStorage.setItem('token', 'demo-token')
      ElMessage.success('登录成功')
      router.push('/dashboard')
    } else {
      ElMessage.error('用户名或密码错误')
    }
    loading.value = false
  }, 500)
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
}

.login-card {
  width: 400px;
  padding: 48px;
  background: var(--bg-card);
  border-radius: 16px;
  border: 1px solid var(--border-color);
}

.logo {
  text-align: center;
  margin-bottom: 40px;
}

.logo .el-icon {
  color: var(--primary-color);
}

.logo h1 {
  font-size: 28px;
  margin: 16px 0 8px;
  color: var(--text-primary);
}

.logo p {
  color: var(--text-muted);
  font-size: 14px;
}
</style>
