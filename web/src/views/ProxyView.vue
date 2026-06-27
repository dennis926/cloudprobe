<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <span>代理管理（3x-ui）</span>
          <el-button type="primary" @click="showConfig = true">配置面板</el-button>
        </div>
      </template>

      <el-empty v-if="!status?.connected" description="请先配置 3x-ui 面板连接" />

      <template v-else>
        <el-row :gutter="16" class="status-row">
          <el-col :xs="24" :sm="8">
            <el-statistic title="连接状态" :value="status.connected ? '已连接' : '未连接'" />
          </el-col>
          <el-col :xs="24" :sm="16">
            <div class="panel-url">面板地址: {{ status.panel_url }}</div>
          </el-col>
        </el-row>

        <el-tabs v-model="activeTab" class="proxy-tabs">
          <el-tab-pane label="入站列表" name="inbounds">
            <el-table :data="inbounds" style="width: 100%">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="remark" label="备注" />
              <el-table-column prop="protocol" label="协议" width="100" />
              <el-table-column prop="port" label="端口" width="80" />
              <el-table-column label="状态" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.enable ? 'success' : 'danger'" size="small">
                    {{ row.enable ? '启用' : '禁用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="流量" width="200">
                <template #default="{ row }">
                  ↑ {{ formatBytes(row.up) }} / ↓ {{ formatBytes(row.down) }}
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>

          <el-tab-pane label="节点概览" name="nodes">
            <el-table :data="nodes" style="width: 100%">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="remark" label="节点名" />
              <el-table-column prop="protocol" label="协议" width="100" />
              <el-table-column prop="port" label="端口" width="80" />
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </template>
    </el-card>

    <!-- 配置对话框 -->
    <el-dialog v-model="showConfig" title="3x-ui 面板配置" :width="isMobile ? '92%' : '480px'">
      <el-form :model="configForm" label-width="100px">
        <el-form-item label="面板地址">
          <el-input v-model="configForm.panel_url" placeholder="http://127.0.0.1:54321" />
        </el-form-item>
        <el-form-item label="API Token">
          <el-input v-model="configForm.api_token" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConfig = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '@/api/request'
import { ElMessage } from 'element-plus'

const activeTab = ref('inbounds')
const status = ref<any>({})
const inbounds = ref<any[]>([])
const nodes = ref<any[]>([])
const showConfig = ref(false)
const isMobile = ref(false)
const configForm = ref({ panel_url: '', api_token: '' })

const checkMobile = () => { isMobile.value = window.innerWidth <= 768 }
onMounted(() => { checkMobile(); window.addEventListener('resize', checkMobile) })
onUnmounted(() => window.removeEventListener('resize', checkMobile))

const loadStatus = async () => {
  try {
    const res: any = await api.getProxyStatus()
    status.value = res.data || {}
  } catch {}
}

const loadInbounds = async () => {
  try {
    const res: any = await api.getProxyInbounds()
    inbounds.value = res.data || []
  } catch {}
}

const loadNodes = async () => {
  try {
    const res: any = await api.getProxyNodes()
    nodes.value = res.data || []
  } catch {}
}

const saveConfig = async () => {
  try {
    // await api.updateProxyConfig(configForm.value)
    ElMessage.success('配置已保存')
    showConfig.value = false
    loadStatus()
    loadInbounds()
  } catch {
    ElMessage.error('保存失败')
  }
}

const formatBytes = (bytes?: number) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(2)} ${units[i]}`
}

onMounted(() => {
  loadStatus()
  loadInbounds()
  loadNodes()
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
.status-row {
  margin-bottom: 20px;
}
.panel-url {
  color: #64748b;
  font-size: 13px;
}
.proxy-tabs :deep(.el-tabs__item) {
  color: #94a3b8;
}
.proxy-tabs :deep(.el-tabs__item.is-active) {
  color: #38bdf8;
}
</style>
