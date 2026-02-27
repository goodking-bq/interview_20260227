package storage

import (
	"fmt"
	"os"
)

// Config 存储配置
type Config struct {
	Type string // 存储类型: "memory", "sqlite", "redis"
	DSN  string // 数据源连接字符串
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() Config {
	return Config{
		Type: getEnvOrDefault("STORAGE_TYPE", "memory"),
		DSN:  getEnvOrDefault("STORAGE_DSN", ""),
	}
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewStorage 根据配置创建对应的存储实现
func NewStorage(cfg Config) (Storage, error) {
	switch cfg.Type {
	case "memory":
		return NewMemoryStorage(), nil
	case "sqlite":
		return NewSQLiteStorage(cfg.DSN)
	case "redis":
		return NewRedisStorage(cfg.DSN)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", cfg.Type)
	}
}
