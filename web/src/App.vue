<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useTheme } from '@/composables/useTheme'

const { initTheme } = useTheme()
onMounted(initTheme)
</script>

<style>
html, body, #app {
  height: 100%;
  margin: 0;
  background: var(--cp-bg-primary);
  transition: background-color 0.3s;
}

/* 暗色 CSS 变量 */
[data-theme="dark"] {
  --cp-bg-primary: #0f172a;
  --cp-bg-secondary: rgba(30, 41, 59, 0.6);
  --cp-bg-card: rgba(15, 23, 42, 0.95);
  --cp-bg-sidebar: rgba(15, 23, 42, 0.95);
  --cp-border: #1e293b;
  --cp-text-primary: #f1f5f9;
  --cp-text-secondary: #94a3b8;
  --cp-text-muted: #64748b;
  --cp-text-table: #cbd5e1;
  --cp-accent: #38bdf8;
  --cp-accent-bg: rgba(56, 189, 248, 0.08);
  --cp-accent-bg-strong: rgba(56, 189, 248, 0.12);
  --cp-success: #22c55e;
  --cp-danger: #ef4444;
  --cp-warning: #f59e0b;
  --cp-overlay: rgba(0, 0, 0, 0.5);
  /* Element Plus 暗色覆盖 */
  --el-bg-color: #0f172a;
  --el-bg-color-overlay: #1e293b;
  --el-text-color-primary: #f1f5f9;
  --el-text-color-regular: #cbd5e1;
  --el-border-color: #1e293b;
  --el-fill-color-blank: #1e293b;
}

/* 亮色 CSS 变量 */
[data-theme="light"] {
  --cp-bg-primary: #f1f5f9;
  --cp-bg-secondary: #ffffff;
  --cp-bg-card: #ffffff;
  --cp-bg-sidebar: #ffffff;
  --cp-border: #e2e8f0;
  --cp-text-primary: #0f172a;
  --cp-text-secondary: #475569;
  --cp-text-muted: #94a3b8;
  --cp-text-table: #334155;
  --cp-accent: #0284c7;
  --cp-accent-bg: rgba(2, 132, 199, 0.08);
  --cp-accent-bg-strong: rgba(2, 132, 199, 0.12);
  --cp-success: #16a34a;
  --cp-danger: #dc2626;
  --cp-warning: #d97706;
  --cp-overlay: rgba(0, 0, 0, 0.3);
}
</style>
