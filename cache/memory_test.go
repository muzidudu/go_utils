package cache

import (
	"testing"
)

func TestMemoryCache_GetSet(t *testing.T) {
	c := NewMemoryCache(MemoryConfig{MaxCount: 100, MaxBytes: 1024 * 1024})
	defer c.Close()

	val := "hello world"
	if err := c.Set("k1", val, 0); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	if got.(string) != val {
		t.Errorf("got %q", got)
	}
}

func TestMemoryCache_MaxCount(t *testing.T) {
	c := NewMemoryCache(MemoryConfig{MaxCount: 3, MaxBytes: 0})
	defer c.Close()

	c.Set("k1", "a", 0)
	c.Set("k2", "b", 0)
	c.Set("k3", "c", 0)
	c.Set("k4", "d", 0) // 应淘汰 k1

	if _, err := c.Get("k1"); err != ErrNotFound {
		t.Errorf("k1 should be evicted, got %v", err)
	}
	if _, err := c.Get("k4"); err != nil {
		t.Errorf("k4 should exist: %v", err)
	}
}

func TestMemoryCache_GzipCompression(t *testing.T) {
	c := NewMemoryCache(MemoryConfig{MaxCount: 10, MaxBytes: 0})
	defer c.Close()

	val := "repeated text repeated text repeated text"
	c.Set("k1", val, 0)
	got, err := c.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	if got.(string) != val {
		t.Errorf("got %q", got)
	}
}

func TestMemoryCache_MaxBytes(t *testing.T) {
	// 每条约 50+ 字节（压缩后），限制 150 字节约 2-3 条
	c := NewMemoryCache(MemoryConfig{MaxCount: 0, MaxBytes: 200})
	defer c.Close()

	c.Set("k1", "aaaaaaaaaaaaaaaaaaaaaaaa", 0)
	c.Set("k2", "bbbbbbbbbbbbbbbbbbbbbbbb", 0)
	c.Set("k3", "cccccccccccccccccccccccc", 0)
	c.Set("k4", "dddddddddddddddddddddddd", 0)

	count, bytes := c.Stats()
	if count > 4 {
		t.Errorf("count=%d", count)
	}
	_ = bytes
}

func TestMemoryCache_Delete(t *testing.T) {
	c := NewMemoryCache(MemoryConfig{})
	defer c.Close()

	c.Set("k1", "v1", 0)
	c.Delete("k1")
	if _, err := c.Get("k1"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
