# go_utils/search/zh

在 [Bleve](https://blevesearch.com) 上使用 [gse](https://github.com/go-ego/gse) 做中文分词的索引映射，与父包 [`search`](../) 的 `BleveEngine`、`Search`、`Highlight` 等 API 完全兼容；字段布局与 `search.NewHighlightableMapping` 一致（`title` / `content`，`Store` + `IncludeTermVectors`），可直接用于全文高亮。

## 安装

父模块已依赖 `github.com/go-ego/gse`，仅需拉取 `search`：

```bash
go get github.com/muzidudu/go_utils/search
```

## 快速开始

```go
package main

import (
	"fmt"
	"log"

	"github.com/muzidudu/go_utils/search"
	"github.com/muzidudu/go_utils/search/zh"
)

func main() {
	m, err := zh.NewHighlightableMapping()
	if err != nil {
		log.Fatal(err)
	}

	engine, err := search.NewUsing(search.Config{
		Path:    "data/zh.bleve",
		Mapping: m,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	if err := engine.IndexDoc("1", map[string]string{
		"title":   "北京大学",
		"content": "简介与招生说明。",
	}); err != nil {
		log.Fatal(err)
	}

	q := search.Match("北京大学")
	result, err := engine.Search(q, &search.SearchOptions{
		Size:   10,
		Fields: []string{"*"},
		Highlight: &search.HighlightOptions{
			Style:  "html",
			Fields: []string{"title", "content"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("命中 %d 条\n", result.Total)
	for _, hit := range result.Hits {
		fmt.Println(hit.Fragments)
	}
}
```

## 分析器与默认行为

- 映射内注册的分析器名称为 **`zh_gse`**（常量 `zh.AnalyzerName`）。
- **`title`、`content` 均绑定 `zh_gse`**，并开启 `Store`、`IncludeTermVectors`，满足 Bleve 高亮要求。
- 索引映射的 **`DefaultAnalyzer` 同样设为 `zh_gse`**，因此未指定字段的 `search.Match`、`QueryString` 等与正文使用同一套分词；若改用仅 `standard` 的默认分析器，中文查询词项会与索引不一致，容易出现无结果。

## 映射选项

在 `zh.NewHighlightableMapping(opts...)` 中传入：

| 选项 | 说明 |
|------|------|
| `zh.WithDict(files ...string)` | 指定 gse 词典文件路径（可多个）；不传则使用 gse 自带默认词典。 |
| `zh.WithSearchMode(true)` | **默认 `true`**：使用偏「搜索」的细粒度分词（内部对应 gse 的搜索模式）。设为 `false` 时使用普通分词模式。 |
| `zh.WithAlphaNum(true)` | 对英文、数字等做更细切分（对应 gse `Segmenter.AlphaNum`）。默认 `false`。 |

示例：

```go
m, err := zh.NewHighlightableMapping(
	zh.WithDict("/path/to/user_dict.txt"),
	zh.WithSearchMode(true),
	zh.WithAlphaNum(true),
)
```

词典加载与路径规则以 [gse 文档](https://github.com/go-ego/gse) 为准；首次运行默认词典时，gse 可能向标准输出打印加载日志。

## 高亮

与父包相同：在 `SearchOptions` 中设置 `Highlight`，并对需要展示的字段开启片段即可（见 [`search`  README 高亮章节](../#高亮)）。使用 `zh.NewHighlightableMapping` 时已满足「可高亮字段」的映射条件。

## 查询与分词边界说明

1. **同一查询串在短文本与长句中的分词可能不同**（例如单独出现的「北京大学」与长句中的「北京大学」被切成不同词项组合）。若短语级命中更重要，可配合 `search.MatchPhrase` / `MatchPhraseIn`，或调整文案与查询策略。
2. **中英混合**：分析器在分词后会对词面做 **小写规范化**（`to_lower`），便于英文检索；中文词不受影响。
3. **更换分析器**（例如在 `standard` 与 `zh_gse` 之间切换）后，需 **删除旧索引目录并全量重建索引**，不可直接复用原路径下的索引文件。

## 依赖

- `github.com/blevesearch/bleve/v2`
- `github.com/go-ego/gse`

## License

与 go_utils 项目一致。
