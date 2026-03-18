package search

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
)

// IndexMapping 索引映射类型别名
type IndexMapping = mapping.IndexMappingImpl

// BleveEngine 基于 Bleve 的搜索引擎实现
type BleveEngine struct {
	index bleve.Index
}

// Config 引擎配置
type Config struct {
	// Path 索引存储路径，如 "data/search.bleve"
	Path string
	// Mapping 自定义索引映射，nil 使用默认
	Mapping mapping.IndexMapping
}

// New 创建 Bleve 搜索引擎（新建索引）
func New(cfg Config) (*BleveEngine, error) {
	mapping := cfg.Mapping
	if mapping == nil {
		mapping = bleve.NewIndexMapping()
	}
	index, err := bleve.New(cfg.Path, mapping)
	if err != nil {
		return nil, fmt.Errorf("create bleve index: %w", err)
	}
	return &BleveEngine{index: index}, nil
}

// Open 打开已存在的索引
func Open(path string) (*BleveEngine, error) {
	index, err := bleve.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open bleve index: %w", err)
	}
	return &BleveEngine{index: index}, nil
}

// NewUsing 使用指定类型创建索引（推荐 scorch 以获得更好性能）
func NewUsing(cfg Config) (*BleveEngine, error) {
	mapping := cfg.Mapping
	if mapping == nil {
		mapping = bleve.NewIndexMapping()
	}
	index, err := bleve.NewUsing(cfg.Path, mapping, "scorch", "scorch", nil)
	if err != nil {
		return nil, fmt.Errorf("create bleve index: %w", err)
	}
	return &BleveEngine{index: index}, nil
}

// IndexDoc 索引单个文档
func (e *BleveEngine) IndexDoc(id string, doc interface{}) error {
	return e.index.Index(id, doc)
}

// Batch 批量索引
func (e *BleveEngine) Batch(docs map[string]interface{}) error {
	batch := e.index.NewBatch()
	for id, doc := range docs {
		if err := batch.Index(id, doc); err != nil {
			return fmt.Errorf("batch index %s: %w", id, err)
		}
	}
	return e.index.Batch(batch)
}

// Delete 删除文档
func (e *BleveEngine) Delete(id string) error {
	return e.index.Delete(id)
}

// DeleteBatch 批量删除
func (e *BleveEngine) DeleteBatch(ids []string) error {
	batch := e.index.NewBatch()
	for _, id := range ids {
		batch.Delete(id)
	}
	return e.index.Batch(batch)
}

// Search 执行搜索
func (e *BleveEngine) Search(q query.Query, opts *SearchOptions) (*SearchResult, error) {
	searchReq := bleve.NewSearchRequest(q)
	if opts != nil {
		if opts.From > 0 {
			searchReq.From = opts.From
		}
		if opts.Size > 0 {
			searchReq.Size = opts.Size
		}
		if len(opts.Fields) > 0 {
			searchReq.Fields = opts.Fields
		}
		if len(opts.Sort) > 0 {
			searchReq.SortBy(opts.Sort)
		}
		if opts.Highlight != nil {
			searchReq.Highlight = bleve.NewHighlight()
			if opts.Highlight.Style != "" {
				searchReq.Highlight = bleve.NewHighlightWithStyle(opts.Highlight.Style)
			}
			for _, f := range opts.Highlight.Fields {
				searchReq.Highlight.AddField(f)
			}
		}
	}
	if searchReq.Size == 0 {
		searchReq.Size = 10
	}

	searchResults, err := e.index.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	result := &SearchResult{
		Total: searchResults.Total,
		Took:  searchResults.Took.Nanoseconds(),
		Hits:  make([]*Hit, 0, len(searchResults.Hits)),
	}
	for _, h := range searchResults.Hits {
		if h.Fields == nil {
			h.Fields = make(map[string]interface{})
		}
		hit := &Hit{
			ID:     h.ID,
			Score:  h.Score,
			Fields: h.Fields,
		}
		if len(h.Fragments) > 0 {
			hit.Fragments = h.Fragments
		}
		result.Hits = append(result.Hits, hit)
	}
	return result, nil
}

// Count 统计文档总数
func (e *BleveEngine) Count() (uint64, error) {
	return e.index.DocCount()
}

// Close 关闭索引
func (e *BleveEngine) Close() error {
	return e.index.Close()
}

// BleveIndex 返回底层索引（用于高级用法）
func (e *BleveEngine) BleveIndex() bleve.Index {
	return e.index
}
