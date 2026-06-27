<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <span>系统设置</span>
          <el-button type="primary" @click="saveSettings">保存设置</el-button>
        </div>
      </template>

      <el-tabs v-model="activeTab" class="settings-tabs">
        <el-tab-pane label="服务器" name="server">
          <el-form :model="settings.server" label-width="140px">
            <el-form-item label="运行模式">
              <el-select v-model="settings.server.mode">
                <el-option label="Release" value="release" />
                <el-option label="Debug" value="debug" />
              </el-select>
            </el-form-item>
            <el-form-item label="HTTP端口">
              <el-input-number v-model="settings.server.port" :min="1" :max="65535" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="JWT" name="jwt">
          <el-form :model="settings.jwt" label-width="140px">
            <el-form-item label="Secret">
              <el-input v-model="settings.jwt.secret" show-password />
            </el-form-item>
            <el-form-item label="Access Token过期(小时)">
              <el-input-number v-model="settings.jwt.access_expire" :min="1" :max="168" />
            </el-form-item>
            <el-form-item label="Refresh Token过期(小时)">
              <el-input-number v-model="settings.jwt.refresh_expire" :min="1" :max="720" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="SMTP" name="smtp">
          <el-form :model="settings.smtp" label-width="140px">
            <el-form-item label="SMTP服务器">
              <el-input v-model="settings.smtp.host" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="settings.smtp.port" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item label="用户名">
              <el-input v-model="settings.smtp.user" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="settings.smtp.password" type="password" show-password />
            </el-form-item>
            <el-form-item label="发件人">
              <el-input v-model="settings.smtp.from" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="备份" name="backup">
          <el-form :model="settings.backup" label-width="140px">
            <el-form-item label="启用备份">
              <el-switch v-model="settings.backup.enabled" />
            </el-form-item>
            <el-form-item label="备份邮箱">
              <el-input v-model="settings.backup.email" />
            </el-form-item>
            <el-form-item label="本地保留(天)">
              <el-input-number v-model="settings.backup.keep_local" :min="1" :max="30" />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/request'
import { ElMessage } from 'element-plus'

const activeTab = ref('server')
const settings = ref<any>({
  server: { mode: 'release', port: 8080 },
  jwt: { secret: '', access_expire: 24, refresh_expire: 168 },
  smtp: { host: '', port: 587, user: '', password: '', from: '' },
  backup: { enabled: false, email: '', keep_local: 7 }
})

const loadSettings = async () => {
  try {
    const res: any = await api.getSettings()
    if (res.data) {
      settings.value = {
        server: { mode: res.data.server?.mode || 'release', port: res.data.server?.port || 8080 },
        jwt: {
          secret: res.data.jwt?.secret || '',
          access_expire: res.data.jwt?.access_expire || 24,
          refresh_expire: res.data.jwt?.refresh_expire || 168
        },
        smtp: {
          host: res.data.smtp?.host || '',
          port: res.data.smtp?.port || 587,
          user: res.data.smtp?.user || '',
          password: res.data.smtp?.password || '',
          from: res.data.smtp?.from || ''
        },
        backup: {
          enabled: res.data.backup?.enabled || false,
          email: res.data.backup?.email || '',
          keep_local: res.data.backup?.keep_local || 7
        }
      }
    }
  } catch {
    ElMessage.error('加载设置失败')
  }
}

const saveSettings = async () => {
  try {
    await api.updateSettings(settings.value)
    ElMessage.success('设置已更新（运行时生效，重启后持久化）')
  } catch {
    ElMessage.error('保存失败')
  }
}

onMounted(loadSettings)
</script>

<style scoped>
.page-container { padding-bottom: 40px; }
.page-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid #1e293b;
  border-radius: 12px;
}
.page-card :deep(.el-card__header) {
  border-bottom: 1px solid #1e293b;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #f1f5f9;
  font-weight: 600;
}
.settings-tabs :deep(.el-tabs__item) {
  color: #94a3b8;
}
.settings-tabs :deep(.el-tabs__item.is-active) {
  color: #38bdf8;
}
.settings-tabs :deep(.el-form-item__label) {
  color: #94a3b8;
}
</style>
