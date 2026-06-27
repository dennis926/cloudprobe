<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <span>告警中心</span>
          <div class="header-actions">
            <el-radio-group v-model="filter" size="small">
              <el-radio-button value="all">全部</el-radio-button>
              <el-radio-button value="firing">触发中</el-radio-button>
              <el-radio-button value="resolved">已恢复</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <el-table :data="filteredAlerts" style="width: 100%"
        :header-cell-style="{ background: '#1e293b', color: '#94a3b8' }">
        <el-table-column prop="severity" label="级别" width="80">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : 'warning'" size="small" effect="dark">
              {{ row.severity === 'critical' ? '严重' : '警告' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="告警内容" min-width="300" />
        <el-table-column prop="server_name" label="服务器" width="140">
          <template #default="{ row }">{{ row.server_name || `Server #${row.server_id}` }}</template>
        </el-table-column>
        <el-table-column prop="duration_sec" label="持续时间" width="100">
          <template #default="{ row }">
            {{ formatDuration(row.duration_sec) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'firing' ? 'danger' : 'success'" size="small">
              {{ row.status === 'firing' ? '触发中' : '已恢复' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="started_at" label="触发时间" width="160">
          <template #default="{ row }">{{ formatTime(row.started_at) }}</template>
        </el-table-column>
        <el-table-column prop="resolved_at" label="恢复时间" width="160">
          <template #default="{ row }">{{ row.resolved_at ? formatTime(row.resolved_at) : '-' }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="page-card" style="margin-top: 20px;">
      <template #header>
        <div class="page-header">
          <span>告警规则</span>
          <el-button type="primary" :icon="Plus" @click="showRuleDialog()">新建规则</el-button>
        </div>
      </template>
      <el-table :data="rules" style="width: 100%"
        :header-cell-style="{ background: '#1e293b', color: '#94a3b8' }">
        <el-table-column prop="name" label="规则名称" min-width="150" />
        <el-table-column prop="rule_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ typeMap[row.rule_type] || row.rule_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target_type" label="目标" width="80">
          <template #default="{ row }">{{ targetMap[row.target_type] || row.target_type }}</template>
        </el-table-column>
        <el-table-column prop="threshold" label="阈值" width="80">
          <template #default="{ row }">{{ row.threshold ? row.threshold + '%' : '-' }}</template>
        </el-table-column>
        <el-table-column prop="enabled" label="启用" width="70">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" size="small" @change="toggleRule(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showRuleDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteRule(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 规则编辑对话框 -->
    <el-dialog v-model="ruleDialogVisible" :title="ruleEdit ? '编辑规则' : '新建规则'" :width="isMobile ? '92%' : '560px'">
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="规则名称">
          <el-input v-model="ruleForm.name" placeholder="如：CPU使用率告警" />
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="ruleForm.rule_type" style="width: 100%">
            <el-option label="离线检测" value="offline" />
            <el-option label="CPU使用率" value="cpu" />
            <el-option label="内存使用率" value="memory" />
            <el-option label="磁盘使用率" value="disk" />
            <el-option label="系统负载" value="load" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标">
          <el-select v-model="ruleForm.target_type" style="width: 100%">
            <el-option label="全部服务器" value="all" />
            <el-option label="指定服务器" value="server" />
            <el-option label="服务器分组" value="group" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="ruleForm.target_type !== 'all'" label="目标ID">
          <el-input-number v-model="ruleForm.target_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="ruleForm.rule_type !== 'offline'" label="阈值(%)">
          <el-input-number v-model="ruleForm.threshold" :min="1" :max="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="持续时间(秒)">
          <el-input-number v-model="ruleForm.duration" :min="30" :step="30" style="width: 100%" />
        </el-form-item>
        <el-form-item label="通知渠道">
          <el-input v-model="ruleForm.channels" placeholder='如：["email","wechat"]' />
        </el-form-item>
        <el-form-item label="升级渠道">
          <el-input v-model="ruleForm.upgrade_channels" placeholder='如：["telegram"]' />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const filter = ref('all')
const alerts = ref<any[]>([])
const rules = ref<any[]>([])
const ruleDialogVisible = ref(false)
const ruleEdit = ref(false)
const isMobile = ref(false)
const checkMobile = () => { isMobile.value = window.innerWidth <= 768 }
onMounted(() => { checkMobile(); window.addEventListener('resize', checkMobile) })
onUnmounted(() => window.removeEventListener('resize', checkMobile))
const ruleForm = ref<any>({
  name: '', rule_type: 'cpu', target_type: 'all', target_id: null,
  threshold: 80, duration: 60, channels: '["email"]', upgrade_channels: '[]', enabled: true
})

const typeMap: Record<string, string> = { offline: '离线', cpu: 'CPU', memory: '内存', disk: '磁盘', load: '负载' }
const targetMap: Record<string, string> = { all: '全部', server: '服务器', group: '分组' }

const filteredAlerts = computed(() => {
  if (filter.value === 'all') return alerts.value
  return alerts.value.filter((a: any) => a.status === filter.value)
})

const formatDuration = (sec: number) => {
  if (!sec) return '-'
  if (sec < 60) return `${sec}秒`
  if (sec < 3600) return `${Math.floor(sec/60)}分${sec%60}秒`
  return `${Math.floor(sec/3600)}时${Math.floor(sec%3600/60)}分`
}

const formatTime = (t: string) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

const loadAlerts = async () => {
  try {
    const res: any = await api.getAlerts()
    alerts.value = res.data?.list || res.data || []
  } catch {}
}

const loadRules = async () => {
  try {
    const res: any = await api.getAlertRules()
    rules.value = res.data || []
  } catch {}
}

const showRuleDialog = (row?: any) => {
  if (row) {
    ruleEdit.value = true
    ruleForm.value = { ...row }
  } else {
    ruleEdit.value = false
    ruleForm.value = {
      name: '', rule_type: 'cpu', target_type: 'all', target_id: null,
      threshold: 80, duration: 60, channels: '["email"]', upgrade_channels: '[]', enabled: true
    }
  }
  ruleDialogVisible.value = true
}

const saveRule = async () => {
  try {
    if (ruleEdit.value) {
      await api.updateAlertRule(ruleForm.value.id, ruleForm.value)
    } else {
      await api.createAlertRule(ruleForm.value)
    }
    ElMessage.success('保存成功')
    ruleDialogVisible.value = false
    loadRules()
  } catch { ElMessage.error('保存失败') }
}

const toggleRule = async (row: any) => {
  try {
    await api.updateAlertRule(row.id, { enabled: row.enabled })
  } catch { row.enabled = !row.enabled; }
}

const deleteRule = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定删除？', '提示', { type: 'warning' })
    await api.deleteAlertRule(id)
    ElMessage.success('已删除')
    loadRules()
  } catch {}
}

onMounted(() => { loadAlerts(); loadRules() })
</script>

<style scoped>
.page-container { padding-bottom: 40px; }
.page-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid #1e293b;
  border-radius: 12px;
}
.page-card :deep(.el-card__header) { border-bottom: 1px solid #1e293b; }
.page-header {
  display: flex; align-items: center; justify-content: space-between;
  color: #f1f5f9; font-weight: 600;
}
.page-card :deep(.el-table) { background: transparent; }
.page-card :deep(.el-table__row) { background: transparent; }
.page-card :deep(.el-table td) { border-bottom: 1px solid #1e293b; color: #cbd5e1; }
.header-actions :deep(.el-radio-group) { --el-radio-button-checked-bg-color: #38bdf8; }

@media (max-width: 768px) {
  .page-header { flex-wrap: wrap; gap: 8px; }
  .header-actions { width: 100%; overflow-x: auto; }
  .page-header .el-button { width: 100%; }
}
</style>
