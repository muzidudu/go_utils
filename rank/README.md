# go_utils/rank

基于 [Redis](https://redis.io) 有序集合（`ZSET`）的**按类型、按自然日**排行榜库：支持单日/区间/今天、昨天、本周、最近 N 天等多时间窗查询，多日分数按 **SUM** 合并；提供按名次裁剪与按日期扫描清理，便于控制**沉积数据**体量。

## 特性

- **多业务类型隔离**：`type` 字符串区分不同排行（如浏览量、搜索热词），独立 key 空间  
- **时间窗**：`TimeSpec` 支持单日、今天/昨天、闭区间、最近 N 天（可含/不含今天）、本周（`Config.WeekStart`，默认周一）  
- **读写与维护**：`rank.Store` 涵盖计分、TopN、按天裁剪、区间裁剪、`SCAN` + 日期解析的 `PruneTypeBefore`  
- **依赖注入**：`*Board` 实现 `Read` / `Write` / `Maintenance` / `Store` 接口，便于单测 mock  
- **合并上限**：`Config.MaxUnionKeys` 限制单次 `ZUNIONSTORE` 源 key 数量（默认 400），避免过长区间拖垮 Redis  

## 安装

```bash
go get github.com/muzidudu/go_utils/rank
```

依赖：`github.com/redis/go-redis/v9`（注入已存在的 `redis.Client` 或兼容 `redis.Cmdable` 的客户端）。

## 数据与 key 约定

- **Key 形态**：`{Prefix}rank:{type}:YYYYMMDD`（`YYYYMMDD` 由 `Config.Location` 决定自然日）  
- **Member / Score**：业务侧自行约定 `member` 字符串（如 `doc:1`、规范化后的搜索词）；分数多为累加计数（`ZINCRBY`）  
- **多日查询**：内部对涉及的日 key 做 `ZUNIONSTORE … AGGREGATE SUM`，写入临时 key 后 `ZREVRANGE` 取 TopN，再删除临时 key  

## 快速开始

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/muzidudu/go_utils/rank"
	"github.com/redis/go-redis/v9"
)

func main() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	b := rank.NewBoard(rdb, rank.Config{
		Prefix:   "app",
		Location: loc,
	})
	ctx := context.Background()

	// 今日热词 +1
	if err := b.Incr(ctx, "search_kw", "golang 教程", 1); err != nil {
		log.Fatal(err)
	}

	// 最近 7 天（含今天）Top 20
	items, err := b.TopN(ctx, "search_kw", rank.LastNDays(7, true), 20)
	if err != nil {
		log.Fatal(err)
	}
	for _, it := range items {
		log.Printf("%s %.0f\n", it.Member, it.Score)
	}
}
```

## 配置 `Config`

| 字段 | 说明 |
|------|------|
| `Prefix` | key 前缀；非空且未以 `:` 结尾时会自动补 `:` |
| `Location` | 自然日与时区；`nil` 为 `time.Local` |
| `WeekStart` | 「本周」的周起始星期；零值为 `time.Monday` |
| `MaxUnionKeys` | 单次合并的日 key 数上限；`0` 为 400 |

## 时间规格 `TimeSpec`

| 构造函数 | 含义 |
|----------|------|
| `Day(t)` | 某一自然日 |
| `Today()` / `Yesterday()` | 今天 / 昨天 |
| `DateRange(from, to)` | 闭区间 `[from, to]` 的每个自然日 |
| `LastNDays(n, includeToday)` | 连续 N 个自然日；第二参数为 `true` 时含今天 |
| `ThisWeek()` | 当前自然周 7 天（从 `WeekStart` 起算） |

## 维护与 `keepTop`

| 方法 | 说明 |
|------|------|
| `TrimDay` / `TrimDateRange` | 对单日或 `[from, to]` 内每日： `keepTop == 0` 时 **删除整个日 key**；`keepTop > 0` 时只保留分数最高的 `keepTop` 个 member |
| `PruneTypeBefore` | 对 `SCAN` 匹配到的 `rank:{type}:*` key，若 key 末尾 **8 位日期** **严格小于** `before` 所在自然日，则按同上 `keepTop` 规则处理；返回处理的 key 数量 |

`type` 中请勿包含 Redis glob 通配字符，以免 `PruneTypeBefore` 的匹配模式产生歧义。

## 接口概览

| 接口 | 方法 |
|------|------|
| `Write` | `Incr`, `IncrOn`, `SetScore`, `RemMember` |
| `Read` | `TopN`, `Score`, `Count` |
| `Maintenance` | `TrimDay`, `TrimDateRange`, `PruneTypeBefore` |
| `Store` | 嵌入上述三者；`*Board` 已实现 |

## 与全文搜索（search）

若使用 [go_utils/search](../search/) 做中文全文检索，**热词/查询串统计**可写入本库的 `member`（与 Bleve 索引分离）。说明见 [search/README.md](../search/README.md) 与 [search/zh/README.md](../search/zh/README.md)。

## 说明

- **Redis Cluster**：跨日合并需多 key 落在可执行 `ZUNIONSTORE` 的上下文；生产环境若用 Cluster，需自行保证 key 分布或使用 hash tag 等策略（本库未内置）。  
- **大区间**：日 key 过多时会触发 `ErrTooManyKeys`，可缩短查询区间或提高 `MaxUnionKeys`（需评估 Redis 负载）。  

## License

与 go_utils 项目一致。
