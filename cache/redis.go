package cache

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisDialTimeout      = 5 * time.Second
	redisPingTimeout      = 5 * time.Second
	redisReconnectTimeout = 2 * time.Second
	minReconnectBackoff   = 500 * time.Millisecond
	maxReconnectBackoff   = 30 * time.Second
)

// RedisCache Redis 缓存，支持 gzip 压缩与 key 前缀。
// 底层客户端异常关闭或连接断开时，下一次操作会自动重建连接并重试一次。
type RedisCache struct {
	mu               sync.RWMutex
	ctx              context.Context
	client           *redis.Client
	cfg              RedisConfig
	prefix           string
	closed           bool // 显式 Close() 后不再重连
	lastReconnectAt  time.Time
	reconnectBackoff time.Duration
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Prefix   string // key 前缀，如 "app:"，空表示无前缀
}

func redisOptions(cfg RedisConfig) *redis.Options {
	return &redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		DialTimeout:     redisDialTimeout,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        10,
		MinIdleConns:    1,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

func normalizePrefix(prefix string) string {
	if !strings.HasSuffix(prefix, ":") {
		return prefix + ":"
	}
	return prefix
}

// NewRedisCache 创建 Redis 缓存
func NewRedisCache(cfg RedisConfig) (*RedisCache, error) {
	cfg.Prefix = normalizePrefix(cfg.Prefix)
	client := redis.NewClient(redisOptions(cfg))
	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &RedisCache{
		ctx:              context.Background(),
		client:           client,
		cfg:              cfg,
		prefix:           cfg.Prefix,
		reconnectBackoff: minReconnectBackoff,
	}, nil
}

func (r *RedisCache) fullKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return r.prefix + key
}

func (r *RedisCache) isClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.closed
}

func (r *RedisCache) getClient() (*redis.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed || r.client == nil {
		return nil, redis.ErrClosed
	}
	return r.client, nil
}

// isReconnectable 判断是否因连接异常而需要重建客户端。
func isReconnectable(err error) bool {
	if err == nil || errors.Is(err, redis.Nil) {
		return false
	}
	if errors.Is(err, redis.ErrClosed) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "use of closed network connection")
}

func (r *RedisCache) reconnect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return redis.ErrClosed
	}
	if !r.lastReconnectAt.IsZero() && time.Since(r.lastReconnectAt) < r.reconnectBackoff {
		return errors.New("cache: redis reconnect backing off")
	}

	ping := func(c *redis.Client) error {
		ctx, cancel := context.WithTimeout(context.Background(), redisReconnectTimeout)
		defer cancel()
		return c.Ping(ctx).Err()
	}
	if r.client != nil && ping(r.client) == nil {
		r.lastReconnectAt = time.Time{}
		r.reconnectBackoff = minReconnectBackoff
		return nil
	}

	client := redis.NewClient(redisOptions(r.cfg))
	if err := ping(client); err != nil {
		_ = client.Close()
		if r.lastReconnectAt.IsZero() {
			r.reconnectBackoff = minReconnectBackoff
		} else {
			r.reconnectBackoff *= 2
			if r.reconnectBackoff > maxReconnectBackoff {
				r.reconnectBackoff = maxReconnectBackoff
			}
		}
		r.lastReconnectAt = time.Now()
		return err
	}

	old := r.client
	r.client = client
	r.lastReconnectAt = time.Time{}
	r.reconnectBackoff = minReconnectBackoff
	if old != nil {
		_ = old.Close()
	}
	return nil
}

func (r *RedisCache) withClient(fn func(*redis.Client) error) error {
	if r.isClosed() {
		return redis.ErrClosed
	}
	client, err := r.getClient()
	if err != nil {
		if reconnErr := r.reconnect(); reconnErr != nil {
			return err
		}
		client, err = r.getClient()
		if err != nil {
			return err
		}
	}
	err = fn(client)
	if !isReconnectable(err) {
		return err
	}
	if reconnErr := r.reconnect(); reconnErr != nil {
		return err
	}
	client, err2 := r.getClient()
	if err2 != nil {
		return err
	}
	return fn(client)
}

// Get 获取缓存，自动解压并反序列化
func (r *RedisCache) Get(key string) (any, error) {
	var data []byte
	err := r.withClient(func(c *redis.Client) error {
		var e error
		data, e = c.Get(r.ctx, r.fullKey(key)).Bytes()
		return e
	})
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
	return r.withClient(func(c *redis.Client) error {
		return c.Set(r.ctx, r.fullKey(key), compressed, ttl).Err()
	})
}

// Delete 删除缓存
func (r *RedisCache) Delete(key string) error {
	return r.withClient(func(c *redis.Client) error {
		return c.Del(r.ctx, r.fullKey(key)).Err()
	})
}

// DeleteByPrefix 按前缀删除所有匹配的 key，返回删除数量
// keyPrefix 如 "user:" 会删除 "app:user:1", "app:user:2" 等
func (r *RedisCache) DeleteByPrefix(keyPrefix string) (int64, error) {
	var n int64
	err := r.withClient(func(c *redis.Client) error {
		pattern := r.fullKey(keyPrefix) + "*"
		keys, err := c.Keys(r.ctx, pattern).Result()
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			n = 0
			return nil
		}
		n, err = c.Del(r.ctx, keys...).Result()
		return err
	})
	return n, err
}

// Exists 检查 key 是否存在
func (r *RedisCache) Exists(key string) (bool, error) {
	var ok bool
	err := r.withClient(func(c *redis.Client) error {
		n, err := c.Exists(r.ctx, r.fullKey(key)).Result()
		ok = n > 0
		return err
	})
	return ok, err
}

// Close 关闭连接。调用后不再自动重连。
func (r *RedisCache) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	if r.client == nil {
		return nil
	}
	err := r.client.Close()
	r.client = nil
	return err
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.client
}
