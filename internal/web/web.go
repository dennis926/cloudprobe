package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloudprobe/internal/agent"
	"cloudprobe/internal/api"
	"cloudprobe/internal/auth"
	"cloudprobe/internal/cache"
	"cloudprobe/internal/config"
	"cloudprobe/internal/database"
	grpcserver "cloudprobe/internal/grpc"
	"cloudprobe/internal/service"
	"cloudprobe/internal/task"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Server Web服务器
type Server struct {
	cfg         *config.Config
	router      *gin.Engine
	http        *http.Server
	logger      *zap.Logger
	alertEngine *service.AlertEngine
	scheduler   *task.Scheduler
	grpcServer  interface{ Stop() }
}

// NewServer 创建Web服务器
func NewServer(cfg *config.Config) (*Server, error) {
	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	// 初始化数据库
	if err := database.Init(&cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// 初始化Redis
	if err := cache.Init(&cfg.Redis); err != nil {
		logger.Warn("Redis init failed, continuing without cache", zap.Error(err))
	}

	// 初始化通知服务
	service.InitNotifyService(logger)

	// 初始化Agent WebSocket管理器
	agent.Init(logger)

	// 初始化告警引擎
	alertEngine := service.InitAlertEngine(logger)
	alertEngine.Start()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由器
	router := gin.New()
	router.Use(api.RecoveryMiddleware())
	router.Use(api.LoggerMiddleware(logger))
	router.Use(api.CORSMiddleware())

	// 初始化定时任务调度器
	scheduler := task.NewScheduler(logger)
	scheduler.Start()

	// 启动gRPC服务（海外部署方案）
	var grpcSvr interface{ Stop() }
	if cfg.Agent.GRPCAddr != "" {
		gs, err := grpcserver.StartGRPCServer(cfg.Agent.GRPCAddr, logger)
		if err != nil {
			logger.Warn("gRPC server failed to start", zap.Error(err))
		} else {
			grpcSvr = gs
		}
	}

	s := &Server{
		cfg:         cfg,
		router:      router,
		logger:      logger,
		alertEngine: alertEngine,
		scheduler:   scheduler,
		grpcServer:  grpcSvr,
	}

	s.initRoutes()

	return s, nil
}

// initRoutes 初始化路由
func (s *Server) initRoutes() {
	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger 文档（需先执行 swag init -g cmd/dashboard/main.go -o docs 生成 docs 目录）
	// 启用后取消下方注释并添加 _ "cloudprobe/docs" 到 import
	if false {
		s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API路由组
	apiGroup := s.router.Group("/api/v1")
	{
		// 认证（公开）
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/login", s.handleLogin)
			authGroup.POST("/refresh", s.handleRefresh)
		}

		// 需要认证的路由
		authorized := apiGroup.Group("")
		authorized.Use(auth.AuthMiddleware())
		{
			// 服务器管理
			authorized.GET("/servers", s.handleListServers)
			authorized.GET("/servers/:id", s.handleGetServer)
			authorized.POST("/servers", auth.AdminRequired(), s.handleCreateServer)
			authorized.PUT("/servers/:id", auth.AdminRequired(), s.handleUpdateServer)
			authorized.DELETE("/servers/:id", auth.AdminRequired(), s.handleDeleteServer)

			// 服务器分组
			authorized.GET("/groups", s.handleListGroups)
			authorized.POST("/groups", auth.AdminRequired(), s.handleCreateGroup)
			authorized.PUT("/groups/:id", auth.AdminRequired(), s.handleUpdateGroup)
			authorized.DELETE("/groups/:id", auth.AdminRequired(), s.handleDeleteGroup)

			// 监控数据
			authorized.GET("/servers/:id/metrics", s.handleGetMetrics)
			authorized.GET("/metrics/realtime", s.handleGetRealtime)

			// 告警
			authorized.GET("/alerts", s.handleListAlerts)
			authorized.GET("/alerts/rules", s.handleListAlertRules)
			authorized.POST("/alerts/rules", auth.AdminRequired(), s.handleCreateAlertRule)
			authorized.PUT("/alerts/rules/:id", auth.AdminRequired(), s.handleUpdateAlertRule)
			authorized.DELETE("/alerts/rules/:id", auth.AdminRequired(), s.handleDeleteAlertRule)

			// 通知渠道
			authorized.GET("/notifications/channels", auth.AdminRequired(), s.handleListChannels)
			authorized.POST("/notifications/channels", auth.AdminRequired(), s.handleCreateChannel)
			authorized.POST("/notifications/test", auth.AdminRequired(), s.handleTestNotify)

			// 代理管理（3x-ui）
			authorized.GET("/proxy/status", auth.AdminRequired(), s.handleProxyStatus)
			authorized.POST("/proxy/config", auth.AdminRequired(), s.handleProxyConfig)
			authorized.GET("/proxy/inbounds", auth.AdminRequired(), s.handleProxyInbounds)
			authorized.GET("/proxy/clients", auth.AdminRequired(), s.handleProxyClients)
			authorized.GET("/proxy/nodes", auth.AdminRequired(), s.handleProxyNodes)
			authorized.GET("/proxy/xray/status", auth.AdminRequired(), s.handleProxyXrayStatus)

			// 系统设置
			authorized.GET("/system/info", s.handleSystemInfo)
			authorized.GET("/system/settings", auth.AdminRequired(), s.handleGetSettings)
			authorized.PUT("/system/settings", auth.AdminRequired(), s.handleUpdateSettings)
		}
	}

	// WebSocket
	s.router.GET("/ws/ssh/:id", auth.AuthMiddleware(), s.handleWebSSH)
	s.router.GET("/ws/ssh", s.handleWebSSHDirect)
	s.router.GET("/ws/agent", s.handleAgentWS)

	// 静态文件（前端构建产物）
	s.router.Static("/assets", "./web/dist/assets")
	s.router.StaticFile("/", "./web/dist/index.html")
	s.router.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	s.http = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.Server.WriteTimeout) * time.Second,
	}

	s.logger.Info("Server starting", zap.String("addr", s.http.Addr))
	return s.http.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.scheduler != nil {
		s.scheduler.Stop()
	}
	if s.alertEngine != nil {
		s.alertEngine.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
	}
}
