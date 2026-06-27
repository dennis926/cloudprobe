package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloudprobe/internal/config"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

// Init 初始化Redis连接
func Init(cfg *config.RedisConfig) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// GetClient 获取Redis客户端
func GetClient() *redis.Client {
	return rdb
}

// SetServerStatus 设置服务器在线状态
func SetServerStatus(serverID uint, status string, ttl time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("server:%d:status", serverID)
	return rdb.Set(ctx, key, status, ttl).Err()
}

// GetServerStatus 获取服务器在线状态
func GetServerStatus(serverID uint) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("server:%d:status", serverID)
	return rdb.Get(ctx, key).Result()
}

// SetServerMetrics 缓存服务器最新指标
func SetServerMetrics(serverID uint, metrics map[string]interface{}, ttl time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("server:%d:metrics", serverID)
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, data, ttl).Err()
}

// GetServerMetrics 获取服务器最新指标
func GetServerMetrics(serverID uint) (map[string]interface{}, error) {
	ctx := context.Background()
	key := fmt.Sprintf("server:%d:metrics", serverID)
	data, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var metrics map[string]interface{}
	if err := json.Unmarshal([]byte(data), &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

// GetAllServerMetrics 获取所有服务器实时指标
func GetAllServerMetrics() (map[uint]map[string]interface{}, error) {
	ctx := context.Background()
	pattern := "server:*:metrics"
	keys, err := rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[uint]map[string]interface{})
	for _, key := range keys {
		var serverID uint
		if _, err := fmt.Sscanf(key, "server:%d:metrics", &serverID); err != nil {
			continue
		}
		metrics, err := GetServerMetrics(serverID)
		if err != nil {
			continue
		}
		result[serverID] = metrics
	}
	return result, nil
}

// SetServerHeartbeat 设置服务器心跳时间
func SetServerHeartbeat(serverID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("server:%d:heartbeat", serverID)
	return rdb.Set(ctx, key, time.Now().Unix(), 2*time.Minute).Err()
}

// GetServerHeartbeat 获取服务器心跳时间
func GetServerHeartbeat(serverID uint) (int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("server:%d:heartbeat", serverID)
	val, err := rdb.Get(ctx, key).Int64()
	if err != nil {
		return 0, err
	}
	return val, nil
}
