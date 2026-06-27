import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

const request: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response
    if (data.code !== undefined && data.code !== 0 && data.code !== 200) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message))
    }
    return data
  },
  async (error) => {
    const { response } = error
    if (response) {
      switch (response.status) {
        case 401:
          ElMessage.error('登录已过期，请重新登录')
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          router.push('/login')
          break
        case 403:
          ElMessage.error('权限不足')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          ElMessage.error(response.data?.message || '网络错误')
      }
    } else {
      ElMessage.error('网络连接失败')
    }
    return Promise.reject(error)
  }
)

export default request

// API 封装
export const api = {
  // 认证
  login: (username: string, password: string) =>
    request.post('/auth/login', { username, password }),
  refresh: (refreshToken: string) =>
    request.post('/auth/refresh', { refresh_token: refreshToken }),

  // 服务器
  getServers: (params?: any) => request.get('/servers', { params }),
  getServer: (id: number) => request.get(`/servers/${id}`),
  createServer: (data: any) => request.post('/servers', data),
  updateServer: (id: number, data: any) => request.put(`/servers/${id}`, data),
  deleteServer: (id: number) => request.delete(`/servers/${id}`),

  // 监控
  getMetrics: (id: number) => request.get(`/servers/${id}/metrics`),
  getRealtime: () => request.get('/metrics/realtime'),

  // 告警
  getAlerts: () => request.get('/alerts'),
  getAlertRules: () => request.get('/alerts/rules'),
  createAlertRule: (data: any) => request.post('/alerts/rules', data),
  updateAlertRule: (id: number, data: any) => request.put(`/alerts/rules/${id}`, data),
  deleteAlertRule: (id: number) => request.delete(`/alerts/rules/${id}`),

  // 通知渠道
  getChannels: () => request.get('/notifications/channels'),
  createChannel: (data: any) => request.post('/notifications/channels', data),
  testNotify: (data: any) => request.post('/notifications/test', data),

  // 系统
  getSystemInfo: () => request.get('/system/info'),
  getSettings: () => request.get('/system/settings'),
  updateSettings: (data: any) => request.put('/system/settings', data),

  // 分组
  getGroups: (params?: any) => request.get('/groups', { params }),
  createGroup: (data: any) => request.post('/groups', data),
  updateGroup: (id: number, data: any) => request.put(`/groups/${id}`, data),
  deleteGroup: (id: number) => request.delete(`/groups/${id}`),

  // 代理
  getProxyStatus: () => request.get('/proxy/status'),
  getProxyInbounds: () => request.get('/proxy/inbounds'),
}
