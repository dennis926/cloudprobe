// @title CloudProbe API
// @version 1.0
// @description CloudProbe 服务器监控系统 API 文档
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer Token (格式: Bearer {token})
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

// @Summary 用户登录
// @Description 使用用户名密码获取 JWT Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body object{username=string,password=string} true "登录凭证"
// @Success 200 {object} api.SuccessResponse{data=object{access_token=string,refresh_token=string,user=object{id=uint,username=string,role=string}}}
// @Failure 401 {object} api.ErrorResponse
// @Router /auth/login [post]
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

// @Summary 刷新Token
// @Description 使用 Refresh Token 获取新的 Access Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body object{refresh_token=string} true "Refresh Token"
// @Success 200 {object} api.SuccessResponse{data=object{access_token=string,refresh_token=string}}
// @Failure 401 {object} api.ErrorResponse
// @Router /auth/refresh [post]
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

// @Summary 服务器列表
// @Description 获取所有服务器列表，支持按分组和状态筛选
// @Tags 服务器
// @Produce json
// @Security BearerAuth
// @Param group_id query int false "分组ID"
// @Param status query string false "状态(online/offline)"
// @Success 200 {object} api.SuccessResponse{data=[]model.Server}
// @Router /servers [get]
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

// @Summary 服务器详情
// @Tags 服务器
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} api.SuccessResponse{data=model.Server}
// @Failure 404 {object} api.ErrorResponse
// @Router /servers/{id} [get]
func (s *Server) handleGetServer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var server model.Server
	if err := database.GetDB().Preload("Group").Preload("Tags").Preload("Bill").First(&server, id).Error; err != nil {
		api.JSONError(c, http.StatusNotFound, "server not found")
		return
	}
	api.JSONSuccess(c, server)
}

// @Summary 创建服务器
// @Tags 服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.Server true "服务器信息"
// @Success 200 {object} api.SuccessResponse{data=model.Server}
// @Router /servers [post]
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

// @Summary 更新服务器
// @Tags 服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param body body model.Server true "更新信息"
// @Success 200 {object} api.SuccessResponse
// @Router /servers/{id} [put]
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

// @Summary 删除服务器
// @Tags 服务器
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} api.SuccessResponse
// @Router /servers/{id} [delete]
func (s *Server) handleDeleteServer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&model.Server{}, id).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to delete server")
		return
	}
	api.JSONSuccess(c, nil)
}

// ==================== Metrics Handlers ====================

// @Summary 服务器监控指标
// @Description 获取服务器历史监控指标数据
// @Tags 监控
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param type query string false "指标类型(cpu/memory/disk)"
// @Param start query string false "开始时间(RFC3339)"
// @Param end query string false "结束时间(RFC3339)"
// @Success 200 {object} api.SuccessResponse
// @Router /servers/{id}/metrics [get]
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

// @Summary 实时指标
// @Description 获取所有服务器的最新实时指标
// @Tags 监控
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /metrics/realtime [get]
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

// @Summary 告警列表
// @Tags 告警
// @Produce json
// @Security BearerAuth
// @Param status query string false "状态(firing/resolved)"
// @Success 200 {object} api.SuccessResponse
// @Router /alerts [get]
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

// @Summary 告警规则列表
// @Tags 告警
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /alerts/rules [get]
func (s *Server) handleListAlertRules(c *gin.Context) {
	var rules []model.AlertRule
	if err := database.GetDB().Find(&rules).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to list rules")
		return
	}
	api.JSONSuccess(c, rules)
}

// @Summary 创建告警规则
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.AlertRule true "告警规则"
// @Success 200 {object} api.SuccessResponse
// @Router /alerts/rules [post]
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

// @Summary 更新告警规则
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "规则ID"
// @Param body body model.AlertRule true "更新内容"
// @Success 200 {object} api.SuccessResponse
// @Router /alerts/rules/{id} [put]
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

// @Summary 删除告警规则
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "规则ID"
// @Success 200 {object} api.SuccessResponse
// @Router /alerts/rules/{id} [delete]
func (s *Server) handleDeleteAlertRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&model.AlertRule{}, id).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to delete rule")
		return
	}
	api.JSONSuccess(c, nil)
}

