package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Notify   NotifyConfig   `mapstructure:"notify"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
	Backup   BackupConfig   `mapstructure:"backup"`
	XUI      XUIConfig      `mapstructure:"xui"`
}

// ServerConfig Web服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	GRPCPort     int    `mapstructure:"grpc_port"`
	Mode         string `mapstructure:"mode"` // release / debug
	BasePath     string `mapstructure:"base_path"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type         string `mapstructure:"type"`
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	ExpireHours   time.Duration `mapstructure:"expire_hours"`
	RefreshHours  time.Duration `mapstructure:"refresh_hours"`
}

// AgentConfig Agent配置
type AgentConfig struct {
	Token     string `mapstructure:"token"`
	Interval  int    `mapstructure:"interval"`
	GRPCAddr  string `mapstructure:"grpc_addr"`
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	Channels []string `mapstructure:"channels"`
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// BackupConfig 备份配置
type BackupConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	Email        string   `mapstructure:"email"`
	Schedule     string   `mapstructure:"schedule"`
	KeepLocal    int      `mapstructure:"keep_local"`
	EmailTo      []string `mapstructure:"email_to"`
	EmailSubject string   `mapstructure:"email_subject"`
}

// XUIConfig 3x-ui集成配置
type XUIConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	PanelURL string `mapstructure:"panel_url"`
	APIToken string `mapstructure:"api_token"`
	BasePath string `mapstructure:"base_path"`
}

var globalConfig *Config

// Load 加载配置
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/cloudprobe")
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	// 从环境变量读取
	viper.SetEnvPrefix("cp")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found, using defaults and env: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	globalConfig = &cfg
	return &cfg
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		return Load()
	}
	return globalConfig
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8000)
	viper.SetDefault("server.grpc_port", 50051)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.base_path", "/")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	viper.SetDefault("database.type", "postgres")
	viper.SetDefault("database.dsn", "postgres://cpuser:cppass@localhost:5432/cloudprobe?sslmode=disable")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)

	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("jwt.secret", "cloudprobe-secret-change-me")
	viper.SetDefault("jwt.expire_hours", 24)
	viper.SetDefault("jwt.refresh_hours", 168)

	viper.SetDefault("agent.mode", "foreign")
	viper.SetDefault("agent.report_interval", 30)

	viper.SetDefault("backup.enabled", true)
	viper.SetDefault("backup.schedule", "0 2 * * *")
	viper.SetDefault("backup.keep_local", 7)
	viper.SetDefault("backup.email_subject", "CloudProbe Daily Backup")
}
