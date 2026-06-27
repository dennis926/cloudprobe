<template>
  <div class="ssh-container">
    <div class="ssh-header">
      <div class="ssh-title">
        <el-icon><Monitor /></el-icon>
        <span>WebSSH - {{ serverName }}</span>
      </div>
      <div class="ssh-actions">
        <el-tag :type="connected ? 'success' : 'danger'" size="small">
          {{ connected ? '已连接' : '未连接' }}
        </el-tag>
        <el-button
          :type="connected ? 'danger' : 'primary'"
          size="small"
          @click="toggleConnection"
        >
          {{ connected ? '断开' : '连接' }}
        </el-button>
      </div>
    </div>
    <div ref="terminalRef" class="terminal-wrapper" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { Monitor } from '@element-plus/icons-vue'
import { Terminal as XTerm } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

const route = useRoute()
const serverId = route.params.serverId as string
const serverName = ref(`Server #${serverId}`)
const terminalRef = ref<HTMLDivElement>()
const connected = ref(false)

let term: XTerm | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null

const initTerminal = () => {
  if (!terminalRef.value) return

  term = new XTerm({
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: {
      background: '#0f172a',
      foreground: '#e2e8f0',
      cursor: '#38bdf8',
      selectionBackground: '#334155',
      black: '#0f172a',
      red: '#ef4444',
      green: '#22c55e',
      yellow: '#eab308',
      blue: '#38bdf8',
      magenta: '#a78bfa',
      cyan: '#22d3ee',
      white: '#e2e8f0'
    },
    cursorBlink: true,
    scrollback: 10000
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalRef.value)
  fitAddon.fit()

  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'data', data }))
    }
  })

  term.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resize', cols, rows }))
    }
  })
}

const connect = () => {
  if (ws) return

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/ssh/${serverId}`

  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    connected.value = true
    term?.writeln('\r\n\x1b[32m[Connected]\x1b[0m 连接成功\r\n')
  }

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'data' && term) {
        term.write(msg.data)
      } else if (msg.type === 'error') {
        term?.writeln(`\r\n\x1b[31m[Error] ${msg.data}\x1b[0m\r\n`)
      } else if (msg.type === 'heartbeat') {
        // ignore
      }
    } catch {
      // 非JSON消息直接写入
      term?.write(event.data)
    }
  }

  ws.onclose = () => {
    connected.value = false
    ws = null
    term?.writeln('\r\n\x1b[31m[Disconnected]\x1b[0m 连接已断开\r\n')
  }

  ws.onerror = () => {
    connected.value = false
    term?.writeln('\r\n\x1b[31m[Error] 连接失败\x1b[0m\r\n')
  }
}

const disconnect = () => {
  if (ws) {
    ws.close()
    ws = null
  }
  connected.value = false
}

const toggleConnection = () => {
  if (connected.value) {
    disconnect()
  } else {
    connect()
  }
}

const handleResize = () => {
  fitAddon?.fit()
}

onMounted(async () => {
  await nextTick()
  initTerminal()
  connect()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  disconnect()
  term?.dispose()
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.ssh-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 84px);
  background: #0f172a;
  border: 1px solid #1e293b;
  border-radius: 12px;
  overflow: hidden;
}

.ssh-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-bottom: 1px solid #1e293b;
  background: rgba(30, 41, 59, 0.6);
}

.ssh-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #f1f5f9;
  font-weight: 600;
}

.ssh-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.terminal-wrapper {
  flex: 1;
  padding: 8px;
  overflow: hidden;
}

.terminal-wrapper :deep(.xterm) {
  height: 100%;
}
</style>
