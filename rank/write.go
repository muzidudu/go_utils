package rank

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Item 为排行榜上的一条记录。
type Item struct {
	Member string
	Score  float64
}

// Incr 将 member 在「今天」（Board 的 Location）下的 type 中增加 delta。
func (b *Board) Incr(ctx context.Context, typ, member string, delta float64) error {
	day := b.dayInLoc(b.now())
	return b.IncrOn(ctx, typ, member, delta, day)
}

// IncrOn 在指定自然日 type 的日 key 中增加分数；day 的时分秒会被忽略，按 Location 取日期。
func (b *Board) IncrOn(ctx context.Context, typ, member string, delta float64, day time.Time) error {
	if err := checkType(typ); err != nil {
		return err
	}
	if err := checkMember(member); err != nil {
		return err
	}
	key := b.zkey(typ, formatDayString(b.dayInLoc(day)))
	return b.rdb.ZIncrBy(ctx, key, delta, member).Err()
}

// SetScore 将某一天某 member 的分数设为 value（非累加）。若需累加请用 IncrOn。
func (b *Board) SetScore(ctx context.Context, typ, member string, day time.Time, value float64) error {
	if err := checkType(typ); err != nil {
		return err
	}
	if err := checkMember(member); err != nil {
		return err
	}
	key := b.zkey(typ, formatDayString(b.dayInLoc(day)))
	return b.rdb.ZAdd(ctx, key, redis.Z{Score: value, Member: member}).Err()
}

// RemMember 从某一天中移除 member。
func (b *Board) RemMember(ctx context.Context, typ, member string, day time.Time) error {
	if err := checkType(typ); err != nil {
		return err
	}
	if err := checkMember(member); err != nil {
		return err
	}
	key := b.zkey(typ, formatDayString(b.dayInLoc(day)))
	return b.rdb.ZRem(ctx, key, member).Err()
}
