package model

import (
	"time"

	"gorm.io/gorm"
)

// AlertRule 告警规则
type AlertRule struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	RuleType    string         `gorm:"size:30;not null" json:"rule_type"` // offline / cpu / memory / disk / traffic / load / port / ssl / container / xui_node / xray
	TargetType  string         `gorm:"size:20;not null" json:"target_type"` // server / group / all
	TargetID    *uint          `json:"target_id"`
	Threshold   *float64       `json:"threshold"`
	Duration    int            `gorm:"default:60" json:"duration"` // 持续秒数触发
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	Channels    string         `gorm:"type:text" json:"channels"` // JSON数组: ["wechat","feishu"]
	UpgradeChannels string     `gorm:"type:text" json:"upgrade_channels"` // 升级通知渠道
	CreatedBy   uint           `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Alert 告警记录
type Alert struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	RuleID      uint       `json:"rule_id"`
	Rule        *AlertRule `json:"rule,omitempty"`
	ServerID    uint       `json:"server_id"`
	Server      *Server    `json:"server,omitempty"`
	Status      string     `gorm:"size:20;default:firing" json:"status"` // firing / resolved
	Message     string     `gorm:"type:text" json:"message"`
	Severity    string     `gorm:"size:20" json:"severity"` // info / warning / critical
	StartedAt   time.Time  `json:"started_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	DurationSec int        `json:"duration_sec"`
	NotifiedAt  *time.Time `json:"notified_at"`
	UpgradedAt  *time.Time `json:"upgraded_at"`
}

// NotificationChannel 通知渠道配置
type NotificationChannel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Channel   string    `gorm:"size:20;not null" json:"channel"` // wechat / qq / email / feishu / telegram
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	Config    string    `gorm:"type:text" json:"config"` // JSON配置
	CreatedAt time.Time `json:"created_at"`
}

// NotificationLog 通知发送日志
type NotificationLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	AlertID   uint      `json:"alert_id"`
	Channel   string    `gorm:"size:20" json:"channel"`
	Recipient string    `gorm:"size:100" json:"recipient"`
	Status    string    `gorm:"size:20" json:"status"` // sent / failed
	ErrorMsg  string    `gorm:"type:text" json:"error_msg"`
	SentAt    time.Time `json:"sent_at"`
}
