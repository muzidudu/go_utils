package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// ---------------------------------------------------------------------------
// ZSet 测试
// ---------------------------------------------------------------------------

func TestRedisCache_ZAddZScore(t *testing.T) {
	c := newTestRedis(t)

	n, err := c.ZAdd("z1",
		ZItem{Score: 1, Member: "a"},
		ZItem{Score: 2, Member: "b"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("expected 2 new members, got %d", n)
	}

	score, err := c.ZScore("z1", "a")
	if err != nil {
		t.Fatal(err)
	}
	if score != 1 {
		t.Errorf("expected score 1, got %f", score)
	}

	// 不存在的成员
	_, err = c.ZScore("z1", "nonex")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRedisCache_ZIncrBy(t *testing.T) {
	c := newTestRedis(t)

	c.ZAdd("z1", ZItem{Score: 1, Member: "a"})

	score, err := c.ZIncrBy("z1", "a", 2.5)
	if err != nil {
		t.Fatal(err)
	}
	if score != 3.5 {
		t.Errorf("expected 3.5, got %f", score)
	}

	// 不存在的成员会自动创建
	score, err = c.ZIncrBy("z1", "b", 5)
	if err != nil {
		t.Fatal(err)
	}
	if score != 5 {
		t.Errorf("expected 5, got %f", score)
	}

	card, _ := c.ZCard("z1")
	if card != 2 {
		t.Errorf("expected 2 members, got %d", card)
	}
}

func TestRedisCache_ZRem(t *testing.T) {
	c := newTestRedis(t)

	c.ZAdd("z1",
		ZItem{Score: 1, Member: "a"},
		ZItem{Score: 2, Member: "b"},
		ZItem{Score: 3, Member: "c"},
	)

	n, err := c.ZRem("z1", "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("expected 2 removed, got %d", n)
	}

	card, _ := c.ZCard("z1")
	if card != 1 {
		t.Errorf("expected 1 remaining, got %d", card)
	}
}

func TestRedisCache_ZRank(t *testing.T) {
	c := newTestRedis(t)

	c.ZAdd("z1",
		ZItem{Score: 10, Member: "a"},
		ZItem{Score: 20, Member: "b"},
		ZItem{Score: 30, Member: "c"},
	)

	rank, err := c.ZRank("z1", "b")
	if err != nil {
		t.Fatal(err)
	}
	if rank != 1 {
		t.Errorf("expected rank 1, got %d", rank)
	}

	// 不存在的成员
	_, err = c.ZRank("z1", "nonex")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRedisCache_ZRange(t *testing.T) {
	c := newTestRedis(t)

	c.ZAdd("z1",
		ZItem{Score: 1, Member: "a"},
		ZItem{Score: 2, Member: "b"},
		ZItem{Score: 3, Member: "c"},
	)

	members, err := c.ZRange("z1", 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 3 {
		t.Fatalf("expected 3 members, got %d", len(members))
	}
	if members[0] != "a" || members[2] != "c" {
		t.Errorf("unexpected order: %v", members)
	}
}

func TestRedisCache_ZRangeWithScores(t *testing.T) {
	c := newTestRedis(t)

	c.ZAdd("z1",
		ZItem{Score: 1.5, Member: "a"},
		ZItem{Score: 2.5, Member: "b"},
	)

	pairs, err := c.ZRangeWithScores("z1", 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0].Member != "a" || pairs[0].Score != 1.5 {
		t.Errorf("unexpected pair[0]: %+v", pairs[0])
	}
	if pairs[1].Member != "b" || pairs[1].Score != 2.5 {
		t.Errorf("unexpected pair[1]: %+v", pairs[1])
	}
}

func TestRedisCache_ZRangeByScore(t *testing.T) {
	c := newTestRedis(t)

	c.ZAdd("z1",
		ZItem{Score: 1, Member: "a"},
		ZItem{Score: 2, Member: "b"},
		ZItem{Score: 3, Member: "c"},
		ZItem{Score: 4, Member: "d"},
	)

	// 查询 [2, 3] 范围
	members, err := c.ZRangeByScore("z1", 2, 3, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0] != "b" || members[1] != "c" {
		t.Errorf("unexpected: %v", members)
	}
}

func TestRedisCache_ZRangeByScoreWithScores(t *testing.T) {
	c := newTestRedis(t)

	c.ZAdd("z1",
		ZItem{Score: 1, Member: "a"},
		ZItem{Score: 2, Member: "b"},
		ZItem{Score: 3, Member: "c"},
		ZItem{Score: 4, Member: "d"},
	)

	// 查询 [2, 4] 范围，分页取前 2 个
	pairs, err := c.ZRangeByScoreWithScores("z1", 2, 4, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0].Member != "b" || pairs[0].Score != 2 {
		t.Errorf("unexpected pair[0]: %+v", pairs[0])
	}
	if pairs[1].Member != "c" || pairs[1].Score != 3 {
		t.Errorf("unexpected pair[1]: %+v", pairs[1])
	}
}

func TestRedisCache_ZRangeByScore_Pagination(t *testing.T) {
	c := newTestRedis(t)

	for i := 1; i <= 5; i++ {
		c.ZAdd("z1", ZItem{Score: float64(i), Member: string(rune('a' + i - 1))})
	}

	// count<=0 表示返回全部
	members, err := c.ZRangeByScore("z1", 1, 5, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 5 {
		t.Errorf("expected 5 members, got %d", len(members))
	}
}

// ---------------------------------------------------------------------------
// Hash 测试
// ---------------------------------------------------------------------------

func TestRedisCache_HSetHGet(t *testing.T) {
	c := newTestRedis(t)

	if err := c.HSet("h1", "f1", "v1", "f2", "v2"); err != nil {
		t.Fatal(err)
	}

	val, err := c.HGet("h1", "f1")
	if err != nil {
		t.Fatal(err)
	}
	if val != "v1" {
		t.Errorf("expected v1, got %s", val)
	}

	// 不存在的字段
	_, err = c.HGet("h1", "nonex")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRedisCache_HGetAll(t *testing.T) {
	c := newTestRedis(t)

	c.HSet("h1", "f1", "v1", "f2", "v2")

	all, err := c.HGetAll("h1")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 fields, got %d", len(all))
	}
	if all["f1"] != "v1" || all["f2"] != "v2" {
		t.Errorf("unexpected: %v", all)
	}
}

func TestRedisCache_HDel(t *testing.T) {
	c := newTestRedis(t)

	c.HSet("h1", "f1", "v1", "f2", "v2")

	n, err := c.HDel("h1", "f1")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected 1 deleted, got %d", n)
	}

	ok, _ := c.HExists("h1", "f1")
	if ok {
		t.Error("f1 should not exist")
	}
	ok, _ = c.HExists("h1", "f2")
	if !ok {
		t.Error("f2 should still exist")
	}
}

// ---------------------------------------------------------------------------
// Expire 测试
// ---------------------------------------------------------------------------

func TestRedisCache_Expire(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := NewRedisCache(RedisConfig{Addr: mr.Addr()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { mr.Close(); _ = c.Close() })

	c.Set("k1", "v1", 0)

	ok, err := c.Expire("k1", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected true for existing key")
	}

	mr.FastForward(time.Second + time.Millisecond)

	if _, err := c.Get("k1"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound after expire, got %v", err)
	}
}

func TestRedisCache_Expire_NonExistentKey(t *testing.T) {
	c := newTestRedis(t)

	ok, err := c.Expire("nonex", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected false for non-existent key")
	}
}

// ---------------------------------------------------------------------------
// 前缀测试
// ---------------------------------------------------------------------------

func TestRedisCache_ZSetPrefix(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := NewRedisCache(RedisConfig{Addr: mr.Addr(), Prefix: "app:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { mr.Close(); _ = c.Close() })

	c.ZAdd("z1", ZItem{Score: 1, Member: "a"})

	// 确认底层 key 带有前缀
	exists, _ := mr.ZMembers("app:z1")
	if len(exists) != 1 {
		t.Errorf("expected 1 member in app:z1, got %v", exists)
	}
}

func TestRedisCache_HSetPrefix(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := NewRedisCache(RedisConfig{Addr: mr.Addr(), Prefix: "app:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { mr.Close(); _ = c.Close() })

	c.HSet("h1", "f1", "v1")

	// 确认底层 key 带有前缀
	val := mr.HGet("app:h1", "f1")
	if val != "v1" {
		t.Errorf("expected v1 in app:h1, got %q", val)
	}
}
