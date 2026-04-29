package rank

import "errors"

var (
	// ErrInvalidRange 表示日期区间 from 在 to 之后。
	ErrInvalidRange = errors.New("rank: invalid date range (from after to)")

	// ErrInvalidN 表示 LastNDays 的 n 非法（例如 n < 1）。
	ErrInvalidN = errors.New("rank: invalid n for last n days")

	// ErrTypeEmpty 表示排行类型 type 为空。
	ErrTypeEmpty = errors.New("rank: type is empty")

	// ErrMemberEmpty 表示 member 为空。
	ErrMemberEmpty = errors.New("rank: member is empty")

	// ErrTooManyKeys 表示合并的日 key 数量超过 MaxUnionKeys。
	ErrTooManyKeys = errors.New("rank: too many day keys in single union")

	// ErrInvalidTimeSpec 表示 TimeSpec 无法解析或 Kind 不合法。
	ErrInvalidTimeSpec = errors.New("rank: invalid time spec")
)
