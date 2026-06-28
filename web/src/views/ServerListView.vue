<template>
  <div class="page-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="page-header">
          <span>服务器管理</span>
          <div class="header-actions">
            <el-select v-model="filterGroup" placeholder="全部分组" clearable size="small" style="width: 140px; margin-right: 12px;">
              <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="g.id" />
            </el-select>
            <el-button type="primary" :icon="Plus" @click="showDialog()">添加节点</el-button>
          </div>
        </div>
      </template>

      <el-table :data="servers" style="width: 100%" class="server-table"
        :header-cell-style="{ background: '#1e293b', color: '#94a3b8' }">
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column label="IP地址" min-width="130">
          <template #default="{ row }">{{ row.ip_public || row.public_ip || '-' }}</template>
        </el-table-column>
        <el-table-column prop="os_type" label="系统" min-width="100" />
        <el-table-column prop="location" label="位置" min-width="100">
          <template #default="{ row }">{{ row.location || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'danger'" size="small">
              {{ row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" :width="isMobile ? '200' : '320'" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="$router.push(`/servers/${row.id}`)">详情</el-button>
            <el-button link type="primary" @click="$router.push(`/ssh/${row.id}`)">SSH</el-button>
            <el-button link type="success" @click="showInstallCmd(row)">安装</el-button>
            <el-button link type="warning" @click="showDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteServer(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑节点' : '添加监控节点'" :width="isMobile ? '92%' : '520px'">
      <!-- 新建模式：简化表单 -->
      <template v-if="!isEdit">
        <el-alert type="info" :closable="false" show-icon style="margin-bottom: 16px;">
          <template #title>创建节点后，系统会自动生成安装命令。Agent 安装后将自动上报服务器信息（CPU、内存、磁盘、IP等）。</template>
        </el-alert>
        <el-form :model="form" label-width="100px">
          <el-form-item label="分组">
            <el-select v-model="form.group_id" placeholder="选择分组（可选）" clearable style="width: 100%">
              <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="g.id" />
            </el-select>
          </el-form-item>
        </el-form>
      </template>
      <!-- 编辑模式：完整表单 -->
      <el-form v-else :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="如：生产服务器A" />
        </el-form-item>
        <el-form-item label="公网IP" required>
          <el-input v-model="form.ip_public" placeholder="如：1.2.3.4" />
        </el-form-item>
        <el-form-item label="内网IP">
          <el-input v-model="form.ip_local" placeholder="可选" />
        </el-form-item>
        <el-form-item label="位置">
          <el-input v-model="form.location" placeholder="如：上海、东京、洛杉矶" />
        </el-form-item>
        <el-form-item label="系统">
          <el-input v-model="form.os_type" placeholder="自动检测或手动填写" />
        </el-form-item>
        <el-form-item label="SSH端口">
          <el-input-number v-model="form.ssh_port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="SSH用户">
          <el-input v-model="form.ssh_user" placeholder="root" />
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="form.group_id" placeholder="选择分组" clearable style="width: 100%">
            <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">账单信息</el-divider>
        <el-form-item label="付费类型">
          <el-select v-model="form.billing_type" placeholder="选择付费类型" clearable style="width: 100%">
            <el-option label="预付费" value="prepaid" />
            <el-option label="后付费" value="postpaid" />
          </el-select>
        </el-form-item>
        <el-form-item label="计费周期">
          <el-select v-model="form.bill_cycle" placeholder="选择计费周期" clearable style="width: 100%">
            <el-option label="月付" value="monthly" />
            <el-option label="季付" value="quarterly" />
            <el-option label="半年付" value="semiannual" />
            <el-option label="年付" value="yearly" />
            <el-option label="两年付" value="biennial" />
            <el-option label="三年付" value="triennial" />
            <el-option label="免费" value="free" />
            <el-option label="按量计费" value="payg" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格">
          <el-input-number v-model="form.bill_price" :min="0" :precision="2" placeholder="0" />
          <span style="margin-left: 8px; color: #94a3b8; font-size: 13px;">元</span>
        </el-form-item>
        <el-form-item label="到期时间">
          <el-date-picker v-model="form.bill_expired_at" type="date" placeholder="选择到期时间" value-format="YYYY-MM-DDTHH:mm:ssZ" style="width: 100%" />
        </el-form-item>
        <el-form-item label="自动续费">
          <el-switch v-model="form.bill_auto_renewal" />
        </el-form-item>

        <el-divider content-position="left">备注</el-divider>
        <el-form-item label="公开备注">
          <el-input v-model="form.public_note" type="textarea" :rows="2" placeholder="所有用户可见" />
        </el-form-item>
        <el-form-item label="私有备注">
          <el-input v-model="form.private_note" type="textarea" :rows="2" placeholder="仅管理员可见" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveServer">{{ isEdit ? '保存' : '创建节点' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { api } from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const servers = ref<any[]>([])
const groups = ref<any[]>([])
const filterGroup = ref<number | undefined>(undefined)
const dialogVisible = ref(false)
const isMobile = ref(false)

const checkMobile = () => { isMobile.value = window.innerWidth <= 768 }
onMounted(() => { checkMobile(); window.addEventListener('resize', checkMobile) })
onUnmounted(() => window.removeEventListener('resize', checkMobile))
const isEdit = ref(false)
const saving = ref(false)
const selectedToken = ref('')

const defaultForm = () => ({
  name: '',
  ip_public: '',
  ip_local: '',
  location: '',
  os_type: '',
  ssh_port: 22,
  ssh_user: 'root',
  group_id: undefined as number | undefined,
  billing_type: '',
  bill_cycle: '',
  bill_price: 0,
  bill_expired_at: '',
  bill_auto_renewal: true,
  public_note: '',
  private_note: '',
})

const form = ref<any>(defaultForm())

const loadGroups = async () => {
  const res: any = await api.getGroups()
  groups.value = res.data || []
}

const loadServers = async () => {
  const params: any = {}
  if (filterGroup.value) {
    params.group_id = filterGroup.value
  }
  const res: any = await api.getServers(params)
  const list = res.data?.list || res.data || []
  servers.value = list
}

const showInstallCmd = async (row: any) => {
  try {
    const res: any = await api.getInstallCommand(row.id)
    await navigator.clipboard.writeText(res.data?.command || '')
    ElMessage.success('安装命令已复制到剪贴板')
  } catch {
    ElMessage.error('获取安装命令失败')
  }
}

const showDialog = (row?: any) => {
  if (row) {
    isEdit.value = true
    form.value = {
      ...row,
      group_id: row.group_id,
      billing_type: row.bill?.billing_type || '',
      bill_cycle: row.bill?.cycle || '',
      bill_price: row.bill?.price || 0,
      bill_expired_at: row.bill?.expired_at || '',
      bill_auto_renewal: row.bill?.auto_renewal ?? true,
      public_note: row.public_note || '',
      private_note: row.private_note || '',
    }
  } else {
    isEdit.value = false
    form.value = defaultForm()
  }
  dialogVisible.value = true
}

const saveServer = async () => {
  if (isEdit.value && !form.value.name) {
    ElMessage.warning('名称为必填')
    return
  }
  saving.value = true
  try {
    if (isEdit.value) {
      await api.updateServer(form.value.id, form.value)
      ElMessage.success('更新成功')
      dialogVisible.value = false
      loadServers()
    } else {
      const res: any = await api.createServer(form.value)
      dialogVisible.value = false
      loadServers()

      const token = res.data?.agent_token
      if (token) {
        selectedToken.value = token
        const cmd = `curl -fsSL ${window.location.origin}/install.sh | bash -s -- "${token}"`

        // 复制到剪贴板（HTTPS环境可用clipboard API，HTTP降级）
        try {
          if (navigator.clipboard && window.isSecureContext) {
            await navigator.clipboard.writeText(cmd)
            ElMessage.success('安装命令已复制到剪贴板')
          }
        } catch { /* ignore */ }

        ElMessageBox.alert(
          `<div style="margin-top: 12px;">
            <p style="margin-bottom: 8px; color: #94a3b8;">在目标服务器上执行以下命令：</p>
            <code id="install-cmd" style="display: block; padding: 12px; background: #0f172a; border-radius: 8px; color: #e2e8f0; font-size: 13px; word-break: break-all; border: 1px solid #1e293b; user-select: all;">${cmd}</code>
            <p style="margin-top: 8px; color: #94a3b8; font-size: 12px;">点击代码框可全选，Ctrl+C 复制。Agent 安装后会自动上报系统信息</p>
          </div>`,
          '安装命令',
          {
            dangerouslyUseHTMLString: true,
            confirmButtonText: '确定',
            showCancelButton: false,
            customStyle: { background: '#1e293b', border: '1px solid #334155', color: '#f1f5f9' },
          }
        )
      } else {
        ElMessage.success('节点创建成功')
      }
    }
  } catch {
    ElMessage.error('操作失败')
  } finally {
    saving.value = false
  }
}

const deleteServer = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定删除此服务器吗？删除后不可恢复。', '提示', { type: 'warning' })
    await api.deleteServer(id)
    ElMessage.success('删除成功')
    loadServers()
  } catch { /* cancel */ }
}

onMounted(loadServers)
onMounted(loadGroups)
watch(filterGroup, () => loadServers())
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
.header-actions {
  display: flex; align-items: center;
}
.page-card :deep(.el-table) { background: transparent; }
.page-card :deep(.el-table__row) { background: transparent; }
.page-card :deep(.el-table td) { border-bottom: 1px solid #1e293b; color: #cbd5e1; }
.page-card :deep(.el-divider__text) {
  background: transparent;
  color: #94a3b8;
  font-size: 13px;
}
.page-card :deep(.el-divider) {
  border-top-color: #1e293b;
}

@media (max-width: 768px) {
  .page-header { flex-wrap: wrap; gap: 8px; }
  .page-header .el-button { width: 100%; }
  .server-table { --el-table-border: none; }
}
</style>
