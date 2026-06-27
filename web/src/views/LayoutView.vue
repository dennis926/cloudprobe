<template>
  <el-container class="layout-container">
    <!-- 移动端遮罩 -->
    <div v-if="isMobile && !isCollapse" class="sidebar-overlay" @click="toggleCollapse"></div>

    <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar" :class="{ 'sidebar-mobile': isMobile, 'sidebar-open': isMobile && !isCollapse }">
      <div class="sidebar-header">
        <el-icon :size="28" color="#38bdf8"><Monitor /></el-icon>
        <span v-show="!isCollapse" class="sidebar-title">CloudProbe</span>
      </div>

      <el-menu
        :default-active="route.path"
        :collapse="isCollapse"
        :collapse-transition="false"
        router
        class="sidebar-menu"
        background-color="transparent"
        text-color="#94a3b8"
        active-text-color="#38bdf8"
        @select="onMenuSelect"
      >
        <el-menu-item index="/">
          <el-icon><Odometer /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>

        <el-menu-item index="/servers">
          <el-icon><Server /></el-icon>
          <template #title>服务器</template>
        </el-menu-item>

        <el-menu-item index="/alerts">
          <el-icon><Bell /></el-icon>
          <template #title>告警规则</template>
        </el-menu-item>

        <el-menu-item index="/notifications">
          <el-icon><Message /></el-icon>
          <template #title>通知渠道</template>
        </el-menu-item>

        <el-menu-item index="/proxy">
          <el-icon><Switch /></el-icon>
          <template #title>代理管理</template>
        </el-menu-item>

        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <template #title>系统设置</template>
        </el-menu-item>
      </el-menu>

      <div v-if="!isMobile" class="sidebar-footer">
        <el-button
          type="text"
          :icon="isCollapse ? Expand : Fold"
          class="collapse-btn"
          @click="toggleCollapse"
        />
      </div>
    </el-aside>

    <el-container>
      <el-header class="main-header" :height="isMobile ? '52px' : '60px'">
        <div class="header-left">
          <!-- 移动端汉堡菜单 -->
          <el-button v-if="isMobile" type="text" :icon="Expand" class="mobile-menu-btn" @click="toggleCollapse" />
          <breadcrumb v-if="!isMobile" />
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="isMobile ? 28 : 32" :icon="UserFilled" />
              <span v-if="!isMobile" class="username">{{ authStore.user?.username }}</span>
              <el-icon v-if="!isMobile"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main-content" :class="{ 'main-mobile': isMobile }">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Monitor, Odometer, Server, Bell, Message, Switch, Setting,
  Expand, Fold, UserFilled, ArrowDown, SwitchButton
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const isCollapse = ref(false)
const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
  if (isMobile.value) {
    isCollapse.value = true
  }
}

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const onMenuSelect = () => {
  // 移动端点击菜单后自动收起侧边栏
  if (isMobile.value) {
    isCollapse.value = true
  }
}

const handleCommand = async (command: string) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      authStore.logout()
      ElMessage.success('已退出登录')
      router.push('/login')
    } catch {
      // 取消
    }
  }
}

onMounted(checkMobile)
onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
// 使用 addEventListener 而非 onMounted 内直接写，确保能正确移除
window.addEventListener('resize', checkMobile)
</script>

<style scoped>
.layout-container {
  min-height: 100vh;
  background: #0f172a;
}

.sidebar {
  background: rgba(15, 23, 42, 0.95);
  border-right: 1px solid #1e293b;
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
  overflow: hidden;
}

/* 移动端侧边栏 */
.sidebar-mobile {
  position: fixed;
  z-index: 100;
  height: 100vh;
  width: 220px !important;
  transform: translateX(-100%);
  transition: transform 0.3s ease;
}
.sidebar-mobile.sidebar-open {
  transform: translateX(0);
}

/* 遮罩 */
.sidebar-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 99;
}

.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 12px;
  border-bottom: 1px solid #1e293b;
}

.sidebar-title {
  color: #f1f5f9;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.5px;
  white-space: nowrap;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
}

.sidebar-menu :deep(.el-menu-item:hover) {
  background: rgba(56, 189, 248, 0.08);
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background: rgba(56, 189, 248, 0.12);
}

.sidebar-footer {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-top: 1px solid #1e293b;
}

.collapse-btn {
  color: #64748b;
}

.collapse-btn:hover {
  color: #38bdf8;
}

.main-header {
  background: rgba(15, 23, 42, 0.95);
  border-bottom: 1px solid #1e293b;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.mobile-menu-btn {
  color: #94a3b8;
  margin-right: 8px;
}
.mobile-menu-btn:hover {
  color: #38bdf8;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #94a3b8;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.2s;
}

.user-info:hover {
  background: rgba(56, 189, 248, 0.08);
}

.username {
  font-size: 14px;
}

.main-content {
  padding: 20px;
  background: #0f172a;
}

.main-mobile {
  padding: 12px;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .main-header {
    padding: 0 12px;
  }
}
</style>
