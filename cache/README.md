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

// 缓存键：分段用 ":" 拼接；查询类参数先 JSON 再 MD5（prefix:query:十六进制摘要）
uid := int64(1001)
userKey := f.BuildKey("user", uid, "profile")
listKey := f.BuildQueryKey("posts", map[string]string{"page": "1", "size": "20"})
_ = f.Set(userKey, `{"name":""}`, time.Minute)
_ = f.Set(listKey, "[]", time.Minute)
// 亦可使用包函数 cache.BuildKey / cache.BuildQueryKey，行为一致
```

## 多服务场景

不同服务需要不同后端时，在**启动/依赖注入**里各建一个 `CacheFactory`（或 `cache.Cache` 实现），分别注入即可；业务层只依赖 `cache.Cache`，不区分内存与 Redis。

- **仅内存**：`Redis: nil`，只配 `Memory`。
- **优先 Redis**：配置 `Redis`；若连接失败，当前工厂会**自动降级到内存**。若某服务**必须**使用 Redis，启动后检查 `IsRedis()` 为 `false` 时直接退出，或单独对 Redis 做连通性校验。
- **键空间**：各服务使用不同的 `RedisConfig.Prefix`（及业务上的 `BuildKey` 前缀），避免冲突。

```go
// User 服务：只要进程内缓存
userFactory := cache.NewCacheFactory(cache.FactoryConfig{
    Redis:  nil,
    Memory: &cache.MemoryConfig{MaxCount: 10000, MaxBytes: 64 * 1024 * 1024},
})
defer userFactory.Close()

// Art 服务：Redis + 前缀隔离；可选 Memory 作降级兜底
artFactory := cache.NewCacheFactory(cache.FactoryConfig{
    Redis: &cache.RedisConfig{
        Addr:   "localhost:6379",
        DB:     0,
        Prefix: "art:",
    },
    Memory: &cache.MemoryConfig{MaxCount: 10000, MaxBytes: 64 * 1024 * 1024},
    Prefix: "art:",
})
defer artFactory.Close()

// 注入：构造函数只收 cache.Cache，便于测试时换实现
// userSvc := NewUserService(userFactory)   // *CacheFactory 已实现与 Cache 相同的方法集
// artSvc := NewArtService(artFactory)
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
    BuildKey(prefix string, parts ...interface{}) string
    BuildQueryKey(prefix string, params interface{}) string
}
```

实现方：`MemoryCache`、`RedisCache`、`CacheFactory`。键构建与存储后端无关，均委托包函数 `BuildKey` / `BuildQueryKey`。
