package zh_test

import (
	"fmt"
	"os"

	"github.com/muzidudu/go_utils/search"
	"github.com/muzidudu/go_utils/search/zh"
)

func ExampleNewHighlightableMapping() {
	path := "testdata_example_zh.bleve"
	defer os.RemoveAll(path)

	m, err := zh.NewHighlightableMapping()
	if err != nil {
		panic(err)
	}

	engine, err := search.NewUsing(search.Config{Path: path, Mapping: m})
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	_ = engine.IndexDoc("1", map[string]string{
		"title":   "北京大学",
		"content": "简介。",
	})

	result, _ := engine.Search(search.Match("北京大学"), &search.SearchOptions{
		Size:   10,
		Fields: []string{"*"},
		Highlight: &search.HighlightOptions{
			Style:  "html",
			Fields: []string{"title"},
		},
	})
	fmt.Printf("hits=%d\n", result.Total)
	if result.Total > 0 && len(result.Hits[0].Fragments) > 0 {
		fmt.Println("highlight_ok")
	}
	// Output:
	// hits=1
	// highlight_ok
}
