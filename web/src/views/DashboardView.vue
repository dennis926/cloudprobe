<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: rgba(56, 189, 248, 0.12);">
            <el-icon :size="24" color="#38bdf8"><Server /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">服务器总数</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: rgba(34, 197, 94, 0.12);">
            <el-icon :size="24" color="#22c55e"><CircleCheck /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value" style="color: #22c55e;">{{ stats.online }}</div>
            <div class="stat-label">在线</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: rgba(239, 68, 68, 0.12);">
            <el-icon :size="24" color="#ef4444"><CircleClose /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value" style="color: #ef4444;">{{ stats.offline }}</div>
            <div class="stat-label">离线</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: rgba(245, 158, 11, 0.12);">
            <el-icon :size="24" color="#f59e0b"><Bell /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value" style="color: #f59e0b;">{{ stats.alerts }}</div>
            <div class="stat-label">活跃告警</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 服务器列表 -->
    <el-card class="dashboard-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>服务器状态</span>
          <el-button type="primary" text @click="$router.push('/servers')">
            查看全部
          </el-button>
        </div>
      </template>

      <el-table
        :data="servers"
        style="width: 100%"
        :header-cell-style="{ background: '#1e293b', color: '#94a3b8' }"
      >
        <el-table-column label="名称" min-width="140">
          <template #default="{ row }">
            <div class="server-name">
              <el-tag
                :type="row.status === 'online' ? 'success' : 'danger'"
                size="small"
                effect="dark"
                class="status-tag"
              >
                {{ row.status === 'online' ? '在线' : '离线' }}
              </el-tag>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="IP地址" prop="public_ip" min-width="140" />

        <el-table-column label="系统" min-width="120">
          <template #default="{ row }">
            <span class="os-tag">{{ row.os_type }}</span>
          </template>
        </el-table-column>

        <el-table-column label="位置" min-width="120">
          <template #default="{ row }">
            {{ row.location || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="CPU" min-width="100">
          <template #default>
            <el-progress :percentage="0" :stroke-width="6" color="#38bdf8" />
          </template>
        </el-table-column>

        <el-table-column label="内存" min-width="100">
          <template #default>
            <el-progress :percentage="0" :stroke-width="6" color="#a78bfa" />
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              size="small"
              @click="$router.push(`/ssh/${row.id}`)"
            >
              SSH
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Server, CircleCheck, CircleClose, Bell } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { ElMessage } from 'element-plus'

const stats = ref({ total: 0, online: 0, offline: 0, alerts: 0 })
const servers = ref<any[]>([])

const loadData = async () => {
  try {
    const [serversRes, alertsRes] = await Promise.all([
      api.getServers(),
      api.getAlerts().catch(() => ({ data: [] }))
    ])

    if (serversRes.data) {
      const list = serversRes.data.list || serversRes.data || []
      servers.value = list.slice(0, 10)
      stats.value.total = list.length
      stats.value.online = list.filter((s: any) => s.status === 'online').length
      stats.value.offline = list.filter((s: any) => s.status === 'offline').length
    }

    const alertList = alertsRes.data?.list || alertsRes.data || []
    stats.value.alerts = alertList.filter((a: any) => a.status === 'firing').length
  } catch (error) {
    ElMessage.error('加载数据失败')
  }
}

onMounted(loadData)
</script>

<style scoped>
.dashboard {
  padding-bottom: 40px;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid #1e293b;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: transform 0.2s, border-color 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  border-color: #334155;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #f1f5f9;
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: #64748b;
  margin-top: 6px;
}

.dashboard-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid #1e293b;
  border-radius: 12px;
}

.dashboard-card :deep(.el-card__header) {
  border-bottom: 1px solid #1e293b;
  padding: 16px 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #f1f5f9;
  font-weight: 600;
}

.dashboard-card :deep(.el-card__body) {
  padding: 0;
}

.dashboard-card :deep(.el-table) {
  background: transparent;
}

.dashboard-card :deep(.el-table__row) {
  background: transparent;
}

.dashboard-card :deep(.el-table__row:hover > td) {
  background: rgba(56, 189, 248, 0.04) !important;
}

.dashboard-card :deep(.el-table td) {
  border-bottom: 1px solid #1e293b;
  color: #cbd5e1;
}

.server-name {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #f1f5f9;
}

.status-tag {
  min-width: 44px;
  text-align: center;
}

.os-tag {
  background: rgba(56, 189, 248, 0.1);
  color: #38bdf8;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

@media (max-width: 768px) {
  .stats-row .el-col {
    margin-bottom: 12px;
  }
}
</style>
