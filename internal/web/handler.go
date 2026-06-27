package web

import (
	"net/http"
	"strconv"
	"time"

	"cloudprobe/internal/agent"
	"cloudprobe/internal/api"
	"cloudprobe/internal/auth"
	"cloudprobe/internal/database"
	"cloudprobe/internal/model"
	sshweb "cloudprobe/internal/ssh"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==================== Auth Handlers ====================

func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	var user model.User
	if err := database.GetDB().Where("username = ?", req.Username).First(&user).Error; err != nil {
		api.JSONError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if user.Status != "active" {
		api.JSONError(c, http.StatusForbidden, "account disabled")
		return
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		api.JSONError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, refreshToken, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	api.JSONSuccess(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (s *Server) handleRefresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	claims, err := auth.ParseToken(req.RefreshToken)
	if err != nil {
		api.JSONError(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	accessToken, refreshToken, err := auth.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	api.JSONSuccess(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// ==================== Server Handlers ====================

func (s *Server) handleListServers(c *gin.Context) {
	var servers []model.Server
	db := database.GetDB().Preload("Group").Preload("Tags").Preload("Bill")

	if groupID := c.Query("group_id"); groupID != "" {
		db = db.Where("group_id = ?", groupID)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Find(&servers).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to list servers")
		return
	}

	api.JSONSuccess(c, servers)
}

func (s *Server) handleGetServer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var server model.Server
	if err := database.GetDB().Preload("Group").Preload("Tags").Preload("Bill").First(&server, id).Error; err != nil {
		api.JSONError(c, http.StatusNotFound, "server not found")
		return
	}
	api.JSONSuccess(c, server)
}

func (s *Server) handleCreateServer(c *gin.Context) {
	var req model.Server
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	req.AgentToken = uuid.New().String()
	req.Status = "offline"

	if err := database.GetDB().Create(&req).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to create server")
		return
	}

	api.JSONSuccess(c, req)
}

func (s *Server) handleUpdateServer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.Server
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := database.GetDB().Model(&model.Server{}).Where("id = ?", id).Updates(&req).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to update server")
		return
	}

	api.JSONSuccess(c, nil)
}

func (s *Server) handleDeleteServer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&model.Server{}, id).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to delete server")
		return
	}
	api.JSONSuccess(c, nil)
}

// ==================== Metrics Handlers ====================

func (s *Server) handleGetMetrics(c *gin.Context) {
	// TODO: 从TimescaleDB查询历史监控数据
	api.JSONSuccess(c, gin.H{
		"server_id": c.Param("id"),
		"message":   "metrics endpoint - implement with TimescaleDB",
	})
}

func (s *Server) handleGetRealtime(c *gin.Context) {
	// TODO: 从Redis获取实时状态
	api.JSONSuccess(c, gin.H{
		"message": "realtime endpoint - implement with Redis",
	})
}

// ==================== Alert Handlers ====================

func (s *Server) handleListAlerts(c *gin.Context) {
	var alerts []model.Alert
	db := database.GetDB().Preload("Rule").Preload("Server")
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Order("started_at DESC").Find(&alerts).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to list alerts")
		return
	}
	api.JSONSuccess(c, alerts)
}

func (s *Server) handleListAlertRules(c *gin.Context) {
	var rules []model.AlertRule
	if err := database.GetDB().Find(&rules).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to list rules")
		return
	}
	api.JSONSuccess(c, rules)
}

func (s *Server) handleCreateAlertRule(c *gin.Context) {
	var req model.AlertRule
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}
	if err := database.GetDB().Create(&req).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to create rule")
		return
	}
	api.JSONSuccess(c, req)
}

func (s *Server) handleUpdateAlertRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.AlertRule
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}
	if err := database.GetDB().Model(&model.AlertRule{}).Where("id = ?", id).Updates(&req).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to update rule")
		return
	}
	api.JSONSuccess(c, nil)
}

func (s *Server) handleDeleteAlertRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&model.AlertRule{}, id).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to delete rule")
		return
	}
	api.JSONSuccess(c, nil)
}

// ==================== Notification Handlers ====================

func (s *Server) handleListChannels(c *gin.Context) {
	var channels []model.NotificationChannel
	if err := database.GetDB().Find(&channels).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to list channels")
		return
	}
	api.JSONSuccess(c, channels)
}

func (s *Server) handleCreateChannel(c *gin.Context) {
	var req model.NotificationChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}
	if err := database.GetDB().Create(&req).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to create channel")
		return
	}
	api.JSONSuccess(c, req)
}

func (s *Server) handleTestNotify(c *gin.Context) {
	// TODO: 发送测试通知
	api.JSONSuccess(c, gin.H{"message": "test notification sent"})
}

// ==================== Proxy (3x-ui) Handlers ====================

func (s *Server) handleProxyStatus(c *gin.Context) {
	api.JSONSuccess(c, gin.H{
		"connected": s.cfg.XUI.Enabled,
		"panel_url": s.cfg.XUI.PanelURL,
	})
}

func (s *Server) handleProxyConfig(c *gin.Context) {
	var req struct {
		PanelURL string `json:"panel_url"`
		APIToken string `json:"api_token"`
		BasePath string `json:"base_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}
	s.cfg.XUI.PanelURL = req.PanelURL
	s.cfg.XUI.APIToken = req.APIToken
	s.cfg.XUI.BasePath = req.BasePath
	s.cfg.XUI.Enabled = req.PanelURL != "" && req.APIToken != ""
	api.JSONSuccess(c, nil)
}

func (s *Server) handleProxyInbounds(c *gin.Context) {
	api.JSONSuccess(c, gin.H{"message": "proxy inbounds - implement with 3x-ui API"})
}

func (s *Server) handleProxyClients(c *gin.Context) {
	api.JSONSuccess(c, gin.H{"message": "proxy clients - implement with 3x-ui API"})
}

func (s *Server) handleProxyNodes(c *gin.Context) {
	api.JSONSuccess(c, gin.H{"message": "proxy nodes - implement with 3x-ui API"})
}

func (s *Server) handleProxyXrayStatus(c *gin.Context) {
	api.JSONSuccess(c, gin.H{"message": "proxy xray status - implement with 3x-ui API"})
}

// ==================== System Handlers ====================

func (s *Server) handleSystemInfo(c *gin.Context) {
	api.JSONSuccess(c, gin.H{
		"version":   "1.0.0",
		"agent_count": 0,
		"server_count": 0,
	})
}

func (s *Server) handleGetSettings(c *gin.Context) {
	api.JSONSuccess(c, s.cfg)
}

func (s *Server) handleUpdateSettings(c *gin.Context) {
	// TODO: 更新系统设置
	api.JSONSuccess(c, nil)
}

// ==================== WebSocket Handlers ====================

func (s *Server) handleWebSSH(c *gin.Context) {
	sshweb.HandleWebSSH(s.logger)(c)
}

func (s *Server) handleWebSSHDirect(c *gin.Context) {
	sshweb.HandleWebSSHDirect(s.logger)(c)
}

func (s *Server) handleAgentWS(c *gin.Context) {
	agent.HandleAgentWebSocket(s.logger)(c)
}
