package cache

import (
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// ZItem 对应 redis.Z，表示一个有序集合成员及其分数。
type ZItem struct {
	Score  float64
	Member interface{}
}

// ZPair 表示有序集合中成员与分数的键值对。
type ZPair struct {
	Member string
	Score  float64
}

// ---------------------------------------------------------------------------
// ZSet 基础操作
// ---------------------------------------------------------------------------

// ZAdd 向有序集合添加一个或多个成员，已存在的成员更新分数。
// 返回新增成员的数量（不包含已存在仅更新分数的成员）。
func (r *RedisCache) ZAdd(key string, members ...ZItem) (int64, error) {
	zm := make([]redis.Z, len(members))
	for i, m := range members {
		zm[i] = redis.Z{Score: m.Score, Member: m.Member}
	}
	var n int64
	err := r.withClient(func(c *redis.Client) error {
		var e error
		n, e = c.ZAdd(r.ctx, r.fullKey(key), zm...).Result()
		return e
	})
	return n, err
}

// ZScore 获取有序集合中指定成员的分数。
// 成员不存在时返回 0, ErrNotFound。
func (r *RedisCache) ZScore(key, member string) (float64, error) {
	var score float64
	err := r.withClient(func(c *redis.Client) error {
		var e error
		score, e = c.ZScore(r.ctx, r.fullKey(key), member).Result()
		return e
	})
	if err == redis.Nil {
		return 0, ErrNotFound
	}
	return score, err
}

// ZIncrBy 将有序集合中指定成员的分数增量更新。
// 成员不存在时以 increment 为初始分数。
func (r *RedisCache) ZIncrBy(key, member string, increment float64) (float64, error) {
	var score float64
	err := r.withClient(func(c *redis.Client) error {
		var e error
		score, e = c.ZIncrBy(r.ctx, r.fullKey(key), increment, member).Result()
		return e
	})
	return score, err
}

// ZRem 移除有序集合中的一个或多个成员。返回实际移除的数量。
func (r *RedisCache) ZRem(key string, members ...interface{}) (int64, error) {
	var n int64
	err := r.withClient(func(c *redis.Client) error {
		var e error
		n, e = c.ZRem(r.ctx, r.fullKey(key), members...).Result()
		return e
	})
	return n, err
}

// ZCard 返回有序集合的成员数量。
func (r *RedisCache) ZCard(key string) (int64, error) {
	var n int64
	err := r.withClient(func(c *redis.Client) error {
		var e error
		n, e = c.ZCard(r.ctx, r.fullKey(key)).Result()
		return e
	})
	return n, err
}

// ZRank 返回有序集合中指定成员的排名（按分数升序，从 0 开始）。
// 成员不存在时返回 0, ErrNotFound。
func (r *RedisCache) ZRank(key, member string) (int64, error) {
	var rank int64
	err := r.withClient(func(c *redis.Client) error {
		var e error
		rank, e = c.ZRank(r.ctx, r.fullKey(key), member).Result()
		return e
	})
	if err == redis.Nil {
		return 0, ErrNotFound
	}
	return rank, err
}

// ---------------------------------------------------------------------------
// ZSet 范围查询
// ---------------------------------------------------------------------------

// ZRange 返回有序集合中指定下标范围的成员（按分数升序）。
// start/stop 为 0 基下标，-1 表示到末尾。
func (r *RedisCache) ZRange(key string, start, stop int64) ([]string, error) {
	var result []string
	err := r.withClient(func(c *redis.Client) error {
		var e error
		result, e = c.ZRange(r.ctx, r.fullKey(key), start, stop).Result()
		return e
	})
	return result, err
}

// ZRangeWithScores 返回有序集合中指定下标范围的成员及其分数。
func (r *RedisCache) ZRangeWithScores(key string, start, stop int64) ([]ZPair, error) {
	var zms []redis.Z
	err := r.withClient(func(c *redis.Client) error {
		var e error
		zms, e = c.ZRangeWithScores(r.ctx, r.fullKey(key), start, stop).Result()
		return e
	})
	if err != nil {
		return nil, err
	}
	return zToPairs(zms), nil
}

// ZRangeByScore 返回有序集合中分数在 [min, max] 范围内的成员（按分数升序）。
// offset/count 用于分页，count<=0 表示返回全部。
func (r *RedisCache) ZRangeByScore(key string, min, max float64, offset, count int64) ([]string, error) {
	opts := &redis.ZRangeBy{
		Min:    formatFloat(min),
		Max:    formatFloat(max),
		Offset: offset,
		Count:  count,
	}
	var result []string
	err := r.withClient(func(c *redis.Client) error {
		var e error
		result, e = c.ZRangeByScore(r.ctx, r.fullKey(key), opts).Result()
		return e
	})
	return result, err
}

// ZRangeByScoreWithScores 返回有序集合中分数在 [min, max] 范围内的成员及其分数。
// offset/count 用于分页，count<=0 表示返回全部。
func (r *RedisCache) ZRangeByScoreWithScores(key string, min, max float64, offset, count int64) ([]ZPair, error) {
	opts := &redis.ZRangeBy{
		Min:    formatFloat(min),
		Max:    formatFloat(max),
		Offset: offset,
		Count:  count,
	}
	var zms []redis.Z
	err := r.withClient(func(c *redis.Client) error {
		var e error
		zms, e = c.ZRangeByScoreWithScores(r.ctx, r.fullKey(key), opts).Result()
		return e
	})
	if err != nil {
		return nil, err
	}
	return zToPairs(zms), nil
}

// ---------------------------------------------------------------------------
// Hash 操作
// ---------------------------------------------------------------------------

// HSet 向哈希表中设置一个或多个字段值。
// 参数 values 为交替的 field/value 对，如 HSet("k", "f1", "v1", "f2", "v2")。
func (r *RedisCache) HSet(key string, values ...interface{}) error {
	return r.withClient(func(c *redis.Client) error {
		return c.HSet(r.ctx, r.fullKey(key), values...).Err()
	})
}

// HGet 获取哈希表中指定字段的值。
// 字段不存在时返回 "", ErrNotFound。
func (r *RedisCache) HGet(key, field string) (string, error) {
	var val string
	err := r.withClient(func(c *redis.Client) error {
		var e error
		val, e = c.HGet(r.ctx, r.fullKey(key), field).Result()
		return e
	})
	if err == redis.Nil {
		return "", ErrNotFound
	}
	return val, err
}

// HGetAll 获取哈希表中所有字段和值。
func (r *RedisCache) HGetAll(key string) (map[string]string, error) {
	var result map[string]string
	err := r.withClient(func(c *redis.Client) error {
		var e error
		result, e = c.HGetAll(r.ctx, r.fullKey(key)).Result()
		return e
	})
	return result, err
}

// HDel 删除哈希表中的一个或多个字段。返回实际删除的数量。
func (r *RedisCache) HDel(key string, fields ...string) (int64, error) {
	var n int64
	err := r.withClient(func(c *redis.Client) error {
		var e error
		n, e = c.HDel(r.ctx, r.fullKey(key), fields...).Result()
		return e
	})
	return n, err
}

// HExists 检查哈希表中是否存在指定字段。
func (r *RedisCache) HExists(key, field string) (bool, error) {
	var ok bool
	err := r.withClient(func(c *redis.Client) error {
		var e error
		ok, e = c.HExists(r.ctx, r.fullKey(key), field).Result()
		return e
	})
	return ok, err
}

// ---------------------------------------------------------------------------
// Key 过期
// ---------------------------------------------------------------------------

// Expire 为 key 设置过期时间。key 不存在时返回 false。
func (r *RedisCache) Expire(key string, ttl time.Duration) (bool, error) {
	var ok bool
	err := r.withClient(func(c *redis.Client) error {
		var e error
		ok, e = c.Expire(r.ctx, r.fullKey(key), ttl).Result()
		return e
	})
	return ok, err
}

// ExpireAt 为 key 设置过期时间点。
func (r *RedisCache) ExpireAt(key string, at time.Time) (bool, error) {
	var ok bool
	err := r.withClient(func(c *redis.Client) error {
		var e error
		ok, e = c.ExpireAt(r.ctx, r.fullKey(key), at).Result()
		return e
	})
	return ok, err
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

// formatFloat 将 float64 转为字符串，供 ZRangeBy 的 Min/Max 使用。
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// zToPairs 将 []redis.Z 转为 []ZPair。
func zToPairs(zms []redis.Z) []ZPair {
	pairs := make([]ZPair, len(zms))
	for i, z := range zms {
		pairs[i] = ZPair{Member: toString(z.Member), Score: z.Score}
	}
	return pairs
}

// toString 将任意类型转为 string。
func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return fmt.Sprint(v)
}
