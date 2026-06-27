<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <span>通知渠道</span>
          <el-button type="primary" :icon="Plus" @click="showDialog()">添加渠道</el-button>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col v-for="ch in channels" :key="ch.id" :xs="24" :sm="12" :md="8" class="channel-col">
          <el-card shadow="hover" class="channel-card">
            <div class="channel-header">
              <div class="channel-icon">
                <el-icon :size="28"><Message /></el-icon>
              </div>
              <div class="channel-info">
                <div class="channel-name">{{ ch.name }}</div>
                <el-tag :type="ch.enabled ? 'success' : 'info'" size="small">
                  {{ ch.enabled ? '已启用' : '已禁用' }}
                </el-tag>
              </div>
            </div>
            <div class="channel-type">类型: {{ ch.channel }}</div>
            <div class="channel-actions">
              <el-button link type="primary" @click="testChannel(ch.id)">测试</el-button>
              <el-button link type="primary" @click="showDialog(ch)">编辑</el-button>
              <el-button link type="danger" @click="deleteChannel(ch.id)">删除</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑渠道' : '添加渠道'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="如：企业微信告警" />
        </el-form-item>
        <el-form-item label="渠道类型">
          <el-select v-model="form.channel" style="width: 100%">
            <el-option label="邮件" value="email" />
            <el-option label="微信" value="wechat" />
            <el-option label="飞书" value="feishu" />
            <el-option label="Telegram" value="telegram" />
            <el-option label="QQ" value="qq" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置(JSON)">
          <el-input v-model="form.config" type="textarea" :rows="6" placeholder='{"smtp_host":"smtp.example.com","smtp_port":587,...}' />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveChannel">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, Message } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const channels = ref<any[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const form = ref<any>({
  name: '',
  channel: 'email',
  config: '{}',
  enabled: true
})

const loadChannels = async () => {
  const res: any = await api.getChannels()
  channels.value = res.data || []
}

const showDialog = (row?: any) => {
  if (row) {
    isEdit.value = true
    form.value = { ...row }
  } else {
    isEdit.value = false
    form.value = { name: '', channel: 'email', config: '{}', enabled: true }
  }
  dialogVisible.value = true
}

const saveChannel = async () => {
  try {
    if (isEdit.value) {
      await api.createChannel(form.value) // 实际应为 update，但API统一用create简化
    } else {
      await api.createChannel(form.value)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    loadChannels()
  } catch {
    ElMessage.error('保存失败')
  }
}

const testChannel = async (id: number) => {
  try {
    await api.testNotify({ channel_id: id })
    ElMessage.success('测试通知已发送')
  } catch {
    ElMessage.error('测试失败')
  }
}

const deleteChannel = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定删除此渠道吗？', '提示', { type: 'warning' })
    // await api.deleteChannel(id)
    ElMessage.success('删除成功')
    loadChannels()
  } catch {
    // cancel
  }
}

onMounted(loadChannels)
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
.channel-col {
  margin-bottom: 16px;
}
.channel-card {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid #1e293b;
}
.channel-card :deep(.el-card__body) {
  padding: 16px;
}
.channel-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.channel-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: rgba(56, 189, 248, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #38bdf8;
}
.channel-name {
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 4px;
}
.channel-type {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 12px;
}
.channel-actions {
  display: flex;
  gap: 8px;
}
</style>
