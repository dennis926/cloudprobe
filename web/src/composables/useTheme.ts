import { ref, readonly } from 'vue'

export type ThemeMode = 'dark' | 'light' | 'system'

const STORAGE_KEY = 'theme'
const theme = ref<ThemeMode>((localStorage.getItem(STORAGE_KEY) as ThemeMode) || 'dark')
const isDark = ref(true)

let mediaQuery: MediaQueryList | null = null
let mediaListener: ((e: MediaQueryListEvent) => void) | null = null

function getSystemIsDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyTheme(mode: ThemeMode) {
  const dark = mode === 'dark' || (mode === 'system' && getSystemIsDark())
  isDark.value = dark
  document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light')
}

function setTheme(mode: ThemeMode) {
  theme.value = mode
  localStorage.setItem(STORAGE_KEY, mode)
  applyTheme(mode)
  updateMediaListener()
}

function toggleTheme() {
  const current = isDark.value
  setTheme(current ? 'light' : 'dark')
}

function updateMediaListener() {
  if (mediaListener && mediaQuery) {
    mediaQuery.removeEventListener('change', mediaListener)
    mediaListener = null
  }

  if (theme.value === 'system') {
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaListener = (e: MediaQueryListEvent) => {
      isDark.value = e.matches
      document.documentElement.setAttribute('data-theme', e.matches ? 'dark' : 'light')
    }
    mediaQuery.addEventListener('change', mediaListener)
  }
}

function initTheme() {
  applyTheme(theme.value)
  updateMediaListener()
}

export function useTheme() {
  return {
    theme: readonly(theme),
    isDark: readonly(isDark),
    setTheme,
    toggleTheme,
    initTheme,
  }
}
