package rank

import (
	"context"
	"time"
)

// Read 为只读查询能力：合并时间窗、单日分数、成员数。
type Read interface {
	TopN(ctx context.Context, typ string, spec TimeSpec, n int) ([]Item, error)
	Score(ctx context.Context, typ, member string, day time.Time) (float64, error)
	Count(ctx context.Context, typ string, spec TimeSpec) (int64, error)
}

// Write 为计分与成员写入：累加、覆盖分、按日删除成员。
type Write interface {
	Incr(ctx context.Context, typ, member string, delta float64) error
	IncrOn(ctx context.Context, typ, member string, delta float64, day time.Time) error
	SetScore(ctx context.Context, typ, member string, day time.Time, value float64) error
	RemMember(ctx context.Context, typ, member string, day time.Time) error
}

// Maintenance 为按天裁剪、区间裁剪、按时间扫描清理。
type Maintenance interface {
	TrimDay(ctx context.Context, typ string, day time.Time, keepTop uint) error
	TrimDateRange(ctx context.Context, typ string, from, to time.Time, keepTop uint) error
	PruneTypeBefore(ctx context.Context, typ string, before time.Time, keepTop uint) (n int, err error)
}

// Store 为排行榜对外完整能力，*Board 已实现；用于依赖注入与单测。
type Store interface {
	Read
	Write
	Maintenance
}

// 编译期检查：*Board 实现 Store、Read、Write、Maintenance。
var (
	_ Store       = (*Board)(nil)
	_ Read        = (*Board)(nil)
	_ Write       = (*Board)(nil)
	_ Maintenance = (*Board)(nil)
)
