// Package cache 提供内存缓存与 Redis 缓存实现
package cache

import "time"

// Cache 缓存接口
type Cache interface {
	Get(key string) (any, error)
	GetInto(key string, dest any) error
	Set(key string, value any, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Close() error
}
