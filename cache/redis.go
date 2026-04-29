package cache

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis 缓存，支持 gzip 压缩与 key 前缀
type RedisCache struct {
	ctx    context.Context
	client *redis.Client
	prefix string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Prefix   string // key 前缀，如 "app:"，空表示无前缀
}

// NewRedisCache 创建 Redis 缓存
func NewRedisCache(cfg RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	// 如果 prefix 不以 : 结尾，则添加 :
	if !strings.HasSuffix(cfg.Prefix, ":") {
		cfg.Prefix = cfg.Prefix + ":"
	}
	return &RedisCache{ctx: context.Background(), client: client, prefix: cfg.Prefix}, nil
}

func (r *RedisCache) fullKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return r.prefix + key
}

// Get 获取缓存，自动解压并反序列化
func (r *RedisCache) Get(key string) (any, error) {
	data, err := r.client.Get(r.ctx, r.fullKey(key)).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	raw, err := gzipDecompress(data)
	if err != nil {
		return nil, err
	}
	return bytesToAny(raw)
}

// GetInto 获取并反序列化到 dest
func (r *RedisCache) GetInto(key string, dest any) error {
	v, err := r.Get(key)
	if err != nil {
		return err
	}
	data, err := valueToBytes(v)
	if err != nil {
		return err
	}
	return bytesToValue(data, dest)
}

// Set 设置缓存，序列化后 gzip 压缩存储
func (r *RedisCache) Set(key string, value any, ttl time.Duration) error {
	raw, err := valueToBytes(value)
	if err != nil {
		return err
	}
	compressed, err := gzipCompress(raw)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, r.fullKey(key), compressed, ttl).Err()
}

// Delete 删除缓存
func (r *RedisCache) Delete(key string) error {
	return r.client.Del(r.ctx, r.fullKey(key)).Err()
}

// DeleteByPrefix 按前缀删除所有匹配的 key，返回删除数量
// keyPrefix 如 "user:" 会删除 "app:user:1", "app:user:2" 等
func (r *RedisCache) DeleteByPrefix(keyPrefix string) (int64, error) {
	pattern := r.fullKey(keyPrefix) + "*"
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}
	return r.client.Del(r.ctx, keys...).Result()
}

// Exists 检查 key 是否存在
func (r *RedisCache) Exists(key string) (bool, error) {
	n, err := r.client.Exists(r.ctx, r.fullKey(key)).Result()
	return n > 0, err
}

// Close 关闭连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// BuildKey 构建缓存键，语义同包函数 BuildKey。
func (_ *RedisCache) BuildKey(prefix string, parts ...interface{}) string {
	return BuildKey(prefix, parts...)
}

// BuildQueryKey 构建查询参数缓存键，语义同包函数 BuildQueryKey。
func (_ *RedisCache) BuildQueryKey(prefix string, params interface{}) string {
	return BuildQueryKey(prefix, params)
}

// Client 返回底层 Redis 客户端
func (r *RedisCache) Client() *redis.Client {
	return r.client
}
