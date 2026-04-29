package rank

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func testBoard(t *testing.T) (*Board, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	b := NewBoard(rdb, Config{Prefix: "t", Location: time.FixedZone("Test", 8*3600)})
	return b, mr
}

func TestIncr_TopN_Today(t *testing.T) {
	b, _ := testBoard(t)
	ctx := context.Background()
	if err := b.Incr(ctx, "v", "a", 1); err != nil {
		t.Fatal(err)
	}
	if err := b.Incr(ctx, "v", "b", 3); err != nil {
		t.Fatal(err)
	}
	items, err := b.TopN(ctx, "v", Today(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Member != "b" || items[1].Member != "a" {
		t.Fatalf("items=%+v", items)
	}
}

func TestTopN_Range_Sum(t *testing.T) {
	b, _ := testBoard(t)
	ctx := context.Background()
	day0 := time.Date(2024, 6, 1, 0, 0, 0, 0, b.loc)
	_ = b.IncrOn(ctx, "v", "x", 1, day0)
	_ = b.IncrOn(ctx, "v", "x", 2, day0.AddDate(0, 0, 1))
	items, err := b.TopN(ctx, "v", DateRange(day0, day0.AddDate(0, 0, 1)), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Member != "x" || items[0].Score != 3 {
		t.Fatalf("items=%+v", items)
	}
}

func TestTrimDay_keepTop0(t *testing.T) {
	b, _ := testBoard(t)
	ctx := context.Background()
	_ = b.Incr(ctx, "v", "a", 1)
	day := b.dayInLoc(b.now())
	if err := b.TrimDay(ctx, "v", day, 0); err != nil {
		t.Fatal(err)
	}
	if n, _ := b.RDB().Exists(ctx, b.zkey("v", formatDayString(day))).Result(); n != 0 {
		t.Fatalf("key should be deleted")
	}
}

func TestTrimDay_keepTop1(t *testing.T) {
	b, _ := testBoard(t)
	ctx := context.Background()
	_ = b.Incr(ctx, "v", "a", 1)
	_ = b.Incr(ctx, "v", "b", 5)
	day := b.dayInLoc(b.now())
	if err := b.TrimDay(ctx, "v", day, 1); err != nil {
		t.Fatal(err)
	}
	items, _ := b.TopN(ctx, "v", Day(day), 5)
	if len(items) != 1 || items[0].Member != "b" {
		t.Fatalf("items=%+v", items)
	}
}

func TestPruneTypeBefore(t *testing.T) {
	b, _ := testBoard(t)
	ctx := context.Background()
	d0 := time.Date(2020, 1, 1, 0, 0, 0, 0, b.loc)
	_ = b.IncrOn(ctx, "v", "a", 1, d0)
	_ = b.IncrOn(ctx, "v", "a", 1, d0.AddDate(0, 0, 1))
	before := time.Date(2020, 1, 2, 0, 0, 0, 0, b.loc)
	n, err := b.PruneTypeBefore(ctx, "v", before, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
	_ = b.IncrOn(ctx, "v", "a", 1, d0)
	n2, err := b.PruneTypeBefore(ctx, "v", before, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n2 != 1 {
		t.Fatalf("n2=%d", n2)
	}
}

func TestLastNDays_ExcludeToday(t *testing.T) {
	b, _ := testBoard(t)
	ctx := context.Background()
	day0 := b.dayInLoc(b.now().AddDate(0, 0, -1))
	_ = b.IncrOn(ctx, "v", "x", 1, day0)
	items, err := b.TopN(ctx, "v", LastNDays(1, false), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Member != "x" {
		t.Fatalf("items=%+v", items)
	}
}

func TestThisWeek_7Keys(t *testing.T) {
	b, _ := testBoard(t)
	keys, err := b.dayKeysForSpec("v", ThisWeek())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 7 {
		t.Fatalf("len=%d", len(keys))
	}
}

func TestCount(t *testing.T) {
	b, _ := testBoard(t)
	ctx := context.Background()
	_ = b.Incr(ctx, "v", "a", 1)
	_ = b.Incr(ctx, "v", "b", 1)
	c, err := b.Count(ctx, "v", Today())
	if err != nil {
		t.Fatal(err)
	}
	if c != 2 {
		t.Fatalf("c=%d", c)
	}
}
