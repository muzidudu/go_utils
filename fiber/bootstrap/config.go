package bootstrap

import (
	"fmt"

	"github.com/muzidudu/go_utils/configmgr"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
}

type ServerConfig struct {
	Host  string `mapstructure:"host"`
	Port  int    `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
}

type CacheConfig struct {
	Redis  *RedisCacheConfig  `mapstructure:"redis"`
	Memory *MemoryCacheConfig `mapstructure:"memory"`
}

// RedisCacheConfig 与 cache.RedisConfig 结构一致，用于 yaml 解析
type RedisCacheConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	Prefix   string `mapstructure:"prefix"`
}

// MemoryCacheConfig 与 cache.MemoryConfig 结构一致
type MemoryCacheConfig struct {
	MaxCount int64 `mapstructure:"max_count"`
	MaxBytes int64 `mapstructure:"max_bytes"`
}

// initConfig 加载并解析配置
func initConfig(configPath string) (*Config, error) {
	mgr := configmgr.NewFromPath(configPath)
	if err := mgr.LoadOrInitWithDefaults(map[string]any{
		"server.host":   "0.0.0.0",
		"server.port":   3000,
		"server.debug":  false,
		"database.host": "localhost",
		"database.port": 5432,
		"database.user": "postgres",
	}); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	var cfg Config
	if err := mgr.UnmarshalObject(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
