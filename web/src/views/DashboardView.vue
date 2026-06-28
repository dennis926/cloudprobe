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
            <el-icon :size="24" color="#38bdf8"><Cpu /></el-icon>
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
            <div class="view-toggle">
              <el-tooltip content="网格视图" placement="top">
                <el-button
                  :type="viewMode === 'grid' ? 'primary' : 'default'"
                  :icon="Grid"
                  size="small"
                  circle
                  @click="viewMode = 'grid'"
                />
              </el-tooltip>
              <el-tooltip content="表格视图" placement="top">
                <el-button
                  :type="viewMode === 'table' ? 'primary' : 'default'"
                  :icon="List"
                  size="small"
                  circle
                  @click="viewMode = 'table'"
                />
              </el-tooltip>
            </div>
            <el-button type="primary" :icon="Plus" @click="$router.push('/servers')">管理</el-button>
          </div>
        </div>
      </template>

      <!-- 网格卡片视图 -->
      <div v-if="viewMode === 'grid'" class="server-grid">
        <div
          v-for="server in servers"
          :key="server.id"
          class="server-card"
          :class="{ 'is-online': server.status === 'online', 'is-offline': server.status === 'offline' }"
        >
          <!-- 卡片头部：名称 + 状态 + IP -->
          <div class="sc-header">
            <div class="sc-title-row">
              <span class="sc-status-dot" :class="server.status === 'online' ? 'dot-online' : 'dot-offline'"></span>
              <span class="sc-name">{{ server.name }}</span>
            </div>
            <div class="sc-ip-row">
              <span class="sc-ip">
                <template v-if="ipVisibility[server.id]">
                  {{ server.ip_public || server.public_ip || '-' }}
                </template>
                <template v-else>***.***.***</template>
              </span>
              <el-icon class="sc-ip-toggle" @click="toggleIP(server.id)">
                <View v-if="ipVisibility[server.id]" />
                <Hide v-else />
              </el-icon>
            </div>
            <div class="sc-location" v-if="server.location">
              {{ server.location }}
            </div>
          </div>

          <!-- 资源使用率 -->
          <div class="sc-metrics">
            <div class="sc-metric">
              <div class="sc-metric-label">CPU</div>
              <div class="sc-metric-bar">
                <div class="sc-bar-track">
                  <div
                    class="sc-bar-fill"
                    :style="{ width: getServerMetric(server.id, 'cpu_percent') + '%', background: getBarColor(getServerMetric(server.id, 'cpu_percent')) }"
                  ></div>
                </div>
                <span class="sc-metric-val" :style="{ color: getBarColor(getServerMetric(server.id, 'cpu_percent')) }">
                  {{ getServerMetric(server.id, 'cpu_percent') }}%
                </span>
              </div>
            </div>
            <div class="sc-metric">
              <div class="sc-metric-label">内存</div>
              <div class="sc-metric-bar">
                <div class="sc-bar-track">
                  <div
                    class="sc-bar-fill"
                    :style="{ width: getServerMetric(server.id, 'memory_percent') + '%', background: getBarColor(getServerMetric(server.id, 'memory_percent')) }"
                  ></div>
                </div>
                <span class="sc-metric-val" :style="{ color: getBarColor(getServerMetric(server.id, 'memory_percent')) }">
                  {{ getServerMetric(server.id, 'memory_percent') }}%
                </span>
              </div>
            </div>
            <div class="sc-metric">
              <div class="sc-metric-label">磁盘</div>
              <div class="sc-metric-bar">
                <div class="sc-bar-track">
                  <div
                    class="sc-bar-fill"
                    :style="{ width: getServerMetric(server.id, 'disk_percent') + '%', background: getBarColor(getServerMetric(server.id, 'disk_percent')) }"
                  ></div>
                </div>
                <span class="sc-metric-val" :style="{ color: getBarColor(getServerMetric(server.id, 'disk_percent')) }">
                  {{ getServerMetric(server.id, 'disk_percent') }}%
                </span>
              </div>
            </div>
          </div>

          <!-- 流量信息 -->
          <div class="sc-traffic">
            <span class="sc-traffic-item">
              <span class="sc-traffic-arrow upload">&#x2191;</span>
              {{ getTrafficRate(server.id, 'upload') }}
            </span>
            <span class="sc-traffic-item">
              <span class="sc-traffic-arrow download">&#x2193;</span>
              {{ getTrafficRate(server.id, 'download') }}
            </span>
          </div>

          <!-- 到期时间 -->
          <div v-if="getExpiryInfo(server)" class="sc-expiry">
            <span class="sc-expiry-text" :style="{ color: getExpiryInfo(server)!.color }">
              {{ getExpiryInfo(server)!.text }}
            </span>
          </div>

          <!-- 卡片底部操作 -->
          <div class="sc-footer">
            <el-button link type="primary" size="small" @click="$router.push(`/servers/${server.id}`)">详情</el-button>
            <el-button link type="primary" size="small" @click="$router.push(`/ssh/${server.id}`)">SSH</el-button>
          </div>
        </div>
      </div>

      <!-- 表格视图 -->
      <el-table
        v-if="viewMode === 'table'"
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
import { ref, reactive, watch, onMounted, onUnmounted } from 'vue'
import { Cpu, CircleCheck, CircleClose, Bell, Plus, Grid, List, View, Hide } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { useWebSocket } from '@/composables/useWebSocket'

