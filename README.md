# go_utils

Go 工具库集合，基于 Go 1.25。

## 模块

| 模块 | 说明 |
|------|------|
| [configmgr](./configmgr) | 基于 Viper 的配置管理，支持对象/数组、默认值、热重载 |
| [cache](./cache) | 内存缓存与 Redis 缓存，支持 gzip 压缩、前缀、降级 |

## 使用

```bash
# 初始化工作区
go work sync

# 在项目中引用
go get github.com/muzidudu/go_utils/configmgr
go get github.com/muzidudu/go_utils/cache
```

## 示例

**configmgr**
```go
m := configmgr.NewFromPath("config.yaml")
m.LoadOrInit()
var cfg ServerConfig
m.UnmarshalObjectKey("server", &cfg)
```

**cache**
```go
f := cache.NewCacheFactory(cache.FactoryConfig{
    Redis:  &cache.RedisConfig{Addr: "localhost:6379", Prefix: "app:"},
    Memory: &cache.MemoryConfig{MaxCount: 10000},
})
f.Set("key", []byte("value"), time.Minute)
```

## 文档

- [configmgr 使用说明](./configmgr/USAGE.md)
- [cache README](./cache/README.md)
