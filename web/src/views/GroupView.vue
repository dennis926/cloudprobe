<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <span>分组管理</span>
          <el-button type="primary" :icon="Plus" @click="showDialog()">添加分组</el-button>
        </div>
      </template>

      <el-table :data="groups" style="width: 100%" class="group-table"
        :header-cell-style="{ background: '#1e293b', color: '#94a3b8' }">
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="sort_order" label="排序" width="80" />
        <el-table-column label="服务器数量" width="120">
          <template #default="{ row }">{{ row.server_count || 0 }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" :width="isMobile ? '140' : '200'" fixed="right">
          <template #default="{ row }">
            <el-button link type="warning" @click="showDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteGroup(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑分组' : '添加分组'" :width="isMobile ? '92%' : '480px'">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="分组名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveGroup">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

interface GroupItem {
  id: number
  name: string
  parent_id: number | null
  sort_order: number
  server_count?: number
  created_at: string
  updated_at: string
}

const groups = ref<GroupItem[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const isMobile = ref(false)

const form = ref<any>({
  name: '',
  sort_order: 0,
})

const checkMobile = () => { isMobile.value = window.innerWidth <= 768 }
onMounted(() => { checkMobile(); window.addEventListener('resize', checkMobile) })
onUnmounted(() => window.removeEventListener('resize', checkMobile))

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

const loadGroups = async () => {
  const res: any = await api.getGroups()
  groups.value = res.data || []
}

const loadGroupServerCounts = async () => {
  try {
    const res: any = await api.getServers()
    const servers: any[] = res.data?.list || res.data || []
    const countMap = new Map<number, number>()
    servers.forEach((s: any) => {
      if (s.group_id) {
        countMap.set(s.group_id, (countMap.get(s.group_id) || 0) + 1)
      }
    })
    groups.value.forEach((g: GroupItem) => {
      g.server_count = countMap.get(g.id) || 0
    })
  } catch { /* ignore */ }
}

const showDialog = (row?: GroupItem) => {
  if (row) {
    isEdit.value = true
    form.value = { name: row.name, sort_order: row.sort_order }
  } else {
    isEdit.value = false
    form.value = { name: '', sort_order: 0 }
  }
  dialogVisible.value = true
}

const saveGroup = async () => {
  if (!form.value.name) {
    ElMessage.warning('分组名称不能为空')
    return
  }
  saving.value = true
  try {
    if (isEdit.value) {
      await api.updateGroup(form.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await api.createGroup(form.value)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadGroups().then(loadGroupServerCounts)
  } catch {
    ElMessage.error('操作失败')
  } finally {
    saving.value = false
  }
}

const deleteGroup = async (row: GroupItem) => {
  try {
    await ElMessageBox.confirm(
      row.server_count && row.server_count > 0
        ? `该分组下有 ${row.server_count} 台服务器，无法删除。请先将服务器移出该分组。`
        : '确定删除此分组吗？删除后不可恢复。',
      '提示',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
    await api.deleteGroup(row.id)
    ElMessage.success('删除成功')
    loadGroups().then(loadGroupServerCounts)
  } catch { /* cancel */ }
}

onMounted(() => {
  loadGroups().then(loadGroupServerCounts)
})
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

@media (max-width: 768px) {
  .page-header { flex-wrap: wrap; gap: 8px; }
  .page-header .el-button { width: 100%; }
  .group-table { --el-table-border: none; }
}
</style>
