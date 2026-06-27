package web

import (
	"net/http"
	"strconv"
	"time"

	"cloudprobe/internal/agent"
	"cloudprobe/internal/api"
	"cloudprobe/internal/auth"
	"cloudprobe/internal/cache"
	"cloudprobe/internal/database"
	"cloudprobe/internal/model"
	"cloudprobe/internal/proxy"
	"cloudprobe/internal/service"
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
	serverID, _ := strconv.Atoi(c.Param("id"))
	metricType := c.Query("type")

	// 默认查询最近24小时
	end := time.Now()
	start := end.Add(-24 * time.Hour)
	if startStr := c.Query("start"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr := c.Query("end"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	svc := service.NewMetricService()
	metrics, err := svc.QueryMetrics(uint(serverID), metricType, start, end)
	if err != nil {
		api.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSONSuccess(c, metrics)
}

func (s *Server) handleGetRealtime(c *gin.Context) {
	svc := service.NewMetricService()
	metrics, err := svc.GetLatestMetrics()
	if err != nil {
		api.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSONSuccess(c, metrics)
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
	var req struct {
		ChannelID uint `json:"channel_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}
	if service.GlobalNotifyService == nil {
		api.JSONError(c, http.StatusInternalServerError, "notify service not initialized")
		return
	}
	if err := service.GlobalNotifyService.TestChannel(c.Request.Context(), req.ChannelID); err != nil {
		api.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
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
	if !s.cfg.XUI.Enabled {
		api.JSONError(c, http.StatusBadRequest, "3x-ui not configured")
		return
	}
	client := proxy.NewClient(&s.cfg.XUI)
	inbounds, err := client.GetInbounds(c.Request.Context())
	if err != nil {
		api.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSONSuccess(c, inbounds)
}

func (s *Server) handleProxyClients(c *gin.Context) {
	if !s.cfg.XUI.Enabled {
		api.JSONError(c, http.StatusBadRequest, "3x-ui not configured")
		return
	}
	inboundID, _ := strconv.Atoi(c.Query("inbound_id"))
	if inboundID == 0 {
		api.JSONError(c, http.StatusBadRequest, "inbound_id required")
		return
	}
	client := proxy.NewClient(&s.cfg.XUI)
	clients, err := client.GetClients(c.Request.Context(), inboundID)
	if err != nil {
		api.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSONSuccess(c, clients)
}

func (s *Server) handleProxyNodes(c *gin.Context) {
	if !s.cfg.XUI.Enabled {
		api.JSONError(c, http.StatusBadRequest, "3x-ui not configured")
		return
	}
	client := proxy.NewClient(&s.cfg.XUI)
	nodes, err := client.GetNodes(c.Request.Context())
	if err != nil {
		api.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSONSuccess(c, nodes)
}

func (s *Server) handleProxyXrayStatus(c *gin.Context) {
	if !s.cfg.XUI.Enabled {
		api.JSONError(c, http.StatusBadRequest, "3x-ui not configured")
		return
	}
	client := proxy.NewClient(&s.cfg.XUI)
	status, err := client.GetXrayStatus(c.Request.Context())
	if err != nil {
		api.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSONSuccess(c, status)
}

// ==================== System Handlers ====================

func (s *Server) handleSystemInfo(c *gin.Context) {
	var total, online int64
	database.GetDB().Model(&model.Server{}).Count(&total)
	database.GetDB().Model(&model.Server{}).Where("status = ?", "online").Count(&online)

	var firingAlerts int64
	database.GetDB().Model(&model.Alert{}).Where("status = ?", "firing").Count(&firingAlerts)

	api.JSONSuccess(c, gin.H{
		"version":       "1.0.0",
		"server_count":  total,
		"online_count":  online,
		"alert_count":   firingAlerts,
	})
}

func (s *Server) handleGetSettings(c *gin.Context) {
	api.JSONSuccess(c, s.cfg)
}

func (s *Server) handleUpdateSettings(c *gin.Context) {
	var req struct {
		Server struct {
			Mode string `json:"mode"`
			Port int    `json:"port"`
		} `json:"server"`
		JWT struct {
			AccessExpire  int    `json:"access_expire"`
			RefreshExpire int    `json:"refresh_expire"`
			Secret        string `json:"secret"`
		} `json:"jwt"`
		SMTP struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			User     string `json:"user"`
			Password string `json:"password"`
			From     string `json:"from"`
		} `json:"smtp"`
		Backup struct {
			Enabled   bool   `json:"enabled"`
			Email     string `json:"email"`
			Schedule  string `json:"schedule"`
			KeepLocal int    `json:"keep_local"`
		} `json:"backup"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// 应用运行时配置更新（注意：重启后从配置文件读取）
	s.cfg.Server.Mode = req.Server.Mode
	s.cfg.Server.Port = req.Server.Port
	s.cfg.JWT.AccessExpire = req.JWT.AccessExpire
	s.cfg.JWT.RefreshExpire = req.JWT.RefreshExpire
	s.cfg.JWT.Secret = req.JWT.Secret
	s.cfg.SMTP.Host = req.SMTP.Host
	s.cfg.SMTP.Port = req.SMTP.Port
	s.cfg.SMTP.User = req.SMTP.User
	s.cfg.SMTP.Password = req.SMTP.Password
	s.cfg.SMTP.From = req.SMTP.From
	s.cfg.Backup.Enabled = req.Backup.Enabled
	s.cfg.Backup.Email = req.Backup.Email
	s.cfg.Backup.Schedule = req.Backup.Schedule
	s.cfg.Backup.KeepLocal = req.Backup.KeepLocal

	api.JSONSuccess(c, gin.H{"message": "settings updated (runtime only, restart to persist)"})
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
