package task

import (
	"time"

	"cloudprobe/internal/backup"
	"cloudprobe/internal/config"

	"go.uber.org/zap"
)

// RunBackup 执行数据备份并发送到邮箱
func RunBackup(logger *zap.Logger) {
	logger.Info("running task: data backup")

	cfg := config.Get()
	if cfg.Backup.Email == "" {
		logger.Warn("backup email not configured, skipping")
		return
	}

	exporter := backup.NewExporter()
	mailer := backup.NewMailerFromConfig()

	// 导出最近90天的指标数据
	end := time.Now()
	start := end.AddDate(0, 0, -90)

	files := []string{}

	if metricsFile, err := exporter.ExportMetricsToCSV(start, end); err != nil {
		logger.Error("export metrics failed", zap.Error(err))
	} else {
		files = append(files, metricsFile)
		logger.Info("metrics exported", zap.String("file", metricsFile))
	}

	if serversFile, err := exporter.ExportServersToCSV(); err != nil {
		logger.Error("export servers failed", zap.Error(err))
	} else {
		files = append(files, serversFile)
		logger.Info("servers exported", zap.String("file", serversFile))
	}

	if len(files) > 0 {
		if err := mailer.SendBackupEmail(cfg.Backup.Email, files); err != nil {
			logger.Error("send backup email failed", zap.Error(err))
		} else {
			logger.Info("backup email sent", zap.String("to", cfg.Backup.Email))
		}
	}

	// 清理旧备份文件
	if err := backup.CleanupOldBackups(); err != nil {
		logger.Error("cleanup old backups failed", zap.Error(err))
	}
}
