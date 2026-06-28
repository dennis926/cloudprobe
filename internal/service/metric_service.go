package service

import (
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

	// 转换原始嵌套数据为前端需要的扁平化格式
	flattened := map[string]interface{}{
		"timestamp":     data["timestamp"],
		"uptime":        data["uptime"],
		"process_count": data["process_count"],
	}

	// load
	if load, ok := data["load"].(map[string]interface{}); ok {
		flattened["load1"] = load["load1"]
		flattened["load5"] = load["load5"]
		flattened["load15"] = load["load15"]
	}

	// CPU 使用率（取平均值）
	if cpu, ok := data["cpu"].(map[string]interface{}); ok {
		if percents, ok := cpu["percent"].([]interface{}); ok && len(percents) > 0 {
			var total float64
			for _, p := range percents {
				switch v := p.(type) {
				case float64:
					total += v
				case float32:
					total += float64(v)
				}
			}
			flattened["cpu_percent"] = total / float64(len(percents))
		}
		flattened["cpu_cores"] = cpu["logical_count"]
		flattened["cpu_model"] = cpu["model_name"]
	}

	// 内存
	if mem, ok := data["memory"].(map[string]interface{}); ok {
		flattened["memory_total"] = mem["total"]
		flattened["memory_used"] = mem["used"]
		if up, ok := mem["used_percent"].(float64); ok {
			flattened["memory_percent"] = up
		}
	}

	// 磁盘（取第一个分区）
	if disks, ok := data["disk"].([]interface{}); ok && len(disks) > 0 {
		if disk, ok := disks[0].(map[string]interface{}); ok {
			flattened["disk_total"] = disk["total"]
			flattened["disk_used"] = disk["used"]
			if up, ok := disk["used_percent"].(float64); ok {
				flattened["disk_percent"] = up
			}
		}
	}

	// 网络（先存原始累计值）
	if net, ok := data["network"].(map[string]interface{}); ok {
		if bs, ok := net["bytes_sent"].(float64); ok {
			flattened["net_upload_total"] = bs
		}
		if br, ok := net["bytes_recv"].(float64); ok {
			flattened["net_download_total"] = br
		}
	}

	// 计算网络流量速率（需要上一次的数据）
	prev, _ := cache.GetServerMetrics(serverID)
	if prev != nil {
		if prevTs, ok := prev["timestamp"].(float64); ok {
			if currTs, ok := data["timestamp"].(float64); ok {
				interval := currTs - prevTs
				if interval > 0 {
					if prevUp, ok := prev["net_upload_total"].(float64); ok {
						if currUp, ok := flattened["net_upload_total"].(float64); ok {
							diff := currUp - prevUp
							if diff >= 0 {
								flattened["net_upload"] = diff / interval
							}
						}
					}
					if prevDown, ok := prev["net_download_total"].(float64); ok {
						if currDown, ok := flattened["net_download_total"].(float64); ok {
							diff := currDown - prevDown
							if diff >= 0 {
								flattened["net_download"] = diff / interval
							}
						}
					}
				}
			}
		}
	}

	// 同时写入Redis缓存（保留最新数据）
	if err := cache.SetServerMetrics(serverID, flattened, 5*time.Minute); err != nil {
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
