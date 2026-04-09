package zh

import (
	"fmt"

	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/go-ego/gse"
)

// TokenizerTypeName 为 Bleve registry 中注册的 tokenizer 类型名（AddCustomTokenizer 的 config["type"]）。
const TokenizerTypeName = "go_utils_search_zh_gse"

func init() {
	registry.RegisterTokenizer(TokenizerTypeName, tokenizerConstructor)
}

func tokenizerConstructor(config map[string]interface{}, _ *registry.Cache) (analysis.Tokenizer, error) {
	searchMode := true
	if v, ok := config["search_mode"].(bool); ok {
		searchMode = v
	}
	alphaNum := false
	if v, ok := config["alpha_num"].(bool); ok {
		alphaNum = v
	}

	dictFiles, err := dictFilesFromConfig(config)
	if err != nil {
		return nil, err
	}

	var seg gse.Segmenter
	if len(dictFiles) > 0 {
		seg, err = gse.New(dictFiles...)
	} else {
		seg, err = gse.New()
	}
	if err != nil {
		return nil, fmt.Errorf("gse segmenter: %w", err)
	}
	if alphaNum {
		seg.AlphaNum = true
	}

	return &gseTokenizer{seg: &seg, searchMode: searchMode}, nil
}

func dictFilesFromConfig(config map[string]interface{}) ([]string, error) {
	raw, ok := config["dict_files"]
	if !ok || raw == nil {
		// 新建 mapping 时可能省略；索引 JSON 反序列化后常为 null，均表示使用 gse 默认词典
		return nil, nil
	}
	switch v := raw.(type) {
	case []string:
		return append([]string(nil), v...), nil
	case []interface{}:
		out := make([]string, 0, len(v))
		for i, x := range v {
			s, ok := x.(string)
			if !ok {
				return nil, fmt.Errorf("dict_files[%d] must be string", i)
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("dict_files must be []string")
	}
}