// ==================== Notification Handlers ====================

// @Summary 通知渠道列表
// @Tags 通知
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /notifications/channels [get]
func (s *Server) handleListChannels(c *gin.Context) {
	var channels []model.NotificationChannel
	if err := database.GetDB().Find(&channels).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to list channels")
		return
	}
	api.JSONSuccess(c, channels)
}

// @Summary 创建通知渠道
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.NotificationChannel true "渠道配置"
// @Success 200 {object} api.SuccessResponse
// @Router /notifications/channels [post]
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

// @Summary 测试通知
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object{channel_id=uint} true "渠道ID"
// @Success 200 {object} api.SuccessResponse
// @Router /notifications/test [post]
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

// @Summary 3x-ui 连接状态
// @Tags 代理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /proxy/status [get]
func (s *Server) handleProxyStatus(c *gin.Context) {
	api.JSONSuccess(c, gin.H{
		"connected": s.cfg.XUI.Enabled,
		"panel_url": s.cfg.XUI.PanelURL,
	})
}

// @Summary 配置3x-ui
// @Tags 代理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object{panel_url=string,api_token=string} true "面板配置"
// @Success 200 {object} api.SuccessResponse
// @Router /proxy/config [post]
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

// @Summary 入站列表
// @Tags 代理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /proxy/inbounds [get]
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

// @Summary 客户端列表
// @Tags 代理
// @Produce json
// @Security BearerAuth
// @Param inbound_id query int true "入站ID"
// @Success 200 {object} api.SuccessResponse
// @Router /proxy/clients [get]
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

// @Summary 节点列表
// @Tags 代理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /proxy/nodes [get]
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

// @Summary Xray 运行状态
// @Tags 代理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /proxy/xray/status [get]
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

// @Summary 系统信息
// @Tags 系统
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /system/info [get]
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

// @Summary 获取设置
// @Tags 系统
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /system/settings [get]
func (s *Server) handleGetSettings(c *gin.Context) {
	api.JSONSuccess(c, s.cfg)
}

// @Summary 更新设置
// @Tags 系统
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "设置内容"
// @Success 200 {object} api.SuccessResponse
// @Router /system/settings [put]
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

// ==================== ServerGroup Handlers ====================

// @Summary 创建分组
// @Tags 分组
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.ServerGroup true "分组信息"
// @Success 200 {object} api.SuccessResponse
// @Router /groups [post]
func (s *Server) handleCreateGroup(c *gin.Context) {
	var req model.ServerGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := database.GetDB().Create(&req).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to create group")
		return
	}

	api.JSONSuccess(c, req)
}

// @Summary 分组列表
// @Tags 分组
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.SuccessResponse
// @Router /groups [get]
func (s *Server) handleListGroups(c *gin.Context) {
	var groups []model.ServerGroup
	db := database.GetDB().Order("sort_order ASC, id ASC")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		db = db.Offset(offset).Limit(pageSize)
	}

	if err := db.Find(&groups).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to list groups")
		return
	}

	api.JSONSuccess(c, groups)
}

// @Summary 更新分组
// @Tags 分组
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分组ID"
// @Param body body model.ServerGroup true "更新内容"
// @Success 200 {object} api.SuccessResponse
// @Router /groups/{id} [put]
func (s *Server) handleUpdateGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.ServerGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := database.GetDB().Model(&model.ServerGroup{}).Where("id = ?", id).Updates(&req).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to update group")
		return
	}

	api.JSONSuccess(c, nil)
}

// @Summary 删除分组
// @Tags 分组
// @Produce json
// @Security BearerAuth
// @Param id path int true "分组ID"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /groups/{id} [delete]
func (s *Server) handleDeleteGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	// 检查是否有关联服务器
	var count int64
	if err := database.GetDB().Model(&model.Server{}).Where("group_id = ?", id).Count(&count).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to check group")
		return
	}
	if count > 0 {
		api.JSONError(c, http.StatusBadRequest, "group has associated servers, cannot delete")
		return
	}

	if err := database.GetDB().Delete(&model.ServerGroup{}, id).Error; err != nil {
		api.JSONError(c, http.StatusInternalServerError, "failed to delete group")
		return
	}

	api.JSONSuccess(c, nil)
}
