package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"cloudprobe/internal/database"
	"cloudprobe/internal/model"
	"cloudprobe/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var agentUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// AgentMessage Agent上报消息
type AgentMessage struct {
	Type     string                 `json:"type"` // heartbeat | metrics | system
	Token    string                 `json:"token"`
	Hostname string                 `json:"hostname,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Time     int64                  `json:"time"`
}

// DashboardMessage Dashboard下发消息
type DashboardMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// Connection Agent连接管理
type Connection struct {
	ws       *websocket.Conn
	serverID uint
	logger   *zap.Logger
	lastPing time.Time
	stopCh   chan struct{}
}

// activeConnections 活跃的Agent连接
type connectionManager struct {
	mu          sync.RWMutex
	connections map[uint]*Connection
	logger      *zap.Logger
}

var cm *connectionManager

func initConnectionManager(logger *zap.Logger) {
	cm = &connectionManager{
		connections: make(map[uint]*Connection),
		logger:      logger,
	}
}

func (m *connectionManager) add(serverID uint, conn *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.connections[serverID]; ok {
		old.Close()
	}
	m.connections[serverID] = conn
}

func (m *connectionManager) remove(serverID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, serverID)
}

func (m *connectionManager) get(serverID uint) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.connections[serverID]
	return conn, ok
}

// Init 初始化Agent WebSocket管理器
func Init(logger *zap.Logger) {
	initConnectionManager(logger)
}

// HandleAgentWebSocket 处理Agent WebSocket连接
func HandleAgentWebSocket(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := agentUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("agent websocket upgrade failed", zap.Error(err))
			return
		}
		defer ws.Close()

		// 等待第一条消息进行认证
		ws.SetReadDeadline(time.Now().Add(10 * time.Second))
		var authMsg AgentMessage
		if err := ws.ReadJSON(&authMsg); err != nil {
			logger.Warn("agent auth read failed", zap.Error(err))
			return
		}
		ws.SetReadDeadline(time.Time{})

		// 验证Token
		var server model.Server
		if err := database.GetDB().Where("agent_token = ?", authMsg.Token).First(&server).Error; err != nil {
			logger.Warn("agent auth failed", zap.String("token", authMsg.Token))
			ws.WriteJSON(DashboardMessage{Type: "auth_failed"})
			return
		}

		// 认证成功，更新服务器状态
		serverService := service.NewServerService()
		if err := serverService.UpdateServerStatus(server.ID, "online"); err != nil {
			logger.Error("update server status failed", zap.Error(err))
		}

		conn := &Connection{
			ws:       ws,
			serverID: server.ID,
			logger:   logger,
			lastPing: time.Now(),
			stopCh:   make(chan struct{}),
		}
		cm.add(server.ID, conn)
		defer cm.remove(server.ID)

		logger.Info("agent connected",
			zap.String("server", server.Name),
			zap.Uint("server_id", server.ID),
		)

		// 发送认证成功
		ws.WriteJSON(DashboardMessage{Type: "auth_success"})

		// 启动心跳检测
		go conn.heartbeatChecker()

		// 消息处理循环
		for {
			var msg AgentMessage
			if err := ws.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Warn("agent connection error",
						zap.Uint("server_id", server.ID),
						zap.Error(err),
					)
				}
				break
			}

			conn.lastPing = time.Now()

			switch msg.Type {
			case "heartbeat":
				// 更新心跳时间
			case "metrics":
				if err := handleMetrics(server.ID, &msg); err != nil {
					logger.Error("handle metrics failed", zap.Error(err))
				}
			case "system":
				if err := handleSystemInfo(server.ID, &msg); err != nil {
					logger.Error("handle system info failed", zap.Error(err))
				}
			default:
				logger.Debug("unknown agent message type", zap.String("type", msg.Type))
			}
		}

		// 连接断开，标记离线
		if err := serverService.UpdateServerStatus(server.ID, "offline"); err != nil {
			logger.Error("mark server offline failed", zap.Error(err))
		}

		logger.Info("agent disconnected", zap.Uint("server_id", server.ID))
	}
}

// heartbeatChecker 心跳检测
func (c *Connection) heartbeatChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Since(c.lastPing) > 90*time.Second {
				c.logger.Warn("agent heartbeat timeout", zap.Uint("server_id", c.serverID))
				c.Close()
				return
			}
			// 发送ping
			if err := c.ws.WriteJSON(DashboardMessage{Type: "ping"}); err != nil {
				c.Close()
				return
			}
		case <-c.stopCh:
			return
		}
	}
}

// Close 关闭连接
func (c *Connection) Close() {
	select {
	case <-c.stopCh:
		return
	default:
		close(c.stopCh)
	}
	c.ws.Close()
}

// Send 向Agent发送消息
func (c *Connection) Send(msg *DashboardMessage) error {
	return c.ws.WriteJSON(msg)
}

// handleMetrics 处理指标数据
func handleMetrics(serverID uint, msg *AgentMessage) error {
	if msg.Data == nil {
		return nil
	}
	return service.HandleMetricsFromAgent(serverID, msg.Data)
}

// handleSystemInfo 处理系统信息，更新服务器静态信息
func handleSystemInfo(serverID uint, msg *AgentMessage) error {
	if msg.Data == nil {
		return nil
	}
	svc := service.NewServerService()
	return svc.UpdateServerInfo(serverID, msg.Data)
}

// BroadcastToServer 向指定服务器发送命令
func BroadcastToServer(serverID uint, msg *DashboardMessage) error {
	conn, ok := cm.get(serverID)
	if !ok {
		return fmt.Errorf("server %d not connected", serverID)
	}
	return conn.Send(msg)
}
