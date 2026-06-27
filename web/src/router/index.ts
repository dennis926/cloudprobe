import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/LoginView.vue')
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

export default router
