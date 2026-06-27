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
        text-color="var(--cp-text-secondary)"
        active-text-color="var(--cp-accent)"
        @select="onMenuSelect"
      >
        <el-menu-item index="/">
          <el-icon><Odometer /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>

        <el-menu-item index="/servers">
          <el-icon><Cpu /></el-icon>
          <template #title>服务器</template>
        </el-menu-item>

        <el-menu-item index="/groups">
          <el-icon><Folder /></el-icon>
          <template #title>分组管理</template>
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
          <el-tooltip :content="theme === 'dark' ? '切换亮色' : theme === 'light' ? '切换暗色' : '跟随系统'" placement="bottom">
            <el-button type="text" class="theme-btn" @click="toggleTheme">
              <el-icon :size="18">
                <Moon v-if="isDark" />
                <Sunny v-else />
              </el-icon>
            </el-button>
          </el-tooltip>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="isMobile ? 28 : 32" :icon="UserFilled" />
              <span v-if="!isMobile" class="username">{{ authStore.user?.username }}</span>
              <el-icon v-if="!isMobile"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item divided command="lang-zh-CN">中文</el-dropdown-item>
                <el-dropdown-item command="lang-en-US">English</el-dropdown-item>
                <el-dropdown-item divided command="logout">
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
import { Monitor, Odometer, Cpu, Bell, Message, Switch, Setting,
  Expand, Fold, UserFilled, ArrowDown, SwitchButton, Folder, Sunny, Moon
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import { useI18n } from '@/i18n'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { theme, isDark, toggleTheme } = useTheme()
const { setLocale } = useI18n()
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
  if (command.startsWith('lang-')) {
    setLocale(command.replace('lang-', ''))
    return
  }
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
  background: var(--cp-bg-primary);
}

.sidebar {
  background: var(--cp-bg-card);
  border-right: 1px solid var(--cp-border);
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
  background: var(--cp-overlay);
  z-index: 99;
}

.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 12px;
  border-bottom: 1px solid var(--cp-border);
}

.sidebar-title {
  color: var(--cp-text-primary);
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
  background: var(--cp-accent-bg);
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background: var(--cp-accent-bg-strong);
}

.sidebar-footer {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-top: 1px solid var(--cp-border);
}

.collapse-btn {
  color: var(--cp-text-muted);
}

.collapse-btn:hover {
  color: var(--cp-accent);
}

.main-header {
  background: var(--cp-bg-card);
  border-bottom: 1px solid var(--cp-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.mobile-menu-btn {
  color: var(--cp-text-secondary);
  margin-right: 8px;
}
.mobile-menu-btn:hover {
  color: var(--cp-accent);
}

.header-right {
  display: flex;
  align-items: center;
}

.theme-btn {
  color: var(--cp-text-secondary);
  margin-right: 8px;
}
.theme-btn:hover {
  color: var(--cp-accent);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: var(--cp-text-secondary);
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.2s;
}

.user-info:hover {
  background: var(--cp-accent-bg);
}

.username {
  font-size: 14px;
}

.main-content {
  padding: 20px;
  background: var(--cp-bg-primary);
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
