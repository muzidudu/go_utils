package bootstrap

import (
	"github.com/muzidudu/go_utils/cache"
)

// initCache 初始化缓存
func initCache(cfg *Config) (cache.Cache, error) {
	if cfg.Cache.Redis != nil {
		rc := &cache.RedisConfig{
			Addr:     cfg.Cache.Redis.Addr,
			Password: cfg.Cache.Redis.Password,
			DB:       cfg.Cache.Redis.DB,
			Prefix:   cfg.Cache.Redis.Prefix,
		}
		var mem *cache.MemoryConfig
		if cfg.Cache.Memory != nil {
			mem = &cache.MemoryConfig{
				MaxCount: cfg.Cache.Memory.MaxCount,
				MaxBytes: cfg.Cache.Memory.MaxBytes,
			}
		}
		f := cache.NewCacheFactory(cache.FactoryConfig{
			Redis:  rc,
			Memory: mem,
			Prefix: rc.Prefix,
		})
		return f.Cache(), nil
	}
	if cfg.Cache.Memory != nil {
		return cache.NewMemoryCache(cache.MemoryConfig{
			MaxCount: cfg.Cache.Memory.MaxCount,
			MaxBytes: cfg.Cache.Memory.MaxBytes,
		}), nil
	}
	return cache.NewMemoryCache(cache.MemoryConfig{
		MaxCount: 10000,
		MaxBytes: 64 * 1024 * 1024,
	}), nil
}
