// Package zh 提供基于 gse 中文分词的 Bleve 索引映射，字段布局与 search.NewHighlightableMapping 一致（Store、IncludeTermVectors），可与 search 包的高亮选项配合使用。
//
// 使用本包创建的索引时，mapping 已将 DefaultAnalyzer 设为 zh_gse，无字段限定的 Match/QueryString 与 title、content 使用同一套分词。
package zh
