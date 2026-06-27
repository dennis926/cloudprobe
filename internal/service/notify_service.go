package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"cloudprobe/internal/database"
	"cloudprobe/internal/model"
	"cloudprobe/internal/notify"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NotifyService 通知服务
type NotifyService struct {
	db      *gorm.DB
	manager *notify.Manager
	logger  *zap.Logger
	mu      sync.RWMutex
}

// GlobalNotifyService 全局通知服务实例
var GlobalNotifyService *NotifyService

// NewNotifyService 创建通知服务
func NewNotifyService(db *gorm.DB, logger *zap.Logger) *NotifyService {
	return &NotifyService{
		db:      db,
		manager: notify.NewManager(logger),
		logger:  logger,
	}
}

// InitNotifyService 初始化并加载所有通知渠道
func InitNotifyService(logger *zap.Logger) *NotifyService {
	GlobalNotifyService = NewNotifyService(database.GetDB(), logger)
	if err := GlobalNotifyService.ReloadChannels(); err != nil {
		logger.Error("failed to load notification channels", zap.Error(err))
	}
	return GlobalNotifyService
}

// GetManager 获取通知管理器
func (s *NotifyService) GetManager() *notify.Manager {
	return s.manager
}

// ReloadChannels 从数据库重新加载所有启用的通知渠道
func (s *NotifyService) ReloadChannels() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var channels []model.NotificationChannel
	if err := s.db.Where("enabled = ?", true).Find(&channels).Error; err != nil {
		return err
	}

	// 重新注册所有渠道
	for _, ch := range channels {
		if err := s.registerChannel(&ch); err != nil {
			s.logger.Warn("failed to register channel",
				zap.String("name", ch.Name),
				zap.String("channel", ch.Channel),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("notification channels reloaded", zap.Int("count", len(channels)))
	return nil
}

// registerChannel 注册单个渠道
func (s *NotifyService) registerChannel(ch *model.NotificationChannel) error {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ch.Config), &config); err != nil {
		return fmt.Errorf("parse config failed: %w", err)
	}

	switch ch.Channel {
	case "email":
		channel := &notify.EmailChannel{
			SMTPHost:     getString(config, "smtp_host"),
			SMTPPort:     getInt(config, "smtp_port"),
			SMTPUser:     getString(config, "smtp_user"),
			SMTPPassword: getString(config, "smtp_password"),
			From:         getString(config, "from"),
			UseTLS:       getBool(config, "use_tls"),
			UseSSL:       getBool(config, "use_ssl"),
		}
		s.manager.Register(channel)

	case "wechat":
		channel := &notify.WechatChannel{
			Provider: getString(config, "provider"),
			Token:    getString(config, "token"),
		}
		s.manager.Register(channel)

	case "feishu":
		channel := &notify.FeishuChannel{
			WebhookURL: getString(config, "webhook_url"),
		}
		s.manager.Register(channel)

	case "telegram":
		channel := &notify.TelegramChannel{
			BotToken: getString(config, "bot_token"),
			ChatID:   getString(config, "chat_id"),
		}
		s.manager.Register(channel)

	case "qq":
		channel := &notify.QQChannel{
			APIBaseURL:  getString(config, "api_base_url"),
			AccessToken: getString(config, "access_token"),
			DefaultQQ:   getString(config, "default_qq"),
		}
		s.manager.Register(channel)

	default:
		return fmt.Errorf("unsupported channel type: %s", ch.Channel)
	}

	return nil
}

// TestChannel 测试指定渠道
func (s *NotifyService) TestChannel(ctx context.Context, channelID uint) error {
	var ch model.NotificationChannel
	if err := s.db.First(&ch, channelID).Error; err != nil {
		return err
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ch.Config), &config); err != nil {
		return err
	}

	// 临时创建一个渠道进行测试
	tempMgr := notify.NewManager(s.logger)
	switch ch.Channel {
	case "email":
		tempMgr.Register(&notify.EmailChannel{
			SMTPHost:     getString(config, "smtp_host"),
			SMTPPort:     getInt(config, "smtp_port"),
			SMTPUser:     getString(config, "smtp_user"),
			SMTPPassword: getString(config, "smtp_password"),
			From:         getString(config, "from"),
			UseTLS:       getBool(config, "use_tls"),
			UseSSL:       getBool(config, "use_ssl"),
		})
	case "wechat":
		tempMgr.Register(&notify.WechatChannel{
			Provider: getString(config, "provider"),
			Token:    getString(config, "token"),
		})
	case "feishu":
		tempMgr.Register(&notify.FeishuChannel{
			WebhookURL: getString(config, "webhook_url"),
		})
	case "telegram":
		tempMgr.Register(&notify.TelegramChannel{
			BotToken: getString(config, "bot_token"),
			ChatID:   getString(config, "chat_id"),
		})
	case "qq":
		tempMgr.Register(&notify.QQChannel{
			APIBaseURL:  getString(config, "api_base_url"),
			AccessToken: getString(config, "access_token"),
			DefaultQQ:   getString(config, "default_qq"),
		})
	}

	if chTest, ok := tempMgr.Get(ch.Channel); ok {
		return chTest.Test(ctx)
	}
	return fmt.Errorf("channel not registered for test")
}

// Helper functions for config parsing
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case int64:
			return int(val)
		}
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
