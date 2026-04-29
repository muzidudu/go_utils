package rank

import "time"

// TimeSpecKind 表示时间窗口种类。
type TimeSpecKind int

const (
	// TimeSpecDay 为某一自然日（由 Day 决定，仅取日期部分，按 Board 的 Location）。
	TimeSpecDay TimeSpecKind = iota
	// TimeSpecToday 为“今天”。
	TimeSpecToday
	// TimeSpecYesterday 为“昨天”。
	TimeSpecYesterday
	// TimeSpecDateRange 为闭区间 [From, To] 内的每个自然日。
	TimeSpecDateRange
	// TimeSpecLastNDays 为自某个锚点起向前共 N 个自然日（见 IncludeToday 与锚点由调用方在 Board 内解析为 “今天 / 当前时刻”）。

	TimeSpecLastNDays
	// TimeSpecThisWeek 为从 WeekStart 起算的当前自然周，包含 7 个自然日（周界由 Config.WeekStart 决定并落在同一时区自然日内）。
	TimeSpecThisWeek
)

// TimeSpec 描述一次查询要覆盖的日期集合。
// 对 Kind=TimeSpecDay 使用 Day 字段，其它字段忽略。
// 对 TimeSpecDateRange 使用 From、To，均为自然日，闭区间（按 Location 的日期部分）。
// 对 TimeSpecLastNDays 使用 N 与 IncludeToday：若 IncludeToday 为真，则包含“今天”在内的连续 N 天；否则为不含今天的向前 N 天（仍按自然日）。
type TimeSpec struct {
	Kind   TimeSpecKind
	Day    time.Time
	From   time.Time
	To     time.Time
	N      int
	// IncludeToday 对 TimeSpecLastNDays 有效：为 true 时含今天起向前 N 天，为 false 时从昨天起向前 N 天。
	IncludeToday bool
}

// Day 返回单一自然日。
func Day(t time.Time) TimeSpec {
	return TimeSpec{Kind: TimeSpecDay, Day: t}
}

// Today 表示今天。
func Today() TimeSpec {
	return TimeSpec{Kind: TimeSpecToday}
}

// Yesterday 表示昨天。
func Yesterday() TimeSpec {
	return TimeSpec{Kind: TimeSpecYesterday}
}

// DateRange 表示从 from 到 to 的闭区间自然日；若 from 的日期在 to 之后则后续解析会报错。
func DateRange(from, to time.Time) TimeSpec {
	return TimeSpec{Kind: TimeSpecDateRange, From: from, To: to}
}

// LastNDays 为连续 N 个自然日。includeToday 为 true 时含今天（含“今天”的 N 天），为 false 时从昨天起算 N 天（不含今天）。
func LastNDays(n int, includeToday bool) TimeSpec {
	return TimeSpec{Kind: TimeSpecLastNDays, N: n, IncludeToday: includeToday}
}

// ThisWeek 为当前时间所在、以 WeekStart 为一周起始的自然周内的 7 个自然日。
func ThisWeek() TimeSpec {
	return TimeSpec{Kind: TimeSpecThisWeek}
}
