package zh

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/mapping"
)

const (
	// AnalyzerName 为索引映射内注册的中文分析器名称，与 title/content 字段绑定。
	AnalyzerName = "zh_gse"
	// tokenizerInstanceName 为当前 mapping 内自定义 tokenizer 的实例名（仅用于 AddCustomTokenizer / AddCustomAnalyzer 互相引用）。
	tokenizerInstanceName = "zh_gse_tokenizer"
)

// MappingOption 配置中文分词索引映射。
type MappingOption func(*mappingOptions)

type mappingOptions struct {
	dictFiles  []string
	searchMode bool
	alphaNum   bool
}

// WithDict 指定 gse 词典文件路径（可多个）；不传则使用 gse 默认词典。
func WithDict(files ...string) MappingOption {
	return func(o *mappingOptions) {
		o.dictFiles = append([]string(nil), files...)
	}
}

// WithSearchMode 为 true（默认）时使用搜索分词模式（更细粒度，利于检索）；false 使用普通分词。
func WithSearchMode(search bool) MappingOption {
	return func(o *mappingOptions) {
		o.searchMode = search
	}
}

// WithAlphaNum 为 true 时对英文/数字做更细切分（对应 gse Segmenter.AlphaNum）。
func WithAlphaNum(on bool) MappingOption {
	return func(o *mappingOptions) {
		o.alphaNum = on
	}
}

// NewHighlightableMapping 创建与 search.NewHighlightableMapping 相同字段布局的索引映射：
// title、content 均启用 Store、IncludeTermVectors，并绑定 gse 中文分析器，支持 Bleve 高亮。
func NewHighlightableMapping(opts ...MappingOption) (mapping.IndexMapping, error) {
	o := &mappingOptions{searchMode: true}
	for _, fn := range opts {
		fn(o)
	}

	idxMapping := bleve.NewIndexMapping()

	tokCfg := map[string]interface{}{
		"type":         TokenizerTypeName,
		"search_mode":  o.searchMode,
		"alpha_num":    o.alphaNum,
		"dict_files":   o.dictFiles,
	}
	if err := idxMapping.AddCustomTokenizer(tokenizerInstanceName, tokCfg); err != nil {
		return nil, fmt.Errorf("add gse tokenizer: %w", err)
	}

	if err := idxMapping.AddCustomAnalyzer(AnalyzerName, map[string]interface{}{
		"type":          custom.Name,
		"tokenizer":     tokenizerInstanceName,
		"token_filters": []string{lowercase.Name},
	}); err != nil {
		return nil, fmt.Errorf("add zh_gse analyzer: %w", err)
	}

	// 无字段限定的 Match/QueryString 等使用 DefaultAnalyzer；须与 title/content 一致，否则查询仍按 standard 分词。
	idxMapping.DefaultAnalyzer = AnalyzerName

	docMapping := bleve.NewDocumentMapping()

	titleField := bleve.NewTextFieldMapping()
	titleField.Store = true
	titleField.IncludeTermVectors = true
	titleField.Analyzer = AnalyzerName
	docMapping.AddFieldMappingsAt("title", titleField)

	contentField := bleve.NewTextFieldMapping()
	contentField.Store = true
	contentField.IncludeTermVectors = true
	contentField.Analyzer = AnalyzerName
	docMapping.AddFieldMappingsAt("content", contentField)

	idxMapping.DefaultMapping = docMapping
	return idxMapping, nil
}
