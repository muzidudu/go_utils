package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

// NewHighlightableMapping 创建支持高亮的索引映射
// 返回的 mapping 中，title 和 content 字段已设置 Store=true、IncludeTermVectors=true
func NewHighlightableMapping() mapping.IndexMapping {
	idxMapping := bleve.NewIndexMapping()
	docMapping := bleve.NewDocumentMapping()

	// title 字段：可存储、可高亮
	titleField := bleve.NewTextFieldMapping()
	titleField.Store = true
	titleField.IncludeTermVectors = true
	docMapping.AddFieldMappingsAt("title", titleField)

	// content 字段：可存储、可高亮
	contentField := bleve.NewTextFieldMapping()
	contentField.Store = true
	contentField.IncludeTermVectors = true
	docMapping.AddFieldMappingsAt("content", contentField)

	idxMapping.DefaultMapping = docMapping
	return idxMapping
}
