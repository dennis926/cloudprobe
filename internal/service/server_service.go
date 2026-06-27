package service

import (
	"time"

	"cloudprobe/internal/database"
	"cloudprobe/internal/model"

	"gorm.io/gorm"
)

// ServerService 服务器业务服务
type ServerService struct {
	db *gorm.DB
}

// NewServerService 创建服务器服务
func NewServerService() *ServerService {
	return &ServerService{db: database.GetDB()}
}

// GetServerStatus 获取服务器聚合状态
func (s *ServerService) GetServerStatus() (total, online, offline int64, err error) {
	if err = s.db.Model(&model.Server{}).Count(&total).Error; err != nil {
		return
	}
	if err = s.db.Model(&model.Server{}).Where("status = ?", "online").Count(&online).Error; err != nil {
		return
	}
	if err = s.db.Model(&model.Server{}).Where("status = ?", "offline").Count(&offline).Error; err != nil {
		return
	}
	return
}

// UpdateServerStatus 更新服务器在线状态
func (s *ServerService) UpdateServerStatus(serverID uint, status string) error {
	updates := map[string]interface{}{
		"status":       status,
		"last_seen_at": time.Now(),
	}
	return s.db.Model(&model.Server{}).Where("id = ?", serverID).Updates(updates).Error
}

// MarkOfflineServers 标记超时未上报的服务器为离线
func (s *ServerService) MarkOfflineServers(timeout time.Duration) error {
	threshold := time.Now().Add(-timeout)
	return s.db.Model(&model.Server{}).
		Where("last_seen_at < ? AND status = ?", threshold, "online").
		Updates(map[string]interface{}{
			"status": "offline",
		}).Error
}

// GetServerWithMetrics 获取服务器及最新指标
func (s *ServerService) GetServerWithMetrics(serverID uint) (*model.Server, map[string]interface{}, error) {
	var server model.Server
	if err := s.db.Preload("Group").Preload("Tags").First(&server, serverID).Error; err != nil {
		return nil, nil, err
	}

	// TODO: 从TimescaleDB获取最新指标数据
	metrics := map[string]interface{}{
		"cpu_percent":    0.0,
		"memory_percent": 0.0,
		"disk_percent":   0.0,
		"load_1":         0.0,
		"load_5":         0.0,
		"load_15":        0.0,
		"net_rx":         0,
		"net_tx":         0,
		"uptime":         0,
	}

	return &server, metrics, nil
}

// ListServerMetrics 获取服务器历史指标（简化版）
func (s *ServerService) ListServerMetrics(serverID uint, metricType string, start, end time.Time) ([]map[string]interface{}, error) {
	// TODO: 从TimescaleDB查询时序数据
	// 临时返回空数组，后续接入实际数据
	return []map[string]interface{}{}, nil
}

// GetRealtimeMetrics 获取所有服务器实时指标（从Redis）
func (s *ServerService) GetRealtimeMetrics() (map[uint]map[string]interface{}, error) {
	// TODO: 从Redis获取所有服务器实时状态
	return map[uint]map[string]interface{}{}, nil
}
