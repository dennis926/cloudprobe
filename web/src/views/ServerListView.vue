<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <span>服务器管理</span>
          <el-button type="primary" :icon="Plus" v-if="authStore.isAdmin">添加服务器</el-button>
        </div>
      </template>
      <el-empty v-if="!servers.length" description="暂无服务器数据" />
      <el-table v-else :data="servers" style="width: 100%">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="public_ip" label="IP地址" />
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'danger'">
              {{ row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button link type="primary" @click="$router.push(`/ssh/${row.id}`)">SSH</el-button>
            <el-button link type="primary" @click="$router.push(`/servers/${row.id}`)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const servers = ref<any[]>([])

onMounted(async () => {
  const res: any = await api.getServers()
  servers.value = res.data?.list || res.data || []
})
</script>

<style scoped>
.page-container {
  padding-bottom: 40px;
}
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
