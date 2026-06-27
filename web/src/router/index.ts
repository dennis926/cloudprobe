import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true }
    },
    {
      path: '/',
      component: () => import('@/views/LayoutView.vue'),
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/DashboardView.vue')
        },
        {
          path: 'servers',
          name: 'Servers',
          component: () => import('@/views/ServerListView.vue')
        },
        {
          path: 'servers/:id',
          name: 'ServerDetail',
          component: () => import('@/views/ServerDetailView.vue')
        },
        {
          path: 'groups',
          name: 'Groups',
          component: () => import('@/views/GroupView.vue')
        },
        {
          path: 'alerts',
          name: 'Alerts',
          component: () => import('@/views/AlertRulesView.vue')
        },
        {
          path: 'notifications',
          name: 'Notifications',
          component: () => import('@/views/NotificationView.vue')
        },
        {
          path: 'ssh/:serverId',
          name: 'SSH',
          component: () => import('@/views/SSHView.vue')
        },
        {
          path: 'proxy',
          name: 'Proxy',
          component: () => import('@/views/ProxyView.vue')
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/SettingsView.vue')
        }
      ]
    }
  ]
})

// 导航守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  authStore.loadUser()

  if (to.meta.public) {
    next()
    return
  }

  if (!authStore.isAuthenticated) {
    next('/login')
    return
  }

  next()
})

export default router
