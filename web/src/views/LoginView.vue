<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <div class="logo">
          <el-icon :size="48" color="#38bdf8"><Monitor /></el-icon>
        </div>
        <h1>CloudProbe</h1>
        <p>云服务器探针系统</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="login-form"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>

        <el-button
          type="primary"
          size="large"
          class="login-btn"
          :loading="loading"
          @click="handleLogin"
        >
          登录
        </el-button>
      </el-form>

      <div class="login-footer">
        <p>默认账号: admin / admin</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Monitor } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return

    loading.value = true
    try {
      await authStore.login(form.username, form.password)
      ElMessage.success('登录成功')
      router.push('/')
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '登录失败')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  padding: 20px;
}

.login-box {
  width: 100%;
  max-width: 420px;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 16px;
  padding: 48px 36px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo {
  margin-bottom: 16px;
}

.login-header h1 {
  color: #f1f5f9;
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 8px;
  letter-spacing: -0.5px;
}

.login-header p {
  color: #94a3b8;
  font-size: 14px;
  margin: 0;
}

.login-form :deep(.el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
  box-shadow: 0 0 0 1px #334155 inset;
}

.login-form :deep(.el-input__inner) {
  color: #f1f5f9;
}

.login-form :deep(.el-input__inner::placeholder) {
  color: #64748b;
}

.login-btn {
  width: 100%;
  margin-top: 8px;
  background: linear-gradient(135deg, #38bdf8 0%, #0ea5e9 100%);
  border: none;
  font-weight: 600;
  letter-spacing: 2px;
}

.login-btn:hover {
  background: linear-gradient(135deg, #7dd3fc 0%, #38bdf8 100%);
}

.login-footer {
  text-align: center;
  margin-top: 24px;
}

.login-footer p {
  color: #475569;
  font-size: 12px;
}

@media (max-width: 480px) {
  .login-box {
    padding: 32px 24px;
  }
}
</style>
