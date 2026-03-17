// Package cache 提供内存缓存与 Redis 缓存实现
package cache

import "time"

// Cache 缓存接口
type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Close() error
}
