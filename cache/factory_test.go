package cache

import (
	"testing"
)

func TestCacheFactory_FallbackToMemory(t *testing.T) {
	f := NewCacheFactory(FactoryConfig{
		Redis: &RedisConfig{Addr: "localhost:6399"}, // 不可用端口
		Memory: &MemoryConfig{
			MaxCount: 100,
			MaxBytes: 1024 * 1024,
		},
		Prefix: "test:",
	})
	defer f.Close()

	if f.IsRedis() {
		t.Error("expected fallback to memory")
	}
	if err := f.Set("k1", "v1", 0); err != nil {
		t.Fatal(err)
	}
	got, err := f.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	if got.(string) != "v1" {
		t.Errorf("got %q", got)
	}
}

func TestCacheFactory_MemoryOnly(t *testing.T) {
	f := NewCacheFactory(FactoryConfig{
		Memory: &MemoryConfig{MaxCount: 10},
	})
	defer f.Close()

	f.Set("a", "1", 0)
	got, _ := f.Get("a")
	if got.(string) != "1" {
		t.Errorf("got %q", got)
	}
}

func TestCacheFactory_BuildKeyDelegates(t *testing.T) {
	f := NewCacheFactory(FactoryConfig{Memory: &MemoryConfig{MaxCount: 10}})
	defer f.Close()
	if got, want := f.BuildKey("a", 1, "x"), BuildKey("a", 1, "x"); got != want {
		t.Errorf("BuildKey got %q want %q", got, want)
	}
	params := map[string]string{"page": "1"}
	if got, want := f.BuildQueryKey("list", params), BuildQueryKey("list", params); got != want {
		t.Errorf("BuildQueryKey got %q want %q", got, want)
	}
}
