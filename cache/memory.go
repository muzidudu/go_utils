package cache

import (
	"container/list"
	"sync"
	"time"
)

// MemoryCache 内存缓存，支持最大数量、内存限制、gzip 压缩
type MemoryCache struct {
	mu       sync.RWMutex
	data     map[string]*memItem
	lru      *list.List
	maxCount int64
	maxBytes int64
	curBytes int64
}

type memItem struct {
	key       string
	value     []byte // 已 gzip 压缩
	expireAt  time.Time
	size      int64
	lruElem   *list.Element
}

// MemoryConfig 内存缓存配置
type MemoryConfig struct {
	MaxCount int64 // 最大缓存数量，0 表示不限制
	MaxBytes int64 // 最大内存字节数，0 表示不限制
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(cfg MemoryConfig) *MemoryCache {
	return &MemoryCache{
		data:     make(map[string]*memItem),
		lru:      list.New(),
		maxCount: cfg.MaxCount,
		maxBytes: cfg.MaxBytes,
	}
}

// Get 获取缓存，自动解压并反序列化
func (c *MemoryCache) Get(key string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		c.removeItem(item)
		return nil, ErrNotFound
	}
	c.lru.MoveToFront(item.lruElem)
	raw, err := gzipDecompress(item.value)
	if err != nil {
		return nil, err
	}
	return bytesToAny(raw)
}

// GetInto 获取并反序列化到 dest
func (c *MemoryCache) GetInto(key string, dest any) error {
	v, err := c.Get(key)
	if err != nil {
		return err
	}
	// 通过 json  roundtrip 实现 any -> dest
	data, err := valueToBytes(v)
	if err != nil {
		return err
	}
	return bytesToValue(data, dest)
}

// Set 设置缓存，序列化后 gzip 压缩存储
func (c *MemoryCache) Set(key string, value any, ttl time.Duration) error {
	raw, err := valueToBytes(value)
	if err != nil {
		return err
	}
	compressed, err := gzipCompress(raw)
	if err != nil {
		return err
	}
	size := int64(len(compressed))
	c.mu.Lock()
	defer c.mu.Unlock()
	// 若已存在则先移除
	if old, ok := c.data[key]; ok {
		c.removeItem(old)
	}
	// 淘汰直到满足限制
	for (c.maxCount > 0 && int64(len(c.data)) >= c.maxCount) || (c.maxBytes > 0 && c.curBytes+size > c.maxBytes) {
		if !c.evictOldest() {
			break
		}
	}
	// 单条超限则拒绝
	if c.maxBytes > 0 && size > c.maxBytes {
		return ErrItemTooLarge
	}
	item := &memItem{
		key:      key,
		value:    compressed,
		size:     size,
		lruElem:  c.lru.PushFront(key),
	}
	if ttl > 0 {
		item.expireAt = time.Now().Add(ttl)
	}
	c.data[key] = item
	c.curBytes += size
	return nil
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, ok := c.data[key]; ok {
		c.removeItem(item)
	}
	return nil
}

// Exists 检查 key 是否存在
func (c *MemoryCache) Exists(key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.data[key]
	if !ok {
		return false, nil
	}
	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		return false, nil
	}
	return true, nil
}

// Close 关闭缓存（无操作，兼容接口）
func (c *MemoryCache) Close() error {
	return nil
}

// Stats 返回统计信息
func (c *MemoryCache) Stats() (count int, bytes int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data), c.curBytes
}

func (c *MemoryCache) removeItem(item *memItem) {
	delete(c.data, item.key)
	c.lru.Remove(item.lruElem)
	c.curBytes -= item.size
}

func (c *MemoryCache) evictOldest() bool {
	elem := c.lru.Back()
	if elem == nil {
		return false
	}
	key := elem.Value.(string)
	if item, ok := c.data[key]; ok {
		c.removeItem(item)
		return true
	}
	c.lru.Remove(elem)
	return true
}
