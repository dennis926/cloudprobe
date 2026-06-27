package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Email     string         `gorm:"size:100" json:"email"`
	Role      string         `gorm:"size:20;default:viewer" json:"role"` // admin / viewer
	Status    string         `gorm:"size:20;default:active" json:"status"` // active / disabled
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// APIToken API令牌模型
type APIToken struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Token     string    `gorm:"size:64;uniqueIndex;not null" json:"-"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `json:"user_id"`
	Action    string    `gorm:"size:50" json:"action"`
	Resource  string    `gorm:"size:100" json:"resource"`
	Detail    string    `gorm:"type:text" json:"detail"`
	IP        string    `gorm:"size:50" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}
