<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <div class="server-title">
            <el-button link :icon="ArrowLeft" @click="$router.back()">返回</el-button>
            <span>{{ server?.name || '服务器详情' }}</span>
            <el-tag :type="server?.status === 'online' ? 'success' : 'danger'" size="small">
              {{ server?.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </div>
          <el-button type="primary" @click="$router.push(`/ssh/${serverId}`)">WebSSH</el-button>
        </div>
      </template>

      <el-row :gutter="16" class="info-row">
        <el-col :xs="24" :md="12">
          <div class="info-section">
            <h4>基本信息</h4>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="IP地址">{{ server?.public_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="系统">{{ server?.os_type || '-' }}</el-descriptions-item>
              <el-descriptions-item label="位置">{{ server?.location || '-' }}</el-descriptions-item>
              <el-descriptions-item label="CPU">{{ server?.cpu_info || '-' }}</el-descriptions-item>
              <el-descriptions-item label="内存">{{ formatBytes(server?.memory_total) }}</el-descriptions-item>
              <el-descriptions-item label="磁盘">{{ formatBytes(server?.disk_total) }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-col>
        <el-col :xs="24" :md="12">
          <div class="info-section">
            <h4>实时指标</h4>
            <div class="metric-gauges">
              <div class="gauge-item">
                <div class="gauge-label">CPU</div>
                <el-progress :percentage="Math.round(metrics.cpu_percent || 0)" :stroke-width="10" color="#38bdf8" />
              </div>
              <div class="gauge-item">
                <div class="gauge-label">内存</div>
                <el-progress :percentage="Math.round(metrics.memory_percent || 0)" :stroke-width="10" color="#a78bfa" />
              </div>
              <div class="gauge-item">
                <div class="gauge-label">磁盘</div>
                <el-progress :percentage="Math.round(metrics.disk_percent || 0)" :stroke-width="10" color="#22c55e" />
              </div>
              <div class="gauge-item">
                <div class="gauge-label">负载</div>
                <div class="load-value">{{ metrics.load1?.toFixed(2) || '0.00' }}</div>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>

      <!-- 图表区域 -->
      <div class="chart-section">
        <h4>CPU & 内存趋势（最近24小时）</h4>
        <div ref="chartRef" class="chart-container" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'

const route = useRoute()
const serverId = Number(route.params.id)
const server = ref<any>(null)
const metrics = ref<any>({})
const chartRef = ref<HTMLDivElement>()
let chart: echarts.ECharts | null = null

const loadServer = async () => {
  try {
    const res: any = await api.getServer(serverId)
    server.value = res.data
  } catch {
    ElMessage.error('加载服务器信息失败')
  }
}

const loadMetrics = async () => {
  try {
    const res: any = await api.getMetrics(serverId)
    const data = res.data || []
    updateChart(data)
  } catch {
    // ignore
  }
}

const updateChart = (data: any[]) => {
  if (!chart || !data.length) return

  const times = data.map((d: any) => {
    const t = new Date(d.time)
    return `${t.getHours().toString().padStart(2, '0')}:${t.getMinutes().toString().padStart(2, '0')}`
  }).reverse()
  const cpuData = data.map((d: any) => d.cpu_percent || 0).reverse()
  const memData = data.map((d: any) => d.memory_percent || 0).reverse()

  chart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis' },
    legend: { data: ['CPU', '内存'], textStyle: { color: '#94a3b8' } },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLine: { lineStyle: { color: '#334155' } },
      axisLabel: { color: '#64748b' }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { lineStyle: { color: '#334155' } },
      axisLabel: { color: '#64748b', formatter: '{value}%' },
      splitLine: { lineStyle: { color: '#1e293b' } }
    },
    series: [
      {
        name: 'CPU',
        type: 'line',
        smooth: true,
        data: cpuData,
        itemStyle: { color: '#38bdf8' },
        areaStyle: { color: 'rgba(56,189,248,0.1)' }
      },
      {
        name: '内存',
        type: 'line',
        smooth: true,
        data: memData,
        itemStyle: { color: '#a78bfa' },
        areaStyle: { color: 'rgba(167,139,250,0.1)' }
      }
    ]
  })
}

const formatBytes = (bytes?: number) => {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(2)} ${units[i]}`
}

onMounted(async () => {
  await loadServer()
  await nextTick()
  if (chartRef.value) {
    chart = echarts.init(chartRef.value)
    loadMetrics()
  }
})

onUnmounted(() => {
  chart?.dispose()
})
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
.server-title {
  display: flex;
  align-items: center;
  gap: 12px;
}
.info-row {
  margin-bottom: 20px;
}
.info-section h4 {
  color: #94a3b8;
  font-size: 14px;
  margin-bottom: 12px;
}
.info-section :deep(.el-descriptions__body) {
  background: transparent;
}
.info-section :deep(.el-descriptions__label) {
  background: #1e293b;
  color: #94a3b8;
}
.info-section :deep(.el-descriptions__content) {
  background: rgba(15, 23, 42, 0.4);
  color: #e2e8f0;
}
.metric-gauges {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.gauge-label {
  font-size: 13px;
  color: #94a3b8;
  margin-bottom: 6px;
}
.load-value {
  font-size: 24px;
  font-weight: 700;
  color: #f59e0b;
}
.chart-section {
  margin-top: 20px;
}
.chart-section h4 {
  color: #94a3b8;
  font-size: 14px;
  margin-bottom: 12px;
}
.chart-container {
  width: 100%;
  height: 300px;
  background: rgba(15, 23, 42, 0.4);
  border-radius: 8px;
}
</style>