const ws = useWebSocket()

const stats = ref({ total: 0, online: 0, offline: 0, alerts: 0 })
const servers = ref<any[]>([])
const recentAlerts = ref<any[]>([])

// 视图模式：grid（网格卡片）/ table（表格）
const viewMode = ref<string>(localStorage.getItem('dashboard-view-mode') || 'grid')
watch(viewMode, (val) => localStorage.setItem('dashboard-view-mode', val))

// IP 可见性控制
const ipVisibility = reactive<Record<number, boolean>>({})
const toggleIP = (id: number) => {
  ipVisibility[id] = !ipVisibility[id]
}

// 进度条颜色（蓝 <70% / 橙 <90% / 红 >=90%）
const getBarColor = (val: number): string => {
  if (val >= 90) return '#ef4444'
  if (val >= 70) return '#f59e0b'
  return '#3b82f6'
}

// 格式化流量
const formatBytes = (bytes: number): string => {
  if (!bytes || bytes <= 0) return '0 B/s'
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  let i = 0
  let val = bytes
  while (val >= 1024 && i < units.length - 1) { val /= 1024; i++ }
  return `${val.toFixed(1)} ${units[i]}`
}

// 到期时间倒计时
const getExpiryInfo = (server: any): { text: string; color: string } | null => {
  if (!server.bill || !server.bill.expired_at) return null
  const now = Date.now()
  const expiredAt = new Date(server.bill.expired_at).getTime()
  const diff = expiredAt - now
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days <= 0) return { text: '已过期', color: '#ef4444' }
  if (days <= 30) return { text: `${days}天后到期`, color: '#ef4444' }
  if (days <= 90) return { text: `${days}天后到期`, color: '#f59e0b' }
  return { text: `${days}天后到期`, color: '#64748b' }
}

// 获取上行/下行速率
const getTrafficRate = (serverId: number, direction: 'upload' | 'download'): string => {
  const m = ws.getMetrics(serverId)
  if (!m) return '0 B/s'
  const key = direction === 'upload' ? 'net_upload' : 'net_download'
  return formatBytes(Number(m[key]) || 0)
}

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
.dashboard-card :deep(.el-card__body) .server-grid { padding: 16px; }
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

/* ========== 视图切换按钮 ========== */
.view-toggle {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-right: 8px;
}

/* ========== 网格卡片布局 ========== */
.server-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 16px;
}

@media (max-width: 1200px) {
  .server-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .server-grid { grid-template-columns: 1fr; }
}

/* ========== 服务器卡片 ========== */
.server-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid #1e293b;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  overflow: hidden;
  transition: transform 0.2s, border-color 0.2s, box-shadow 0.2s;
}
.server-card:hover {
  transform: translateY(-2px);
  border-color: #334155;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}
.server-card.is-online {
  border-left: 3px solid #22c55e;
}
.server-card.is-offline {
  border-left: 3px solid #ef4444;
}

/* 卡片头部 */
.sc-header {
  margin-bottom: 12px;
}
.sc-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.sc-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.sc-status-dot.dot-online {
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.5);
}
.sc-status-dot.dot-offline {
  background: #ef4444;
}
.sc-name {
  font-size: 15px;
  font-weight: 600;
  color: #f1f5f9;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sc-ip-row {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #64748b;
  margin-bottom: 2px;
}
.sc-ip {
  font-family: 'Consolas', monospace;
}
.sc-ip-toggle {
  cursor: pointer;
  color: #64748b;
  transition: color 0.2s;
}
.sc-ip-toggle:hover {
  color: #94a3b8;
}
.sc-location {
  font-size: 12px;
  color: #64748b;
  margin-top: 2px;
}

/* 资源使用率 */
.sc-metrics {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}
.sc-metric-label {
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 3px;
}
.sc-metric-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}
.sc-bar-track {
  flex: 1;
  height: 6px;
  background: rgba(51, 65, 85, 0.6);
  border-radius: 3px;
  overflow: hidden;
}
.sc-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.5s ease;
  min-width: 0;
}
.sc-metric-val {
  font-size: 12px;
  font-family: 'Consolas', monospace;
  min-width: 38px;
  text-align: right;
  font-weight: 600;
}

/* 流量信息 */
.sc-traffic {
  display: flex;
  gap: 16px;
  padding: 8px 0;
  border-top: 1px solid #1e293b;
  border-bottom: 1px solid #1e293b;
  margin-bottom: 8px;
}
.sc-traffic-item {
  font-size: 12px;
  color: #94a3b8;
  font-family: 'Consolas', monospace;
  display: flex;
  align-items: center;
  gap: 4px;
}
.sc-traffic-arrow {
  font-size: 13px;
}
.sc-traffic-arrow.upload {
  color: #22c55e;
}
.sc-traffic-arrow.download {
  color: #3b82f6;
}

/* 到期时间 */
.sc-expiry {
  margin-bottom: 8px;
}
.sc-expiry-text {
  font-size: 12px;
  font-weight: 500;
}

/* 卡片底部 */
.sc-footer {
  display: flex;
  gap: 4px;
  justify-content: flex-end;
}
</style>
