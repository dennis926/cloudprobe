import { ref, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

// WebSocket 实时数据 Hook
export function useWebSocket() {
  const connected = ref(false)
  const realtimeData = ref<Record<number, any>>({})
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null

  const authStore = useAuthStore()

  const connect = () => {
    if (!authStore.isAuthenticated) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/realtime`

    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      connected.value = true
      // 心跳保活
      heartbeatTimer = setInterval(() => {
        if (ws?.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'ping' }))
        }
      }, 30000)
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'metrics' && msg.data) {
          // 更新实时数据
          for (const [serverId, metrics] of Object.entries(msg.data)) {
            realtimeData.value[Number(serverId)] = metrics
          }
        } else if (msg.type === 'pong') {
          // ignore
        }
      } catch {
        // ignore
      }
    }

    ws.onclose = () => {
      connected.value = false
      if (heartbeatTimer) {
        clearInterval(heartbeatTimer)
        heartbeatTimer = null
      }
      // 自动重连（5秒后）
      reconnectTimer = setTimeout(() => {
        connect()
      }, 5000)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  const disconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
    if (ws) {
      ws.close()
      ws = null
    }
    connected.value = false
  }

  const getMetrics = (serverId: number) => {
    return realtimeData.value[serverId] || null
  }

  onUnmounted(disconnect)

  return {
    connected,
    realtimeData,
    connect,
    disconnect,
    getMetrics
  }
}
