// Package search 基于 Bleve 的全文搜索引擎库
package search

import (
	"github.com/blevesearch/bleve/v2/search/query"
)

// Engine 搜索引擎接口
type Engine interface {
	// IndexDoc 索引单个文档，id 为文档唯一标识
	IndexDoc(id string, doc interface{}) error
	// Batch 批量索引
	Batch(docs map[string]interface{}) error
	// Delete 删除文档
	Delete(id string) error
	// DeleteBatch 批量删除
	DeleteBatch(ids []string) error
	// Search 执行搜索
	Search(q query.Query, opts *SearchOptions) (*SearchResult, error)
	// Count 统计文档总数
	Count() (uint64, error)
	// Close 关闭索引
	Close() error
}

// Query 查询类型别名，便于使用
type Query = query.Query

// SearchOptions 搜索选项
type SearchOptions struct {
	From   int      // 分页偏移，默认 0
	Size   int      // 每页数量，默认 10
	Fields []string // 返回字段，空表示全部
	Sort   []string // 排序字段，如 ["-score", "created_at"]

	// Highlight 高亮配置，nil 表示不高亮
	// 注意：要高亮的字段必须在索引时设置 Store=true 且 IncludeTermVectors=true
	Highlight *HighlightOptions
}

// HighlightOptions 高亮选项
type HighlightOptions struct {
	// Style 高亮样式："html"（默认，输出 <mark> 标签）或 "ansi"
	Style string
	// Fields 要高亮的字段，空表示自动高亮所有匹配字段
	Fields []string
}

// SearchResult 搜索结果
type SearchResult struct {
	Total uint64
	Hits  []*Hit
	Took  int64 // 耗时（纳秒）
}

// Hit 单条命中
type Hit struct {
	ID        string
	Score     float64
	Fields    map[string]interface{}
	Fragments map[string][]string // 高亮片段，key 为字段名，value 为高亮后的文本片段
}
