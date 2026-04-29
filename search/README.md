# go_utils/search

基于 [Bleve](https://blevesearch.com) 的全文搜索引擎库，提供索引、搜索、批量操作等封装。

## 特性

- **全文检索**：支持 Match、MatchPhrase、QueryString、Term、Prefix、Fuzzy 等查询
- **组合查询**：Conjunction(AND)、Disjunction(OR)、Bool 布尔查询
- **批量操作**：Batch 批量索引、DeleteBatch 批量删除
- **分页与排序**：From/Size 分页，Sort 多字段排序
- **Scorch 索引**：可选 scorch 存储以获得更好性能
- **中文分词（可选）**：子包 [`search/zh`](./zh/) 基于 [gse](https://github.com/go-ego/gse)，映射与 `NewHighlightableMapping` 一致并兼容高亮

## 安装

```bash
go get github.com/muzidudu/go_utils/search
```

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/muzidudu/go_utils/search"
)

func main() {
	// 创建索引（推荐使用 scorch）
	engine, err := search.NewUsing(search.Config{
		Path: "data/search.bleve",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	// 索引文档
	doc := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{
		Title:   "Go 语言入门",
		Content: "Go 是一门简洁高效的编程语言，适合构建搜索引擎。",
	}
	if err := engine.IndexDoc("1", doc); err != nil {
		log.Fatal(err)
	}

	// 搜索
	q := search.Match("Go 语言")
	result, err := engine.Search(q, &search.SearchOptions{
		Size:   10,
		Fields: []string{"*"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("命中 %d 条，耗时 %d ns\n", result.Total, result.Took)
	for _, hit := range result.Hits {
		fmt.Printf("ID=%s Score=%.2f %v\n", hit.ID, hit.Score, hit.Fields)
	}
}
```

## 创建与打开索引

```go
// 新建索引（默认存储）
engine, err := search.New(search.Config{Path: "data/search.bleve"})

// 新建索引（scorch，推荐）
engine, err := search.NewUsing(search.Config{Path: "data/search.bleve"})

// 打开已存在的索引
engine, err := search.Open("data/search.bleve")
```

## 索引文档

```go
// 单个文档
engine.IndexDoc("doc_id", struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}{
	Title:   "标题",
	Content: "内容",
})

// 批量索引
engine.Batch(map[string]interface{}{
	"1": map[string]string{"title": "A", "content": "..."},
	"2": map[string]string{"title": "B", "content": "..."},
})
```

## 查询类型

```go
// 分词匹配（默认对所有 text 字段）
q := search.Match("关键词")

// 短语匹配
q := search.MatchPhrase("精确短语")

// 查询字符串（支持 field:value 语法）
q := search.QueryString("title:golang +content:search")

// 精确词项
q := search.Term("exact")

// 前缀
q := search.Prefix("pre")

// 模糊
q := search.Fuzzy("golan")
```

## 指定字段搜索

在指定字段中搜索，使用 `*In(text, field)` 系列函数：

```go
// 仅在 title 字段搜索
q := search.MatchIn("Go", "title")

// 仅在 content 字段搜索短语
q := search.MatchPhraseIn("全文搜索", "content")

// 仅在 title 中精确匹配
q := search.TermIn("入门", "title")

// 前缀、模糊同理
q := search.PrefixIn("go", "title")
q := search.FuzzyIn("golan", "content")
```

或使用 QueryString 的 `field:value` 语法：

```go
// title 包含 golang 且 content 包含 search
q := search.QueryString("title:golang +content:search")

// 多字段 OR
q := search.QueryString("title:Go content:Go")

// AND 组合
q := search.Conjunction(search.Match("a"), search.Match("b"))

// OR 组合
q := search.Disjunction(search.Match("a"), search.Match("b"))

// 布尔查询
q := search.Bool(
	[]search.Query{search.Match("must")},
	[]search.Query{search.Match("should")},
	[]search.Query{search.Match("mustNot")},
)
```

## 搜索选项

```go
result, err := engine.Search(q, &search.SearchOptions{
	From:   0,                    // 分页偏移
	Size:   20,                   // 每页数量
	Fields: []string{"title", "content"},  // 返回字段，["*"] 表示全部
	Sort:   []string{"-_score", "created_at"},  // 排序，- 表示降序
})
```

## 高亮

要高亮匹配片段，需满足两点：

1. **索引映射**：要高亮的字段必须 `Store=true` 且 `IncludeTermVectors=true`
2. **搜索选项**：传入 `Highlight` 配置

```go
// 使用支持高亮的映射创建索引
engine, _ := search.NewUsing(search.Config{
	Path:    "data/search.bleve",
	Mapping: search.NewHighlightableMapping(),
})

// 搜索时启用高亮
result, _ := engine.Search(q, &search.SearchOptions{
	Size: 10,
	Fields: []string{"*"},
	Highlight: &search.HighlightOptions{
		Style:  "html",  // 输出 <mark> 标签，或 "ansi" 终端高亮
		Fields: []string{"title", "content"},  // 空则自动高亮所有匹配字段
	},
})

// 使用高亮片段
for _, hit := range result.Hits {
	if frags := hit.Fragments["content"]; len(frags) > 0 {
		// frags[0] 为高亮后的片段，如 "Go 是<mark>一门</mark>简洁..."
		fmt.Println(frags[0])
	}
}
```

自定义映射时，为需高亮的字段设置：

```go
fieldMapping := bleve.NewTextFieldMapping()
fieldMapping.Store = true
	fieldMapping.IncludeTermVectors = true
	docMapping.AddFieldMappingsAt("content", fieldMapping)
```

## 中文分词（search/zh）

面向中文正文时，可使用子包 `search/zh` 生成带 **gse 分词** 的索引映射（分析器名 `zh_gse`），仍通过 `search.NewUsing`、`search.Search`、高亮等 API 使用。

```go
import "github.com/muzidudu/go_utils/search/zh"

m, err := zh.NewHighlightableMapping()
if err != nil {
	log.Fatal(err)
}
engine, err := search.NewUsing(search.Config{
	Path:    "data/zh.bleve",
	Mapping: m,
})
```

**说明**：该映射已将索引的 `DefaultAnalyzer` 设为与 `title`/`content` 相同的 `zh_gse`，避免未带字段的 `Match()` 仍按英文默认分析器分词。可选参数、高亮注意与分词边界见 **[search/zh 使用说明](./zh/README.md)**。

## 删除文档

```go
engine.Delete("doc_id")
engine.DeleteBatch([]string{"1", "2", "3"})
```

## 自定义索引映射

```go
import (
	"github.com/blevesearch/bleve/v2"
	"github.com/muzidudu/go_utils/search"
)

mapping := bleve.NewIndexMapping()
// 自定义 document mapping、field mapping 等

engine, err := search.New(search.Config{
	Path:    "data/search.bleve",
	Mapping: mapping,
})
```

## API 概览

| 方法 | 说明 |
|------|------|
| `New(cfg)` | 创建索引（默认存储） |
| `NewUsing(cfg)` | 创建索引（scorch） |
| `Open(path)` | 打开已有索引 |
| `IndexDoc(id, doc)` | 索引单个文档 |
| `Batch(docs)` | 批量索引 |
| `Delete(id)` | 删除文档 |
| `DeleteBatch(ids)` | 批量删除 |
| `Search(q, opts)` | 执行搜索 |
| `Count()` | 文档总数 |
| `Close()` | 关闭索引 |

## 依赖

- `github.com/blevesearch/bleve/v2`
- 使用中文分词子包时另需 `github.com/go-ego/gse`（已作为 `search` 模块传递依赖）

## 同仓库：Redis 排行榜（rank）

本模块负责 **Bleve 全文索引与搜索**；若还需要按 **搜索热词、浏览量** 等做**分日/分区间、可清理** 的统计排行，请使用同仓库中的 **[go_utils/rank](../rank/)**（基于 Redis `ZSET`，与本文职责分离）。

- **安装**：`go get github.com/muzidudu/go_utils/rank`
- **能力概要**：`rank.Store` 接口涵盖按自然日计分、多日合并 TopN、按 `keepTop` 裁剪、按 `before` 日期扫描清理等；时区、key 前缀由 `rank.Config` 配置。

## License

与 go_utils 项目一致。
