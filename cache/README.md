# cache

内存缓存与 Redis 缓存库。

## 缓存工厂 (CacheFactory)（推荐）

Redis 不可用时自动降级到内存缓存。

```go
f := cache.NewCacheFactory(cache.FactoryConfig{
    Redis: &cache.RedisConfig{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        Prefix:   "app:",  // key 前缀
    },
    Memory: &cache.MemoryConfig{
        MaxCount: 10000,
        MaxBytes: 64 * 1024 * 1024,
    },
    Prefix: "app:",
})
defer f.Close()

// Redis 可用用 Redis，不可用自动降级内存
f.Set("key", "value", time.Minute)
data, _ := f.Get("key")  // data 为 any，可按需类型断言
```

## 内存缓存 (MemoryCache)

- **最大缓存数量**：`MaxCount`，超出时 LRU 淘汰
- **内存限制**：`MaxBytes`（字节），超出时 LRU 淘汰
- **gzip 压缩**：存储前自动压缩，读取时自动解压

```go
c := cache.NewMemoryCache(cache.MemoryConfig{
    MaxCount: 1000,
    MaxBytes: 10 * 1024 * 1024,
})
c.Set("key", "value", 0)
```

## Redis 缓存 (RedisCache)

- **gzip 压缩**：存储前压缩，读取时解压
- **Prefix**：key 前缀，用于命名空间隔离
- **DeleteByPrefix**：按前缀批量删除

```go
c, _ := cache.NewRedisCache(cache.RedisConfig{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
    Prefix:   "app:",
})
c.Set("key", "value", time.Minute)
n, _ := c.DeleteByPrefix("user:")  // 删除 user: 开头的所有 key
```

## 统一接口

```go
type Cache interface {
    Get(key string) (any, error)
    Set(key string, value any, ttl time.Duration) error
    GetInto(key string, dest any) error  // 反序列化到目标对象
    Delete(key string) error
    Exists(key string) (bool, error)
    Close() error
}
```
