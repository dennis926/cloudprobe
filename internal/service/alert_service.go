package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cloudprobe/internal/database"
	"cloudprobe/internal/model"
	"cloudprobe/internal/notify"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertEngine 告警引擎
type AlertEngine struct {
	db      *gorm.DB
	notify  *notify.Manager
	logger  *zap.Logger
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
	mu      sync.RWMutex
}

// NewAlertEngine 创建告警引擎
func NewAlertEngine(db *gorm.DB, notifyMgr *notify.Manager, logger *zap.Logger) *AlertEngine {
	return &AlertEngine{
		db:     db,
		notify: notifyMgr,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// Start 启动告警引擎
func (e *AlertEngine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		return
	}
	e.running = true

	e.wg.Add(1)
	go e.loop()
	e.logger.Info("alert engine started")
}

// Stop 停止告警引擎
func (e *AlertEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.running {
		return
	}
	e.running = false
	close(e.stopCh)
	e.wg.Wait()
	e.logger.Info("alert engine stopped")
}

// loop 主循环，每30秒检查一次
func (e *AlertEngine) loop() {
	defer e.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	e.checkAll()

	for {
		select {
		case <-ticker.C:
			e.checkAll()
		case <-e.stopCh:
			return
		}
	}
}

// checkAll 检查所有告警规则
func (e *AlertEngine) checkAll() {
	var rules []model.AlertRule
	if err := e.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		e.logger.Error("failed to load alert rules", zap.Error(err))
		return
	}

	for _, rule := range rules {
		if err := e.checkRule(&rule); err != nil {
			e.logger.Error("check rule failed",
				zap.Uint("rule_id", rule.ID),
				zap.String("rule_name", rule.Name),
				zap.Error(err),
			)
		}
	}
}

// checkRule 检查单条规则
func (e *AlertEngine) checkRule(rule *model.AlertRule) error {
	switch rule.RuleType {
	case "offline":
		return e.checkOffline(rule)
	case "cpu", "memory", "disk", "load":
		return e.checkThreshold(rule)
	default:
		return fmt.Errorf("unsupported rule type: %s", rule.RuleType)
	}
}

