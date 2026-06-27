package model

import "time"

// ServerMetric 服务器时序指标（TimescaleDB hypertable）
type ServerMetric struct {
	Time          time.Time `gorm:"primaryKey" json:"time"`
	ServerID      uint      `gorm:"primaryKey;index:idx_server_time" json:"server_id"`
	CPUPercent    float64   `gorm:"column:cpu_percent" json:"cpu_percent"`
	MemoryPercent float64   `gorm:"column:memory_percent" json:"memory_percent"`
	DiskPercent   float64   `gorm:"column:disk_percent" json:"disk_percent"`
	Load1         float64   `gorm:"column:load1" json:"load1"`
	Load5         float64   `gorm:"column:load5" json:"load5"`
	Load15        float64   `gorm:"column:load15" json:"load15"`
	NetRx         uint64    `gorm:"column:net_rx" json:"net_rx"`
	NetTx         uint64    `gorm:"column:net_tx" json:"net_tx"`
	Uptime        uint64    `gorm:"column:uptime" json:"uptime"`
	ProcessCount  int       `gorm:"column:process_count" json:"process_count"`
}

// TableName 指定表名
func (ServerMetric) TableName() string {
	return "server_metrics"
}
