// Package bootstrap 应用启动引导，自动初始化配置、缓存、数据库
package bootstrap

import (
	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/cache"
	"gorm.io/gorm"
)

// App 应用容器，持有配置、缓存、数据库等
type App struct {
	Config *Config
	Cache  cache.Cache
	DB     *gorm.DB
	Fiber  *fiber.App
}

// New 创建并初始化应用
// configPath: 配置文件路径，如 "config/config.yaml"，空则使用默认 "config/config.yaml"
func New(configPath string) (*App, error) {
	if configPath == "" {
		configPath = "config/config.yaml"
	}
	configPath = filepath.Clean(configPath)

	// 初始化配置
	cfg, err := initConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 初始化缓存
	c, err := initCache(cfg)
	if err != nil {
		return nil, err
	}

	// 初始化数据库
	db := initDatabase(cfg)

	// 初始化 Fiber
	f := initFiber(cfg)

	return &App{
		Config: cfg,
		Cache:  c,
		DB:     db,
		Fiber:  f,
	}, nil
}

// Close 关闭资源
func (a *App) Close() error {
	if a.Cache != nil {
		_ = a.Cache.Close()
	}
	return nil
}

// Migrate 执行数据库迁移（传入需迁移的模型）
func (a *App) Migrate(models ...any) error {
	return AutoMigrate(a.DB, models...)
}
