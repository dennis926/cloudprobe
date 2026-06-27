package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloudprobe/internal/cache"
	"cloudprobe/internal/database"
	"cloudprobe/internal/model"

	"gorm.io/gorm"
)

// MetricService 指标服务
type MetricService struct {
	db *gorm.DB
}

// NewMetricService 创建指标服务
func NewMetricService() *MetricService {
	return &MetricService{db: database.GetDB()}
}

// SaveMetrics 保存指标到TimescaleDB
func (s *MetricService) SaveMetrics(serverID uint, data map[string]interface{}) error {
	metric := &model.ServerMetric{
		Time:     time.Now(),
		ServerID: serverID,
	}

	if v, ok := data["cpu"].(map[string]interface{}); ok {
		if percents, ok := v["percent"].([]interface{}); ok && len(percents) > 0 {
			var sum float64
			for _, p := range percents {
				if f, ok := p.(float64); ok {
					sum += f
				}
			}
			metric.CPUPercent = sum / float64(len(percents))
		}
	}

	if v, ok := data["memory"].(map[string]interface{}); ok {
		if p, ok := v["used_percent"].(float64); ok {
			metric.MemoryPercent = p
		}
	}

	if disks, ok := data["disk"].([]interface{}); ok && len(disks) > 0 {
		var maxPercent float64
		for _, d := range disks {
			if dm, ok := d.(map[string]interface{}); ok {
				if p, ok := dm["used_percent"].(float64); ok && p > maxPercent {
					maxPercent = p
				}
			}
		}
		metric.DiskPercent = maxPercent
	}

	if v, ok := data["load"].(map[string]interface{}); ok {
		if f, ok := v["load1"].(float64); ok {
			metric.Load1 = f
		}
		if f, ok := v["load5"].(float64); ok {
			metric.Load5 = f
		}
		if f, ok := v["load15"].(float64); ok {
			metric.Load15 = f
		}
	}

	if v, ok := data["network"].(map[string]interface{}); ok {
		if f, ok := v["bytes_recv"].(float64); ok {
			metric.NetRx = uint64(f)
		}
		if f, ok := v["bytes_sent"].(float64); ok {
			metric.NetTx = uint64(f)
		}
	}

	if f, ok := data["uptime"].(float64); ok {
		metric.Uptime = uint64(f)
	}
	if f, ok := data["process_count"].(float64); ok {
		metric.ProcessCount = int(f)
	}

	// 写入TimescaleDB
	if err := s.db.Create(metric).Error; err != nil {
		return fmt.Errorf("save metric failed: %w", err)
	}

	// 同时写入Redis缓存（保留最新数据）
	if err := cache.SetServerMetrics(serverID, data, 5*time.Minute); err != nil {
		// Redis失败不影响主流程
	}
	if err := cache.SetServerStatus(serverID, "online", 2*time.Minute); err != nil {
		// Redis失败不影响主流程
	}
	if err := cache.SetServerHeartbeat(serverID); err != nil {
		// Redis失败不影响主流程
	}

	return nil
}

// QueryMetrics 查询时序指标
func (s *MetricService) QueryMetrics(serverID uint, metricType string, start, end time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 使用原始SQL查询以充分利用TimescaleDB的时序优化
	sql := `
		SELECT time_bucket('1 minute', time) as bucket,
		       AVG(cpu_percent) as cpu_avg,
		       AVG(memory_percent) as memory_avg,
		       AVG(disk_percent) as disk_avg,
		       AVG(load1) as load_avg,
		       MAX(net_rx) as net_rx_max,
		       MAX(net_tx) as net_tx_max
		FROM server_metrics
		WHERE server_id = ? AND time >= ? AND time <= ?
		GROUP BY bucket
		ORDER BY bucket DESC
		LIMIT 1440
	`

	rows, err := s.db.Raw(sql, serverID, start, end).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bucket time.Time
		var cpuAvg, memAvg, diskAvg, loadAvg float64
		var netRxMax, netTxMax uint64
		if err := rows.Scan(&bucket, &cpuAvg, &memAvg, &diskAvg, &loadAvg, &netRxMax, &netTxMax); err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"time":            bucket,
			"cpu_percent":     cpuAvg,
			"memory_percent":  memAvg,
			"disk_percent":    diskAvg,
			"load1":           loadAvg,
			"net_rx":          netRxMax,
			"net_tx":          netTxMax,
		})
	}

	return results, nil
}

// GetLatestMetrics 获取所有服务器最新指标
func (s *MetricService) GetLatestMetrics() (map[uint]map[string]interface{}, error) {
	// 优先从Redis获取实时数据
	return cache.GetAllServerMetrics()
}

// CleanupOldMetrics 清理90天前的数据（备用方案，TimescaleDB retention policy 为主）
func (s *MetricService) CleanupOldMetrics(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return s.db.Where("time < ?", cutoff).Delete(&model.ServerMetric{}).Error
}

// handleMetricsFromAgent 处理Agent上报的指标（内部使用）
func HandleMetricsFromAgent(serverID uint, data map[string]interface{}) error {
	svc := NewMetricService()
	return svc.SaveMetrics(serverID, data)
}
