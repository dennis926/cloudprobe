package task

import (
	"context"
	"time"

	"cloudprobe/internal/service"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron   *cron.Cron
	logger *zap.Logger
}

// NewScheduler 创建调度器
func NewScheduler(logger *zap.Logger) *Scheduler {
	return &Scheduler{
		cron:   cron.New(),
		logger: logger,
	}
}

// Start 启动所有定时任务
func (s *Scheduler) Start() {
	// 每10分钟清理一次离线服务器状态
	s.cron.AddFunc("*/10 * * * *", func() {
		s.logger.Info("running task: mark offline servers")
		svc := service.NewServerService()
		if err := svc.MarkOfflineServers(2 * time.Minute); err != nil {
			s.logger.Error("mark offline servers failed", zap.Error(err))
		}
	})

	// 每天凌晨2点清理90天前的指标数据
	s.cron.AddFunc("0 2 * * *", func() {
		s.logger.Info("running task: cleanup old metrics")
		svc := service.NewMetricService()
		if err := svc.CleanupOldMetrics(90); err != nil {
			s.logger.Error("cleanup old metrics failed", zap.Error(err))
		} else {
			s.logger.Info("cleanup old metrics completed")
		}
	})

	// 每小时重新加载通知渠道配置
	s.cron.AddFunc("0 * * * *", func() {
		s.logger.Info("running task: reload notification channels")
		if service.GlobalNotifyService != nil {
			if err := service.GlobalNotifyService.ReloadChannels(); err != nil {
				s.logger.Error("reload channels failed", zap.Error(err))
			}
		}
	})

	// 每周日凌晨3点执行数据备份
	s.cron.AddFunc("0 3 * * 0", func() {
		RunBackup(s.logger)
	})

	s.cron.Start()
	s.logger.Info("task scheduler started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("task scheduler stopped")
}
