<template>
  <div class="settings-page">
    <el-card>
      <template #header><span>系统设置</span></template>
      <el-form label-width="120px">
        <el-form-item label="管理员用户名">
          <el-input v-model="settings.username" disabled />
        </el-form-item>
        <el-form-item label="修改密码">
          <el-input v-model="settings.newPassword" type="password" placeholder="输入新密码" show-password />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="settings.confirmPassword" type="password" placeholder="确认新密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSave">保存设置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    
    <el-card style="margin-top: 24px;">
      <template #header><span>关于</span></template>
      <div class="about-info">
        <p><strong>OmniWire</strong> - WireGuard 服务端 & 端口转发管理系统</p>
        <p>版本: 1.0.0</p>
        <p>技术栈: GoFrame + Vue 3 + Element Plus</p>
        <p>
          <a href="https://github.com" target="_blank">GitHub</a> |
          <a href="https://wireguard.com" target="_blank">WireGuard</a>
        </p>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const settings = ref({
  username: 'admin',
  newPassword: '',
  confirmPassword: ''
})

const handleSave = () => {
  if (settings.value.newPassword && settings.value.newPassword !== settings.value.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  ElMessage.success('设置已保存')
}
</script>

<style scoped>
.settings-page { animation: fadeIn 0.3s ease-out; max-width: 600px; }
.about-info p { margin: 12px 0; color: var(--text-secondary); }
.about-info a { color: var(--primary-color); text-decoration: none; }
.about-info a:hover { text-decoration: underline; }
</style>
