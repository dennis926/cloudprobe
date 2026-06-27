package backup

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cloudprobe/internal/database"
	"cloudprobe/internal/model"

	"gorm.io/gorm"
)

// Exporter 数据导出器
type Exporter struct {
	db *gorm.DB
}

// NewExporter 创建导出器
func NewExporter() *Exporter {
	return &Exporter{db: database.GetDB()}
}

// ExportMetricsToCSV 导出指标数据到CSV
func (e *Exporter) ExportMetricsToCSV(start, end time.Time) (string, error) {
	var metrics []model.ServerMetric
	if err := e.db.Where("time >= ? AND time <= ?", start, end).
		Order("time DESC").
		Find(&metrics).Error; err != nil {
		return "", fmt.Errorf("query metrics failed: %w", err)
	}

	// 创建临时目录
	tmpDir := os.TempDir()
	filename := fmt.Sprintf("cloudprobe_metrics_%s_%s.csv",
		start.Format("20060102"), end.Format("20060102"))
	filepath := filepath.Join(tmpDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"time", "server_id", "cpu_percent", "memory_percent", "disk_percent",
		"load1", "load5", "load15", "net_rx", "net_tx", "uptime", "process_count"}
	if err := writer.Write(headers); err != nil {
		return "", err
	}

	// 写入数据
	for _, m := range metrics {
		record := []string{
			m.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%d", m.ServerID),
			fmt.Sprintf("%.2f", m.CPUPercent),
			fmt.Sprintf("%.2f", m.MemoryPercent),
			fmt.Sprintf("%.2f", m.DiskPercent),
			fmt.Sprintf("%.2f", m.Load1),
			fmt.Sprintf("%.2f", m.Load5),
			fmt.Sprintf("%.2f", m.Load15),
			fmt.Sprintf("%d", m.NetRx),
			fmt.Sprintf("%d", m.NetTx),
			fmt.Sprintf("%d", m.Uptime),
			fmt.Sprintf("%d", m.ProcessCount),
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	return filepath, nil
}

// ExportServersToCSV 导出服务器列表到CSV
func (e *Exporter) ExportServersToCSV() (string, error) {
	var servers []model.Server
	if err := e.db.Find(&servers).Error; err != nil {
		return "", fmt.Errorf("query servers failed: %w", err)
	}

	tmpDir := os.TempDir()
	filename := fmt.Sprintf("cloudprobe_servers_%s.csv", time.Now().Format("20060102"))
	filepath := filepath.Join(tmpDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"id", "name", "public_ip", "os_type", "status", "cpu_info",
		"memory_total", "disk_total", "location", "ssh_port", "ssh_user", "created_at"}
	writer.Write(headers)

	for _, s := range servers {
		record := []string{
			fmt.Sprintf("%d", s.ID), s.Name, s.PublicIP, s.OSType, s.Status,
			s.CPUInfo, fmt.Sprintf("%d", s.MemoryTotal), fmt.Sprintf("%d", s.DiskTotal),
			s.Location, fmt.Sprintf("%d", s.SSHPort), s.SSHUser,
			s.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		writer.Write(record)
	}

	return filepath, nil
}

// CleanupOldBackups 清理7天前的临时备份文件
func CleanupOldBackups() error {
	tmpDir := os.TempDir()
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -7)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !((len(name) > 17 && name[:17] == "cloudprobe_metrics_") ||
			(len(name) > 17 && name[:17] == "cloudprobe_servers_")) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(tmpDir, name))
		}
	}
	return nil
}
