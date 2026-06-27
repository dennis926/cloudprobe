package database

import (
	"fmt"
	"log"
)

// InitTimescaleDB 初始化TimescaleDB扩展和hypertable
func InitTimescaleDB() error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB failed: %w", err)
	}

	// 启用 TimescaleDB 扩展
	if _, err := sqlDB.Exec(`CREATE EXTENSION IF NOT EXISTS timescaledb`); err != nil {
		log.Printf("Warning: timescaledb extension may not be available: %v", err)
		// 如果扩展不可用，普通PostgreSQL也能运行，只是没有超表优化
	}

	// 创建时序指标表（如果不存在）
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS server_metrics (
		time TIMESTAMPTZ NOT NULL,
		server_id BIGINT NOT NULL,
		cpu_percent DOUBLE PRECISION DEFAULT 0,
		memory_percent DOUBLE PRECISION DEFAULT 0,
		disk_percent DOUBLE PRECISION DEFAULT 0,
		load1 DOUBLE PRECISION DEFAULT 0,
		load5 DOUBLE PRECISION DEFAULT 0,
		load15 DOUBLE PRECISION DEFAULT 0,
		net_rx BIGINT DEFAULT 0,
		net_tx BIGINT DEFAULT 0,
		uptime BIGINT DEFAULT 0,
		process_count INT DEFAULT 0
	);
	`
	if _, err := sqlDB.Exec(createTableSQL); err != nil {
		return fmt.Errorf("create metrics table failed: %w", err)
	}

	// 转换为 hypertable（TimescaleDB 特有）
	// 如果扩展未安装，此语句会失败，但不影响基础功能
	if _, err := sqlDB.Exec(`
		SELECT create_hypertable('server_metrics', 'time', 
			if_not_exists => TRUE,
			chunk_time_interval => INTERVAL '1 hour'
		)
	`); err != nil {
		log.Printf("Warning: create_hypertable failed (timescaledb may not be installed): %v", err)
	}

	// 创建复合索引优化按服务器查询
	if _, err := sqlDB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_metrics_server_time 
		ON server_metrics (server_id, time DESC)
	`); err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	// 创建90天自动数据保留策略（仅TimescaleDB支持）
	if _, err := sqlDB.Exec(`
		SELECT add_retention_policy('server_metrics', 
			if_not_exists => TRUE,
			drop_after => INTERVAL '90 days'
		)
	`); err != nil {
		log.Printf("Warning: retention policy not set: %v", err)
	}

	log.Println("TimescaleDB initialized successfully")
	return nil
}
