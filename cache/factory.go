package cache

import (
	"sync"
	"time"
)

// FactoryConfig 缓存工厂配置
type FactoryConfig struct {
	Redis  *RedisConfig  // Redis 配置，nil 表示不使用 Redis
	Memory *MemoryConfig // 内存缓存配置，Redis 不可用时降级使用
	Prefix string        // Redis key 前缀（仅 Redis 使用）
}

// CacheFactory 缓存工厂，Redis 不可用时降级到内存缓存
type CacheFactory struct {
	mu       sync.RWMutex
	cache    Cache
	fallback *MemoryCache
	useRedis bool
}

// NewCacheFactory 创建缓存工厂
// 若 Redis 可用则使用 Redis，否则降级到内存缓存
func NewCacheFactory(cfg FactoryConfig) *CacheFactory {
	f := &CacheFactory{}
	if cfg.Redis != nil {
		redisCfg := *cfg.Redis
		if cfg.Prefix != "" {
			redisCfg.Prefix = cfg.Prefix
		}
		if c, err := NewRedisCache(redisCfg); err == nil {
			f.cache = c
			f.useRedis = true
			return f
		}
	}
	if cfg.Memory != nil {
		f.fallback = NewMemoryCache(*cfg.Memory)
		f.cache = f.fallback
		f.useRedis = false
	} else {
		f.fallback = NewMemoryCache(MemoryConfig{MaxCount: 10000, MaxBytes: 64 * 1024 * 1024})
		f.cache = f.fallback
		f.useRedis = false
	}
	return f
}

// Cache 返回当前使用的缓存（实现 Cache 接口）
func (f *CacheFactory) Cache() Cache {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.cache
}

// IsRedis 是否正在使用 Redis
func (f *CacheFactory) IsRedis() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.useRedis
}

// TryRedis 尝试切换到 Redis（若之前降级到内存且 Redis 已恢复）
func (f *CacheFactory) TryRedis(cfg RedisConfig) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.useRedis {
		return true
	}
	c, err := NewRedisCache(cfg)
	if err != nil {
		return false
	}
	if rc, ok := f.cache.(*RedisCache); ok {
		_ = rc.Close()
	}
	f.cache = c
	f.useRedis = true
	return true
}

// Get 获取缓存
func (f *CacheFactory) Get(key string) ([]byte, error) {
	return f.Cache().Get(key)
}

// Set 设置缓存
func (f *CacheFactory) Set(key string, value []byte, ttl time.Duration) error {
	return f.Cache().Set(key, value, ttl)
}

// Delete 删除缓存
func (f *CacheFactory) Delete(key string) error {
	return f.Cache().Delete(key)
}

// Exists 检查 key 是否存在
func (f *CacheFactory) Exists(key string) (bool, error) {
	return f.Cache().Exists(key)
}

// Close 关闭缓存
func (f *CacheFactory) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.cache != nil {
		return f.cache.Close()
	}
	return nil
}
