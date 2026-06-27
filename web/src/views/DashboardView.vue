<template>
  <div class="dashboard">
    <div class="ws-status" :class="{ active: ws.connected }">
      <span class="ws-dot"></span>
      {{ ws.connected ? '实时连接' : '离线' }}
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :xs="12" :sm="12" :md="6">
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
      <el-col :xs="12" :sm="12" :md="6">
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
      <el-col :xs="12" :sm="12" :md="6">
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
      <el-col :xs="12" :sm="12" :md="6">
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
          <div class="card-actions">
            <el-button type="primary" :icon="Plus" @click="$router.push('/servers')">管理</el-button>
          </div>
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

        <el-table-column label="IP地址" prop="ip_public" min-width="130">
          <template #default="{ row }">
            {{ row.ip_public || row.public_ip || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="系统" min-width="100">
          <template #default="{ row }">
            <span class="os-tag">{{ row.os_type || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="CPU" min-width="120">
          <template #default="{ row }">
            <div class="metric-cell">
              <el-progress
                :percentage="getServerMetric(row.id, 'cpu_percent')"
                :stroke-width="6"
                :color="getMetricColor(getServerMetric(row.id, 'cpu_percent'))"
              />
              <span class="metric-val">{{ getServerMetric(row.id, 'cpu_percent') }}%</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="内存" min-width="120">
          <template #default="{ row }">
            <div class="metric-cell">
              <el-progress
                :percentage="getServerMetric(row.id, 'memory_percent')"
                :stroke-width="6"
                :color="getMetricColor(getServerMetric(row.id, 'memory_percent'))"
              />
              <span class="metric-val">{{ getServerMetric(row.id, 'memory_percent') }}%</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="磁盘" min-width="120">
          <template #default="{ row }">
            <div class="metric-cell">
              <el-progress
                :percentage="getServerMetric(row.id, 'disk_percent')"
                :stroke-width="6"
                :color="getMetricColor(getServerMetric(row.id, 'disk_percent'))"
              />
              <span class="metric-val">{{ getServerMetric(row.id, 'disk_percent') }}%</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="在线" min-width="100">
          <template #default="{ row }">
            <span class="uptime-val">{{ formatUptime(getServerMetric(row.id, 'uptime')) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="$router.push(`/servers/${row.id}`)">详情</el-button>
            <el-button link type="primary" size="small" @click="$router.push(`/ssh/${row.id}`)">SSH</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 最近告警 -->
    <el-card class="dashboard-card alert-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>最近告警</span>
          <el-button type="primary" text @click="$router.push('/alerts')">查看全部</el-button>
        </div>
      </template>
      <el-table :data="recentAlerts" style="width: 100%"
        :header-cell-style="{ background: '#1e293b', color: '#94a3b8' }">
        <el-table-column prop="severity" label="级别" width="80">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : 'warning'" size="small">
              {{ row.severity === 'critical' ? '严重' : '警告' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="内容" min-width="200" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'firing' ? 'danger' : 'success'" size="small">
              {{ row.status === 'firing' ? '触发中' : '已恢复' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="started_at" label="时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.started_at) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Server, CircleCheck, CircleClose, Bell, Plus } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { ElMessage } from 'element-plus'
import { useWebSocket } from '@/composables/useWebSocket'

const ws = useWebSocket()

const stats = ref({ total: 0, online: 0, offline: 0, alerts: 0 })
const servers = ref<any[]>([])
const recentAlerts = ref<any[]>([])

// 定时轮询（兜底 + 首次加载）
let pollTimer: ReturnType<typeof setInterval> | null = null

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
    recentAlerts.value = alertList.slice(0, 5)
    stats.value.alerts = alertList.filter((a: any) => a.status === 'firing').length
  } catch {
    // 静默失败
  }
}

const getServerMetric = (serverId: number, key: string): number => {
  const m = ws.getMetrics(serverId)
  if (m && m[key] !== undefined) return Math.round(Number(m[key]))
  return 0
}

const getMetricColor = (val: number): string => {
  if (val > 90) return '#ef4444'
  if (val > 70) return '#f59e0b'
  return '#22c55e'
}

const formatUptime = (seconds: number): string => {
  if (!seconds) return '-'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  if (d > 0) return `${d}天${h}时`
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}时${m}分`
}

const formatTime = (t: string): string => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(() => {
  loadData()
  ws.connect()
  pollTimer = setInterval(loadData, 60000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  ws.disconnect()
})
</script>

<style scoped>
.dashboard {
  padding-bottom: 40px;
  position: relative;
}

.ws-status {
  position: absolute;
  top: 8px;
  right: 24px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #64748b;
  z-index: 1;
}
.ws-status.active { color: #22c55e; }
.ws-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #64748b;
}
.ws-status.active .ws-dot {
  background: #22c55e;
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.stats-row { margin-bottom: 20px; }
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
  width: 48px; height: 48px;
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
}
.stat-value { font-size: 28px; font-weight: 700; color: #f1f5f9; line-height: 1; }
.stat-label { font-size: 13px; color: #64748b; margin-top: 6px; }

.dashboard-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid #1e293b;
  border-radius: 12px;
  margin-bottom: 20px;
}
.dashboard-card :deep(.el-card__header) {
  border-bottom: 1px solid #1e293b;
  padding: 16px 20px;
}
.card-header {
  display: flex; align-items: center; justify-content: space-between;
  color: #f1f5f9; font-weight: 600;
}
.dashboard-card :deep(.el-card__body) { padding: 0; }
.dashboard-card :deep(.el-table) { background: transparent; }
.dashboard-card :deep(.el-table__row) { background: transparent; }
.dashboard-card :deep(.el-table__row:hover > td) { background: rgba(56, 189, 248, 0.04) !important; }
.dashboard-card :deep(.el-table td) { border-bottom: 1px solid #1e293b; color: #cbd5e1; }

.server-name {
  display: flex; align-items: center; gap: 8px; color: #f1f5f9;
}
.status-tag { min-width: 44px; text-align: center; }
.os-tag {
  background: rgba(56, 189, 248, 0.1); color: #38bdf8;
  padding: 2px 8px; border-radius: 4px; font-size: 12px;
}
.metric-cell { display: flex; align-items: center; gap: 8px; }
.metric-cell :deep(.el-progress) { flex: 1; }
.metric-val {
  font-size: 12px; color: #94a3b8; min-width: 40px; text-align: right;
  font-family: 'Consolas', monospace;
}
.uptime-val { font-size: 13px; color: #94a3b8; }

@media (max-width: 768px) {
  .stats-row .el-col { margin-bottom: 12px; }
  .stat-value { font-size: 22px; }
}
</style>
