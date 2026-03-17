package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func newTestRedis(t *testing.T) *RedisCache {
	mr := miniredis.RunT(t)
	c, err := NewRedisCache(RedisConfig{
		Addr:   mr.Addr(),
		Prefix: "test:",
	})
	if err != nil {
		t.Fatalf("NewRedisCache: %v", err)
	}
	t.Cleanup(func() {
		mr.Close()
		_ = c.Close()
	})
	return c
}

func TestRedisCache_GetSet(t *testing.T) {
	c := newTestRedis(t)

	val := []byte("hello world")
	if err := c.Set("k1", val, 0); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(val) {
		t.Errorf("got %q", got)
	}
}

func TestRedisCache_GzipCompression(t *testing.T) {
	c := newTestRedis(t)

	val := []byte("repeated text repeated text repeated text")
	if err := c.Set("k1", val, 0); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(val) {
		t.Errorf("got %q", got)
	}
}

func TestRedisCache_Prefix(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := NewRedisCache(RedisConfig{
		Addr:   mr.Addr(),
		Prefix: "app:",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	defer mr.Close()

	c.Set("user:1", []byte("alice"), 0)
	got, _ := c.Get("user:1")
	if string(got) != "alice" {
		t.Errorf("got %q", got)
	}
}

func TestRedisCache_Delete(t *testing.T) {
	c := newTestRedis(t)

	c.Set("k1", []byte("v1"), 0)
	c.Delete("k1")
	if _, err := c.Get("k1"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRedisCache_DeleteByPrefix(t *testing.T) {
	c := newTestRedis(t)

	c.Set("user:1", []byte("a"), 0)
	c.Set("user:2", []byte("b"), 0)
	c.Set("order:1", []byte("c"), 0)

	n, err := c.DeleteByPrefix("user:")
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted, got %d", n)
	}
	if _, err := c.Get("user:1"); err != ErrNotFound {
		t.Error("user:1 should be deleted")
	}
	if _, err := c.Get("user:2"); err != ErrNotFound {
		t.Error("user:2 should be deleted")
	}
	got, _ := c.Get("order:1")
	if string(got) != "c" {
		t.Errorf("order:1 should remain, got %q", got)
	}
}

func TestRedisCache_Exists(t *testing.T) {
	c := newTestRedis(t)

	ok, _ := c.Exists("nonex")
	if ok {
		t.Error("expected false")
	}
	c.Set("k1", []byte("v1"), 0)
	ok, _ = c.Exists("k1")
	if !ok {
		t.Error("expected true")
	}
}

func TestRedisCache_TTL(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := NewRedisCache(RedisConfig{Addr: mr.Addr()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { mr.Close(); _ = c.Close() })

	c.Set("k1", []byte("v1"), time.Second)
	got, _ := c.Get("k1")
	if string(got) != "v1" {
		t.Errorf("got %q", got)
	}
	mr.FastForward(time.Second + time.Millisecond)
	if _, err := c.Get("k1"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound after TTL, got %v", err)
	}
}

func TestCacheFactory_WithRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	f := NewCacheFactory(FactoryConfig{
		Redis: &RedisConfig{
			Addr:   mr.Addr(),
			Prefix: "app:",
		},
		Memory: &MemoryConfig{MaxCount: 100},
		Prefix: "app:",
	})
	defer f.Close()

	if !f.IsRedis() {
		t.Error("expected Redis to be used")
	}
	f.Set("k1", []byte("v1"), 0)
	got, _ := f.Get("k1")
	if string(got) != "v1" {
		t.Errorf("got %q", got)
	}
}
