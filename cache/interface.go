// Package cache 提供内存缓存与 Redis 缓存实现
package cache

import "time"

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存
	Get(key string) (any, error)
	// GetInto 获取并反序列化到 dest
	GetInto(key string, dest any) error
	// Set 设置缓存
	Set(key string, value any, ttl time.Duration) error
	// Delete 删除缓存
	Delete(key string) error
	Exists(key string) (bool, error)
	Close() error
	// BuildKey 构建缓存键
	BuildKey(prefix string, parts ...interface{}) string
	// BuildQueryKey 构建查询参数缓存键
	BuildQueryKey(prefix string, params interface{}) string
}

var (
	_ Cache = (*MemoryCache)(nil)
	_ Cache = (*RedisCache)(nil)
	_ Cache = (*CacheFactory)(nil)
)
