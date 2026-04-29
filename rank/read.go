package rank

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// TopN 按 TimeSpec 合并多个自然日 ZSET 后，返回分数最高的前 n 个（n<=0 时返回空 slice）。
func (b *Board) TopN(ctx context.Context, typ string, spec TimeSpec, n int) ([]Item, error) {
	if n <= 0 {
		return nil, nil
	}
	keys, err := b.zkeysForSpec(ctx, typ, spec)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	if len(keys) == 1 {
		return b.topNOneKey(ctx, keys[0], n)
	}
	if len(keys) > b.maxU {
		return nil, ErrTooManyKeys
	}
	tmp := b.tmpUnionKey()
	defer b.rdb.Del(ctx, tmp) //nolint:errcheck
	zs := make([]storeKey, 0, len(keys))
	for _, k := range keys {
		zs = append(zs, storeKey{k: k, w: 1})
	}
	if err := zunionInto(ctx, b.rdb, tmp, zs); err != nil {
		return nil, err
	}
	return b.topNOneKey(ctx, tmp, n)
}

type storeKey struct {
	k string
	w float64
}

func zunionInto(ctx context.Context, rdb redis.Cmdable, dest string, keys []storeKey) error {
	ks := make([]string, len(keys))
	weights := make([]float64, len(keys))
	for i := range keys {
		ks[i] = keys[i].k
		weights[i] = keys[i].w
	}
	// 相同 member 在多日之间求和
	return rdb.ZUnionStore(ctx, dest, &redis.ZStore{
		Keys:      ks,
		Weights:   weights,
		Aggregate: "SUM",
	}).Err()
}

func (b *Board) zkeysForSpec(_ context.Context, typ string, spec TimeSpec) ([]string, error) {
	days, err := b.dayKeysForSpec(typ, spec)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(days))
	for _, d := range days {
		keys = append(keys, b.zkey(typ, d))
	}
	return keys, nil
}

func (b *Board) topNOneKey(ctx context.Context, key string, n int) ([]Item, error) {
	rr, err := b.rdb.ZRevRangeWithScores(ctx, key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	out := make([]Item, 0, len(rr))
	for _, z := range rr {
		m, ok := z.Member.(string)
		if !ok {
			continue
		}
		out = append(out, Item{Member: m, Score: z.Score})
	}
	return out, nil
}

// Score 返回 member 在「指定单一自然日」的分数。若 key 中不存在，分数为 0 且 err 为 nil（与 redis ZSCORE 一致，可用 redis.Nil 判断无成员时可改为返回 0, nil 或单独 API）。
// 为简化：若 member 不在集合中，返回 0, nil；若 key 不存在，也返回 0, nil。可用 Exists 的语义若需要可扩展。
func (b *Board) Score(ctx context.Context, typ, member string, day time.Time) (float64, error) {
	if err := checkType(typ); err != nil {
		return 0, err
	}
	if err := checkMember(member); err != nil {
		return 0, err
	}
	key := b.zkey(typ, formatDayString(b.dayInLoc(day)))
	s, err := b.rdb.ZScore(ctx, key, member).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return s, nil
}

// Count 在 TimeSpec 对应的时间窗内，合并后统计不重复 member 数（ZUNION 后对 temp key ZCARD；单日则直接 ZCARD）。
func (b *Board) Count(ctx context.Context, typ string, spec TimeSpec) (int64, error) {
	keys, err := b.zkeysForSpec(ctx, typ, spec)
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}
	if len(keys) == 1 {
		return b.rdb.ZCard(ctx, keys[0]).Result()
	}
	if len(keys) > b.maxU {
		return 0, ErrTooManyKeys
	}
	tmp := b.tmpUnionKey()
	defer b.rdb.Del(ctx, tmp) //nolint:errcheck
	zs := make([]storeKey, 0, len(keys))
	for _, k := range keys {
		zs = append(zs, storeKey{k: k, w: 1})
	}
	if err := zunionInto(ctx, b.rdb, tmp, zs); err != nil {
		return 0, err
	}
	return b.rdb.ZCard(ctx, tmp).Result()
}
