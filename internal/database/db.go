package database

import (
	"fmt"
	"log"
	"time"

	"cloudprobe/internal/config"
	"cloudprobe/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig) error {
	var err error
	var dialector gorm.Dialector

	switch cfg.Type {
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化TimescaleDB
	if err := InitTimescaleDB(); err != nil {
		log.Printf("Warning: TimescaleDB init failed: %v", err)
	}

	// 初始化默认管理员
	if err := initDefaultUser(); err != nil {
		return fmt.Errorf("failed to init default user: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// autoMigrate 自动迁移数据表
func autoMigrate() error {
	return db.AutoMigrate(
		&model.User{},
		&model.Server{},
		&model.ServerGroup{},
		&model.ServerTag{},
		&model.ServerBill{},
		&model.AlertRule{},
		&model.Alert{},
		&model.NotificationLog{},
		&model.NotificationChannel{},
		&model.APIToken{},
		&model.AuditLog{},
	)
}

// initDefaultUser 初始化默认管理员账户
func initDefaultUser() error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	user := model.User{
		Username: "admin",
		Password: "$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquzi.Ss7KIUgO2t0jWMUW", // admin (bcrypt)
		Role:     "admin",
		Status:   "active",
	}

	return db.Create(&user).Error
}
