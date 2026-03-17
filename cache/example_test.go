package cache_test

import (
	"fmt"
	"log"
	"time"

	"github.com/muzidudu/go_utils/cache"
)

func ExampleNewMemoryCache() {
	c := cache.NewMemoryCache(cache.MemoryConfig{
		MaxCount: 1000,              // 最多 1000 条
		MaxBytes: 10 * 1024 * 1024, // 最多 10MB
	})
	defer c.Close()

	// 存储前自动 gzip 压缩
	if err := c.Set("user:1", `{"name":"alice","age":30}`, 0); err != nil {
		log.Fatal(err)
	}
	data, err := c.Get("user:1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("got %s\n", data.(string))
	// Output: got {"name":"alice","age":30}
}

func ExampleNewCacheFactory() {
	// Redis 不可用时自动降级到内存缓存（使用不可用地址确保降级）
	f := cache.NewCacheFactory(cache.FactoryConfig{
		Redis: &cache.RedisConfig{
			Addr:   "127.0.0.1:6399",
			Prefix: "app:",
		},
		Memory: &cache.MemoryConfig{
			MaxCount: 10000,
			MaxBytes: 64 * 1024 * 1024,
		},
		Prefix: "app:",
	})
	defer f.Close()

	f.Set("key", "value", time.Minute)
	data, _ := f.Get("key")
	fmt.Printf("redis=%v data=%s\n", f.IsRedis(), data.(string))
	// Output: redis=false data=value
}
