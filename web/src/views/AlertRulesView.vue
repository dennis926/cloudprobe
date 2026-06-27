<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <span>告警规则</span>
          <el-button type="primary" :icon="Plus" @click="showDialog()">新建规则</el-button>
        </div>
      </template>

      <el-table :data="rules" style="width: 100%">
        <el-table-column prop="name" label="规则名称" min-width="150" />
        <el-table-column prop="rule_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ typeMap[row.rule_type] || row.rule_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target_type" label="目标" width="80">
          <template #default="{ row }">
            {{ targetMap[row.target_type] || row.target_type }}
          </template>
        </el-table-column>
        <el-table-column prop="threshold" label="阈值" width="100">
          <template #default="{ row }">
            {{ row.threshold !== null ? row.threshold + '%' : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="持续时间(秒)" width="120" />
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleRule(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteRule(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑规则' : '新建规则'" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="规则名称">
          <el-input v-model="form.name" placeholder="如：CPU使用率告警" />
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="form.rule_type" style="width: 100%">
            <el-option label="离线检测" value="offline" />
            <el-option label="CPU使用率" value="cpu" />
            <el-option label="内存使用率" value="memory" />
            <el-option label="磁盘使用率" value="disk" />
            <el-option label="负载" value="load" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标类型">
          <el-select v-model="form.target_type" style="width: 100%">
            <el-option label="全部服务器" value="all" />
            <el-option label="指定服务器" value="server" />
            <el-option label="服务器分组" value="group" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.target_type !== 'all'" label="目标ID">
          <el-input-number v-model="form.target_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="form.rule_type !== 'offline'" label="阈值(%)">
          <el-input-number v-model="form.threshold" :min="1" :max="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="持续时间(秒)">
          <el-input-number v-model="form.duration" :min="30" :step="30" style="width: 100%" />
        </el-form-item>
        <el-form-item label="通知渠道">
          <el-input v-model="form.channels" placeholder='JSON数组，如：["email","wechat"]' />
        </el-form-item>
        <el-form-item label="升级渠道">
          <el-input v-model="form.upgrade_channels" placeholder='JSON数组，如：["telegram"]' />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const rules = ref<any[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const form = ref<any>({
  name: '',
  rule_type: 'cpu',
  target_type: 'all',
  target_id: null,
  threshold: 80,
  duration: 60,
  channels: '["email"]',
  upgrade_channels: '[]',
  enabled: true
})

const typeMap: Record<string, string> = {
  offline: '离线', cpu: 'CPU', memory: '内存', disk: '磁盘', load: '负载'
}
const targetMap: Record<string, string> = {
  all: '全部', server: '服务器', group: '分组'
}

const loadRules = async () => {
  const res: any = await api.getAlertRules()
  rules.value = res.data || []
}

const showDialog = (row?: any) => {
  if (row) {
    isEdit.value = true
    form.value = { ...row }
  } else {
    isEdit.value = false
    form.value = {
      name: '', rule_type: 'cpu', target_type: 'all', target_id: null,
      threshold: 80, duration: 60, channels: '["email"]', upgrade_channels: '[]', enabled: true
    }
  }
  dialogVisible.value = true
}

const saveRule = async () => {
  try {
    if (isEdit.value) {
      await api.updateAlertRule(form.value.id, form.value)
    } else {
      await api.createAlertRule(form.value)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    loadRules()
  } catch {
    ElMessage.error('保存失败')
  }
}

const toggleRule = async (row: any) => {
  try {
    await api.updateAlertRule(row.id, { enabled: row.enabled })
    ElMessage.success('更新成功')
  } catch {
    ElMessage.error('更新失败')
    row.enabled = !row.enabled
  }
}

const deleteRule = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定删除此规则吗？', '提示', { type: 'warning' })
    await api.deleteAlertRule(id)
    ElMessage.success('删除成功')
    loadRules()
  } catch {
    // cancel
  }
}

onMounted(loadRules)
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
</style>
