package model

import (
	"time"

	"gorm.io/gorm"
)

// Server 服务器模型
type Server struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	GroupID      *uint          `json:"group_id"`
	Group        *ServerGroup   `json:"group,omitempty"`
	OSType       string         `gorm:"size:20" json:"os_type"` // linux / windows
	OSVersion    string         `gorm:"size:100" json:"os_version"`
	CPUModel     string         `gorm:"size:200" json:"cpu_model"`
	CPUCores     int            `json:"cpu_cores"`
	MemoryTotal  uint64         `json:"memory_total"` // MB
	DiskTotal    uint64         `json:"disk_total"`   // GB
	Location     string         `gorm:"size:100" json:"location"`
	IPPublic     string         `gorm:"size:50" json:"ip_public"`
	IPPrivate    string         `gorm:"size:50" json:"ip_private"`
	SSHPort      int            `gorm:"default:22" json:"ssh_port"`
	SSHUser      string         `gorm:"size:50" json:"ssh_user"`
	Status       string         `gorm:"size:20;default:offline" json:"status"` // online / offline
	AgentVersion string         `gorm:"size:20" json:"agent_version"`
	AgentToken   string         `gorm:"size:64;uniqueIndex" json:"agent_token,omitempty"`
	Tags         []ServerTag    `json:"tags,omitempty"`
	Bill         *ServerBill    `json:"bill,omitempty"`
	Hidden       bool           `gorm:"default:false" json:"hidden"`         // 对游客隐藏（默认false）
	PublicNote   string         `gorm:"size:500" json:"public_note"`         // 公开备注（所有用户可见）
	PrivateNote  string         `gorm:"size:500" json:"private_note"`        // 私有备注（仅管理员可见）
	Weight       int            `gorm:"default:0" json:"weight"`             // 排序权重（默认0）
	LastSeenAt   *time.Time     `json:"last_seen_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// ServerGroup 服务器分组
type ServerGroup struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Name       string         `gorm:"size:50;not null" json:"name"`
	ParentID   *uint          `json:"parent_id"`
	SortOrder  int            `gorm:"default:0" json:"sort_order"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// ServerTag 服务器标签
type ServerTag struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	ServerID uint   `json:"server_id"`
	Tag      string `gorm:"size:30;not null" json:"tag"`
}

// ServerBill 服务器账单信息
type ServerBill struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	ServerID    uint       `gorm:"uniqueIndex" json:"server_id"`
	Price       float64    `json:"price"`
	Currency    string     `gorm:"size:10;default:CNY" json:"currency"`
	Cycle       string     `gorm:"size:20" json:"cycle"`                // 月付/季付/半年付/年付/两年付/三年付
	RenewDate   *time.Time `json:"renew_date"`
	ExpiredAt   *time.Time `json:"expired_at"`                          // 到期时间
	AutoRenewal bool       `gorm:"default:true" json:"auto_renewal"`    // 自动续费（默认true）
	BillingType string     `gorm:"size:20" json:"billing_type"`          // 付费类型（预付费/后付费）
	Provider    string     `gorm:"size:100" json:"provider"`
	CreatedAt   time.Time  `json:"created_at"`
}
