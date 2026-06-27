package ssh

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"cloudprobe/internal/api"
	"cloudprobe/internal/database"
	"cloudprobe/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应限制来源
	},
}

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type string `json:"type"` // resize | data | heartbeat
	Data string `json:"data"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

// HandleWebSSH 处理WebSSH WebSocket连接
func HandleWebSSH(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			api.JSONError(c, http.StatusBadRequest, "invalid server id")
			return
		}

		// 获取服务器信息
		var server model.Server
		if err := database.GetDB().First(&server, uint(serverID)).Error; err != nil {
			api.JSONError(c, http.StatusNotFound, "server not found")
			return
		}

		// 升级WebSocket
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("websocket upgrade failed", zap.Error(err))
			return
		}
		defer ws.Close()

		// 建立SSH连接
		sshCfg := &Config{
			Host:     server.PublicIP,
			Port:     server.SSHPort,
			User:     server.SSHUser,
			Password: server.SSHPassword,
			Timeout:  10 * time.Second,
		}

		if sshCfg.Port == 0 {
			sshCfg.Port = 22
		}

		client, err := NewClient(sshCfg)
		if err != nil {
			ws.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
			return
		}
		defer client.Close()

		// 打开shell，默认80x24
		if err := client.OpenShell(24, 80); err != nil {
			ws.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
			return
		}

		logger.Info("webssh connected",
			zap.String("server", server.Name),
			zap.String("host", sshCfg.Host),
		)

		// 启动双向数据转发
		done := make(chan struct{})

		// SSH stdout -> WebSocket
		go func() {
			defer close(done)
			buf := make([]byte, 4096)
			for {
				n, err := client.ReadStdout().Read(buf)
				if err != nil {
					if err != io.EOF {
						logger.Warn("ssh stdout read error", zap.Error(err))
					}
					return
				}
				if n > 0 {
					msg := WSMessage{Type: "data", Data: string(buf[:n])}
					if err := ws.WriteJSON(msg); err != nil {
						return
					}
				}
			}
		}()

		// WebSocket -> SSH stdin
		go func() {
			for {
				var msg WSMessage
				if err := ws.ReadJSON(&msg); err != nil {
					return
				}

				switch msg.Type {
				case "data":
					client.Write([]byte(msg.Data))
				case "resize":
					if msg.Cols > 0 && msg.Rows > 0 {
						client.Resize(msg.Rows, msg.Cols)
					}
				case "heartbeat":
					ws.WriteJSON(WSMessage{Type: "heartbeat", Data: "pong"})
				}
			}
		}()

		// 等待连接结束
		select {
		case <-done:
		case <-client.Done():
		}

		logger.Info("webssh disconnected", zap.String("server", server.Name))
	}
}

// HandleWebSSHDirect 直接通过URL参数连接（用于没有预配置服务器的情况）
func HandleWebSSHDirect(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Query("host")
		portStr := c.DefaultQuery("port", "22")
		user := c.Query("user")
		password := c.Query("password")

		if host == "" || user == "" {
			api.JSONError(c, http.StatusBadRequest, "host and user required")
			return
		}

		port, _ := strconv.Atoi(portStr)
		if port == 0 {
			port = 22
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("websocket upgrade failed", zap.Error(err))
			return
		}
		defer ws.Close()

		sshCfg := &Config{
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			Timeout:  10 * time.Second,
		}

		client, err := NewClient(sshCfg)
		if err != nil {
			ws.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
			return
		}
		defer client.Close()

		if err := client.OpenShell(24, 80); err != nil {
			ws.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
			return
		}

		done := make(chan struct{})

		go func() {
			defer close(done)
			buf := make([]byte, 4096)
			for {
				n, err := client.ReadStdout().Read(buf)
				if err != nil {
					return
				}
				if n > 0 {
					ws.WriteJSON(WSMessage{Type: "data", Data: string(buf[:n])})
				}
			}
		}()

		go func() {
			for {
				var msg WSMessage
				if err := ws.ReadJSON(&msg); err != nil {
					return
				}
				switch msg.Type {
				case "data":
					client.Write([]byte(msg.Data))
				case "resize":
					client.Resize(msg.Rows, msg.Cols)
				}
			}
		}()

		<-done
	}
}

// writeJSON 辅助函数
func writeJSON(ws *websocket.Conn, msg WSMessage) error {
	data, _ := json.Marshal(msg)
	return ws.WriteMessage(websocket.TextMessage, data)
}