// checkOffline 检查服务器离线状态
func (e *AlertEngine) checkOffline(rule *model.AlertRule) error {
	servers, err := e.getTargetServers(rule)
	if err != nil {
		return err
	}

	now := time.Now()
	threshold := time.Duration(rule.Duration) * time.Second

	for _, server := range servers {
		var offlineDuration time.Duration
		if server.LastSeenAt != nil {
			offlineDuration = now.Sub(*server.LastSeenAt)
		} else {
			offlineDuration = now.Sub(server.CreatedAt)
		}
		isOffline := offlineDuration > threshold

		alert, err := e.getActiveAlert(rule.ID, server.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if isOffline {
			if alert == nil {
				// 新触发告警
				if err := e.createAlert(rule, server, offlineDuration); err != nil {
					return err
				}
			} else {
				// 更新持续时间，检查是否需要升级
				if err := e.maybeUpgradeAlert(alert, rule, offlineDuration); err != nil {
					return err
				}
			}
		} else {
			if alert != nil {
				// 告警恢复
				if err := e.resolveAlert(alert); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// checkThreshold 检查阈值类告警
func (e *AlertEngine) checkThreshold(rule *model.AlertRule) error {
	if rule.Threshold == nil {
		return nil
	}

	servers, err := e.getTargetServers(rule)
	if err != nil {
		return err
	}

	// 获取最新指标（简化版：从数据库最新一条记录获取）
	for _, server := range servers {
		value, err := e.getLatestMetric(server.ID, rule.RuleType)
		if err != nil {
			continue
		}

		isTriggered := value > *rule.Threshold
		alert, err := e.getActiveAlert(rule.ID, server.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			continue
		}

		if isTriggered {
			if alert == nil {
				// 首次触发，检查持续时间
				if err := e.createThresholdAlert(rule, server, value); err != nil {
					e.logger.Error("create threshold alert failed", zap.Error(err))
				}
			}
		} else {
			if alert != nil {
				if err := e.resolveAlert(alert); err != nil {
					e.logger.Error("resolve alert failed", zap.Error(err))
				}
			}
		}
	}
	return nil
}

// getTargetServers 获取规则目标服务器列表
func (e *AlertEngine) getTargetServers(rule *model.AlertRule) ([]model.Server, error) {
	var servers []model.Server
	switch rule.TargetType {
	case "all":
		if err := e.db.Find(&servers).Error; err != nil {
			return nil, err
		}
	case "server":
		if rule.TargetID == nil {
			return nil, fmt.Errorf("target_id is nil for server target")
		}
		var server model.Server
		if err := e.db.First(&server, *rule.TargetID).Error; err != nil {
			return nil, err
		}
		servers = append(servers, server)
	case "group":
		if rule.TargetID == nil {
			return nil, fmt.Errorf("target_id is nil for group target")
		}
		if err := e.db.Where("group_id = ?", *rule.TargetID).Find(&servers).Error; err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown target type: %s", rule.TargetType)
	}
	return servers, nil
}

// getActiveAlert 获取当前活跃告警
func (e *AlertEngine) getActiveAlert(ruleID, serverID uint) (*model.Alert, error) {
	var alert model.Alert
	err := e.db.Where("rule_id = ? AND server_id = ? AND status = ?", ruleID, serverID, "firing").
		Order("started_at DESC").
		First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// createAlert 创建告警记录并发送首次通知
func (e *AlertEngine) createAlert(rule *model.AlertRule, server model.Server, duration time.Duration) error {
	alert := model.Alert{
		RuleID:      rule.ID,
		ServerID:    server.ID,
		Status:      "firing",
		Message:     fmt.Sprintf("服务器 %s 已离线 %s", server.Name, formatDuration(duration)),
		Severity:    "warning",
		StartedAt:   time.Now(),
		DurationSec: int(duration.Seconds()),
	}

	if err := e.db.Create(&alert).Error; err != nil {
		return err
	}

	e.logger.Info("alert created",
		zap.Uint("alert_id", alert.ID),
		zap.String("server", server.Name),
		zap.String("rule", rule.Name),
	)

	// 发送首次通知（1分钟触发）
	if err := e.sendNotification(rule, &alert, server, false); err != nil {
		e.logger.Error("initial notification failed", zap.Error(err))
	}

	return nil
}

// createThresholdAlert 创建阈值告警
func (e *AlertEngine) createThresholdAlert(rule *model.AlertRule, server model.Server, value float64) error {
	alert := model.Alert{
		RuleID:      rule.ID,
		ServerID:    server.ID,
		Status:      "firing",
		Message:     fmt.Sprintf("服务器 %s %s 使用率 %.1f%%，超过阈值 %.1f%%", server.Name, rule.RuleType, value, *rule.Threshold),
		Severity:    "warning",
		StartedAt:   time.Now(),
		DurationSec: 0,
	}

	if err := e.db.Create(&alert).Error; err != nil {
		return err
	}

	if err := e.sendNotification(rule, &alert, server, false); err != nil {
		e.logger.Error("threshold notification failed", zap.Error(err))
	}
	return nil
}

// maybeUpgradeAlert 检查告警是否需要升级
func (e *AlertEngine) maybeUpgradeAlert(alert *model.Alert, rule *model.AlertRule, duration time.Duration) error {
	alert.DurationSec = int(duration.Seconds())

	// 更新持续时间
	if err := e.db.Model(alert).Update("duration_sec", alert.DurationSec).Error; err != nil {
		return err
	}

	// 检查是否需要升级通知（5分钟 = 300秒）
	if alert.UpgradedAt == nil && duration >= 5*time.Minute {
		now := time.Now()
		alert.UpgradedAt = &now
		alert.Severity = "critical"

		if err := e.db.Model(alert).
			Updates(map[string]interface{}{
				"upgraded_at": &now,
				"severity":     "critical",
			}).Error; err != nil {
			return err
		}

		var server model.Server
		if err := e.db.First(&server, alert.ServerID).Error; err != nil {
			return err
		}

		// 发送升级通知
		if err := e.sendNotification(rule, alert, server, true); err != nil {
			e.logger.Error("upgrade notification failed", zap.Error(err))
		}
	}
	return nil
}

// resolveAlert 告警恢复
func (e *AlertEngine) resolveAlert(alert *model.Alert) error {
	now := time.Now()
	alert.Status = "resolved"
	alert.ResolvedAt = &now

	if err := e.db.Model(alert).
		Updates(map[string]interface{}{
			"status":       "resolved",
			"resolved_at":  &now,
		}).Error; err != nil {
		return err
	}

	e.logger.Info("alert resolved",
		zap.Uint("alert_id", alert.ID),
		zap.Int("duration_sec", alert.DurationSec),
	)
	return nil
}

// sendNotification 发送通知
func (e *AlertEngine) sendNotification(rule *model.AlertRule, alert *model.Alert, server model.Server, isUpgrade bool) error {
	var channelNames []string
	if err := json.Unmarshal([]byte(rule.Channels), &channelNames); err != nil {
		return fmt.Errorf("parse channels failed: %w", err)
	}

	if isUpgrade && rule.UpgradeChannels != "" {
		var upgradeChannels []string
		if err := json.Unmarshal([]byte(rule.UpgradeChannels), &upgradeChannels); err == nil && len(upgradeChannels) > 0 {
			channelNames = upgradeChannels
		}
	}

	if len(channelNames) == 0 {
		return nil
	}

	title := fmt.Sprintf("【CloudProbe告警】%s", rule.Name)
	if isUpgrade {
		title = fmt.Sprintf("【CloudProbe告警升级】%s", rule.Name)
	}

	content := alert.Message
	if isUpgrade {
		content += fmt.Sprintf("\n\n⚠️ 告警已持续 %s，请立即处理！", formatDuration(time.Duration(alert.DurationSec)*time.Second))
	}

	content += fmt.Sprintf("\n\n服务器: %s\nIP: %s\n时间: %s",
		server.Name, server.IPPublic, time.Now().Format("2006-01-02 15:04:05"))

	ctx := context.Background()
	results := e.notify.Broadcast(ctx, channelNames, title, content, nil)

	// 记录通知日志
	for chName, err := range results {
		status := "success"
		errMsg := ""
		if err != nil {
			status = "failed"
			errMsg = err.Error()
		}
		log := model.NotificationLog{
			AlertID:  alert.ID,
			Channel:  chName,
			Status:   status,
			ErrorMsg: errMsg,
			SentAt:   time.Now(),
		}
		e.db.Create(&log)
	}

	// 更新通知时间
	now := time.Now()
	if isUpgrade {
		alert.UpgradedAt = &now
	} else {
		alert.NotifiedAt = &now
	}
	return e.db.Model(alert).Save(alert).Error
}

// getLatestMetric 从TimescaleDB获取最新指标
func (e *AlertEngine) getLatestMetric(serverID uint, metricType string) (float64, error) {
	var result struct {
		Value float64
	}

	var column string
	switch metricType {
	case "cpu":
		column = "cpu_percent"
	case "memory":
		column = "memory_percent"
	case "disk":
		column = "disk_percent"
	case "load":
		column = "load1"
	default:
		return 0, fmt.Errorf("unknown metric type: %s", metricType)
	}

	query := fmt.Sprintf(`SELECT %s as value FROM server_metrics WHERE server_id = ? ORDER BY time DESC LIMIT 1`, column)
	if err := e.db.Raw(query, serverID).Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Value, nil
}

// formatDuration 格式化持续时间
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d分%d秒", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%d小时%d分", int(d.Hours()), int(d.Minutes())%60)
}

// GetAlertStats 获取告警统计
func (e *AlertEngine) GetAlertStats() (map[string]int64, error) {
	var stats struct {
		Firing   int64
		Resolved int64
		Total24h int64
	}

	now := time.Now()
	startOfDay := now.Truncate(24 * time.Hour)

	e.db.Model(&model.Alert{}).Where("status = ?", "firing").Count(&stats.Firing)
	e.db.Model(&model.Alert{}).Where("status = ?", "resolved").Count(&stats.Resolved)
	e.db.Model(&model.Alert{}).Where("created_at >= ?", startOfDay).Count(&stats.Total24h)

	return map[string]int64{
		"firing":    stats.Firing,
		"resolved":  stats.Resolved,
		"total_24h": stats.Total24h,
	}, nil
}

// GlobalAlertEngine 全局告警引擎实例
var GlobalAlertEngine *AlertEngine

// InitAlertEngine 初始化告警引擎
func InitAlertEngine(logger *zap.Logger) *AlertEngine {
	GlobalAlertEngine = NewAlertEngine(database.GetDB(), GlobalNotifyService.GetManager(), logger)
	return GlobalAlertEngine
}
