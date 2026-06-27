package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config Agent配置
type Config struct {
	ServerURL  string `yaml:"server_url"`   // Dashboard WebSocket地址，如 wss://dashboard.example.com/ws/agent
	Token      string `yaml:"token"`        // Agent认证Token
	Interval   int    `yaml:"interval"`     // 上报间隔（秒），默认30
	Heartbeat  int    `yaml:"heartbeat"`    // 心跳间隔（秒），默认30
}

// Load 加载配置
func Load(path string) *Config {
	cfg := &Config{
		Interval:  30,
		Heartbeat: 30,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Config file not found at %s, using defaults\n", path)
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		fmt.Printf("Failed to parse config: %v, using defaults\n", err)
		return cfg
	}

	return cfg
}
