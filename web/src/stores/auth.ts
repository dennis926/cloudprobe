import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/request'

interface User {
  id: number
  username: string
  role: string
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string>(localStorage.getItem('access_token') || '')
  const refreshToken = ref<string>(localStorage.getItem('refresh_token') || '')
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  const setTokens = (at: string, rt: string) => {
    accessToken.value = at
    refreshToken.value = rt
    localStorage.setItem('access_token', at)
    localStorage.setItem('refresh_token', rt)
  }

  const setUser = (u: User) => {
    user.value = u
    localStorage.setItem('user', JSON.stringify(u))
  }

  const loadUser = () => {
    const stored = localStorage.getItem('user')
    if (stored) {
      try {
        user.value = JSON.parse(stored)
      } catch {
        user.value = null
      }
    }
  }

  const login = async (username: string, password: string) => {
    const res: any = await api.login(username, password)
    if (res.data) {
      setTokens(res.data.access_token, res.data.refresh_token)
      setUser(res.data.user)
    }
    return res
  }

  const logout = () => {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
  }

  const refresh = async () => {
    if (!refreshToken.value) return false
    try {
      const res: any = await api.refresh(refreshToken.value)
      if (res.data) {
        setTokens(res.data.access_token, res.data.refresh_token)
        return true
      }
    } catch {
      logout()
    }
    return false
  }

  return {
    accessToken,
    refreshToken,
    user,
    isAuthenticated,
    isAdmin,
    login,
    logout,
    refresh,
    loadUser,
    setTokens,
    setUser
  }
})
