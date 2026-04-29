package rank

import (
	"strings"
	"time"
)

// Config 控制 Redis key 前缀、时区与合并上限等。
type Config struct {
	// Prefix 为 key 命名空间前缀，非空时若未以 : 结尾会自动补 :。
	Prefix string

	// Location 用于“今天/昨天/本周/自然日”的切分。nil 时默认 time.Local。
	Location *time.Location

	// WeekStart 为「本周」的起始星期。零值为 time.Monday（周一为一周开始）。
	WeekStart time.Weekday

	// MaxUnionKeys 为单次多日系合并（ZUNIONSTORE）允许的最大日 key 数。0 表示 400。
	MaxUnionKeys int
}

func (c *Config) normalize() {
	if c.Location == nil {
		c.Location = time.Local
	}
	if c.WeekStart == 0 {
		c.WeekStart = time.Monday
	}
	if c.MaxUnionKeys == 0 {
		c.MaxUnionKeys = 400
	}
	if c.Prefix != "" && !strings.HasSuffix(c.Prefix, ":") {
		c.Prefix += ":"
	}
}
