package rank

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

// Board 为基于 Redis ZSET 的日维度排行榜客户端。
type Board struct {
	rdb  redis.Cmdable
	cfg  Config
	loc  *time.Location
	wd   time.Weekday
	maxU int
}

// NewBoard 使用已存在的 Redis 客户端（可为 *redis.Client 或 *redis.ClusterClient 等）创建 Board。
func NewBoard(rdb redis.Cmdable, cfg Config) *Board {
	cfg.normalize()
	return &Board{
		rdb:  rdb,
		cfg:  cfg,
		loc:  cfg.Location,
		wd:   cfg.WeekStart,
		maxU: cfg.MaxUnionKeys,
	}
}

func (b *Board) fullPrefix() string {
	return b.cfg.Prefix
}

func (b *Board) zkey(typ, yyyymmdd string) string {
	return b.fullPrefix() + "rank:" + typ + ":" + yyyymmdd
}

// dayInLoc 将 t 归一化到该时区的当天 0 点。
func (b *Board) dayInLoc(t time.Time) time.Time {
	t = t.In(b.loc)
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, b.loc)
}

func formatDayString(t time.Time) string {
	return t.Format("20060102")
}

func (b *Board) now() time.Time {
	return time.Now().In(b.loc)
}

func (b *Board) tmpUnionKey() string {
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	return b.fullPrefix() + "tmp:union:" + hex.EncodeToString(buf[:])
}

func (b *Board) dayKeysForSpec(typ string, spec TimeSpec) ([]string, error) {
	if err := checkType(typ); err != nil {
		return nil, err
	}
	switch spec.Kind {
	case TimeSpecDay:
		return []string{formatDayString(b.dayInLoc(spec.Day))}, nil
	case TimeSpecToday:
		return []string{formatDayString(b.dayInLoc(b.now()))}, nil
	case TimeSpecYesterday:
		return []string{formatDayString(b.dayInLoc(b.now().AddDate(0, 0, -1)))}, nil
	case TimeSpecDateRange:
		a := b.dayInLoc(spec.From)
		t := b.dayInLoc(spec.To)
		if a.After(t) {
			return nil, ErrInvalidRange
		}
		var out []string
		for d := a; !d.After(t); d = d.AddDate(0, 0, 1) {
			out = append(out, formatDayString(d))
		}
		return out, nil
	case TimeSpecLastNDays:
		if spec.N < 1 {
			return nil, ErrInvalidN
		}
		today := b.dayInLoc(b.now())
		var out []string
		if spec.IncludeToday {
			for i := spec.N - 1; i >= 0; i-- {
				d := today.AddDate(0, 0, -i)
				out = append(out, formatDayString(d))
			}
		} else {
			end := today.AddDate(0, 0, -1)
			start := end.AddDate(0, 0, -(spec.N - 1))
			for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
				out = append(out, formatDayString(d))
			}
		}
		return out, nil
	case TimeSpecThisWeek:
		today := b.dayInLoc(b.now())
		off := (int(today.Weekday()) - int(b.wd) + 7) % 7
		start := today.AddDate(0, 0, -off)
		var out []string
		for i := 0; i < 7; i++ {
			out = append(out, formatDayString(start.AddDate(0, 0, i)))
		}
		return out, nil
	default:
		return nil, ErrInvalidTimeSpec
	}
}

func checkType(typ string) error {
	if typ == "" {
		return ErrTypeEmpty
	}
	return nil
}

func checkMember(m string) error {
	if m == "" {
		return ErrMemberEmpty
	}
	return nil
}

// RDB  returns the underlying cmdable (for tests / advanced use).
func (b *Board) RDB() redis.Cmdable { return b.rdb }

// Location  returns the configured time zone.
func (b *Board) Location() *time.Location { return b.loc }
