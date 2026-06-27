package notify

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Channel 通知渠道接口
type Channel interface {
	Name() string
	Send(ctx context.Context, title, content string, recipient string) error
	Test(ctx context.Context) error
}

// Manager 通知管理器
type Manager struct {
	channels map[string]Channel
	logger   *zap.Logger
}

// NewManager 创建通知管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		channels: make(map[string]Channel),
		logger:   logger,
	}
}

// Register 注册通知渠道
func (m *Manager) Register(ch Channel) {
	m.channels[ch.Name()] = ch
	m.logger.Info("notification channel registered", zap.String("channel", ch.Name()))
}

// Get 获取指定渠道
func (m *Manager) Get(name string) (Channel, bool) {
	ch, ok := m.channels[name]
	return ch, ok
}

// Send 发送通知到指定渠道
func (m *Manager) Send(ctx context.Context, channelName, title, content, recipient string) error {
	ch, ok := m.Get(channelName)
	if !ok {
		return fmt.Errorf("channel %s not found", channelName)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	start := time.Now()
	err := ch.Send(ctx, title, content, recipient)
	duration := time.Since(start)

	if err != nil {
		m.logger.Error("notification send failed",
			zap.String("channel", channelName),
			zap.String("recipient", recipient),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	m.logger.Info("notification sent",
		zap.String("channel", channelName),
		zap.String("recipient", recipient),
		zap.Duration("duration", duration),
	)
	return nil
}

// Broadcast 广播通知到多个渠道
func (m *Manager) Broadcast(ctx context.Context, channelNames []string, title, content string, recipients map[string]string) map[string]error {
	results := make(map[string]error)
	for _, name := range channelNames {
		recipient := ""
		if recipients != nil {
			recipient = recipients[name]
		}
		if err := m.Send(ctx, name, title, content, recipient); err != nil {
			results[name] = err
		} else {
			results[name] = nil
		}
	}
	return results
}
