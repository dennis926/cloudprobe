package service

import (
	"time"

	"cloudprobe/internal/cache"
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

// GetServerWithMetrics 获取服务器及最新指标（优先从Redis缓存获取）
func (s *ServerService) GetServerWithMetrics(serverID uint) (*model.Server, map[string]interface{}, error) {
	var server model.Server
	if err := s.db.First(&server, serverID).Error; err != nil {
		return nil, nil, err
	}

	// 优先从Redis获取实时指标
	metrics, err := cache.GetServerMetrics(serverID)
	if err != nil || metrics == nil {
		// Redis无数据，返回数据库中的基本状态
		metrics = map[string]interface{}{
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
	}

	return &server, metrics, nil
}

// ListServerMetrics 获取服务器历史指标（调用MetricService）
func (s *ServerService) ListServerMetrics(serverID uint, metricType string, start, end time.Time) ([]map[string]interface{}, error) {
	msvc := NewMetricService()
	return msvc.QueryMetrics(serverID, metricType, start, end)
}

// GetRealtimeMetrics 获取所有服务器实时指标（从Redis）
func (s *ServerService) GetRealtimeMetrics() (map[uint]map[string]interface{}, error) {
	return cache.GetAllServerMetrics()
}

// UpdateServerInfo 更新服务器静态信息
func (s *ServerService) UpdateServerInfo(serverID uint, info map[string]interface{}) error {
	updates := map[string]interface{}{}

	if v, ok := info["hostname"].(string); ok && v != "" {
		updates["name"] = v
	}
	if v, ok := info["os"].(string); ok && v != "" {
		updates["os_type"] = v
	}
	if v, ok := info["platform"].(string); ok && v != "" {
		updates["os_version"] = v
	}

	// CPU 信息
	if cpu, ok := info["cpu"].(map[string]interface{}); ok {
		if model, ok := cpu["model"].(string); ok && model != "" {
			updates["cpu_model"] = model
		}
		if cores, ok := cpu["logical_count"].(float64); ok && cores > 0 {
			updates["cpu_cores"] = int(cores)
		}
	}

	// 内存信息（MB）
	if mem, ok := info["memory"].(map[string]interface{}); ok {
		if total, ok := mem["total"].(float64); ok && total > 0 {
			updates["memory_total"] = uint64(total / 1024 / 1024) // bytes -> MB
		}
	}

	// 磁盘信息（GB，取第一个分区）
	if disks, ok := info["disk"].([]interface{}); ok && len(disks) > 0 {
		if disk, ok := disks[0].(map[string]interface{}); ok {
			if total, ok := disk["total"].(float64); ok && total > 0 {
				updates["disk_total"] = uint64(total / 1024 / 1024 / 1024) // bytes -> GB
			}
		}
	}

	// IP 地址
	if ips, ok := info["ip"].(map[string]interface{}); ok {
		if public, ok := ips["public"].(string); ok && public != "" {
			updates["ip_public"] = public
		}
		if priv, ok := ips["private"].(string); ok && priv != "" {
			updates["ip_private"] = priv
		}
	}

	if len(updates) == 0 {
		return nil
	}
	return s.db.Model(&model.Server{}).Where("id = ?", serverID).Updates(updates).Error
}
