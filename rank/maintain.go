package rank

import (
	"context"
	"time"
)

// TrimDay 对「某一 type + 某自然日」的 ZSET 做裁剪：keepTop==0 时删除整个 key；keepTop>0 时只保留分数最高的 keepTop 个 member。
func (b *Board) TrimDay(ctx context.Context, typ string, day time.Time, keepTop uint) error {
	if err := checkType(typ); err != nil {
		return err
	}
	key := b.zkey(typ, formatDayString(b.dayInLoc(day)))
	if keepTop == 0 {
		return b.rdb.Del(ctx, key).Err()
	}
	c, err := b.rdb.ZCard(ctx, key).Result()
	if err != nil {
		return err
	}
	if c <= int64(keepTop) {
		return nil
	}
	// 删除分数最低的 (c - keepTop) 个：下标 0 到 c-keepTop-1
	to := c - int64(keepTop) - 1
	if to < 0 {
		return nil
	}
	return b.rdb.ZRemRangeByRank(ctx, key, 0, to).Err()
}

// TrimDateRange 对闭区间 [from, to] 内每个自然日各执行一次 TrimDay。
func (b *Board) TrimDateRange(ctx context.Context, typ string, from, to time.Time, keepTop uint) error {
	a, t := b.dayInLoc(from), b.dayInLoc(to)
	if a.After(t) {
		return ErrInvalidRange
	}
	for d := a; !d.After(t); d = d.AddDate(0, 0, 1) {
		if err := b.TrimDay(ctx, typ, d, keepTop); err != nil {
			return err
		}
	}
	return nil
}

// PruneTypeBefore 使用 SCAN 扫描 prefix 下 rank:{type}:* 的日 key，若 key 的日期**严格小于** `before` 所在自然日，则按 keepTop 规则处理（0 为删 key，>0 为只保留前 keepTop 名）。返回被处理的日 key 数量。
// type 中不应含 Redis glob 特殊字符，以免匹配歧义。
func (b *Board) PruneTypeBefore(ctx context.Context, typ string, before time.Time, keepTop uint) (n int, err error) {
	if err := checkType(typ); err != nil {
		return 0, err
	}
	cut := formatDayString(b.dayInLoc(before))
	pat := b.fullPrefix() + "rank:" + typ + ":*"
	var cursor uint64
	for {
		keys, nc, e := b.rdb.Scan(ctx, cursor, pat, 64).Result()
		if e != nil {
			return n, e
		}
		for _, k := range keys {
			ds, ok := daySuffix8(k)
			if !ok {
				continue
			}
			if ds >= cut {
				continue
			}
			if keepTop == 0 {
				if e := b.rdb.Del(ctx, k).Err(); e != nil {
					return n, e
				}
			} else {
				if e := b.trimKeyTop(ctx, k, keepTop); e != nil {
					return n, e
				}
			}
			n++
		}
		cursor = nc
		if cursor == 0 {
			break
		}
	}
	return n, nil
}

func (b *Board) trimKeyTop(ctx context.Context, key string, keepTop uint) error {
	if keepTop == 0 {
		return b.rdb.Del(ctx, key).Err()
	}
	c, err := b.rdb.ZCard(ctx, key).Result()
	if err != nil {
		return err
	}
	if c <= int64(keepTop) {
		return nil
	}
	to := c - int64(keepTop) - 1
	if to < 0 {
		return nil
	}
	return b.rdb.ZRemRangeByRank(ctx, key, 0, to).Err()
}

func daySuffix8(key string) (string, bool) {
	if len(key) < 8 {
		return "", false
	}
	s := key[len(key)-8:]
	for i := 0; i < 8; i++ {
		if s[i] < '0' || s[i] > '9' {
			return "", false
		}
	}
	return s, true
}
