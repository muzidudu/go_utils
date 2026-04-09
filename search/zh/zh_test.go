package zh_test

import (
	"os"
	"testing"

	"github.com/muzidudu/go_utils/search"
	"github.com/muzidudu/go_utils/search/zh"
)

func TestHighlightableMappingChinese(t *testing.T) {
	path := "testdata_zh_highlight.bleve"
	defer os.RemoveAll(path)

	m, err := zh.NewHighlightableMapping()
	if err != nil {
		t.Fatal(err)
	}

	engine, err := search.NewUsing(search.Config{Path: path, Mapping: m})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	// 短串「北京大学」索引为「北京」「大学」；与查询「北京大学」分词一致。长串里「北京大学」常作为整词，与查询词项可能不一致。
	if err := engine.IndexDoc("1", map[string]string{
		"title":   "北京大学",
		"content": "北京大学开设有多门生物学相关课程。",
	}); err != nil {
		t.Fatal(err)
	}

	result, err := engine.Search(search.Match("北京大学"), &search.SearchOptions{
		Size:   10,
		Fields: []string{"*"},
		Highlight: &search.HighlightOptions{
			Style:  "html",
			Fields: []string{"title", "content"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total == 0 {
		t.Fatal("expected at least one hit for 北京大学")
	}
	hit := result.Hits[0]
	if len(hit.Fragments) == 0 {
		t.Fatal("expected highlight fragments with zh_gse mapping")
	}
}

func TestMatchInMixedTitle(t *testing.T) {
	path := "testdata_zh_matchin.bleve"
	defer os.RemoveAll(path)

	m, err := zh.NewHighlightableMapping()
	if err != nil {
		t.Fatal(err)
	}
	engine, err := search.NewUsing(search.Config{Path: path, Mapping: m})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	_ = engine.IndexDoc("1", map[string]string{"title": "Go 语言指南", "content": "其他"})
	_ = engine.IndexDoc("2", map[string]string{"title": "Rust 指南", "content": "Go 语言"})

	result, _ := engine.Search(search.MatchIn("go", "title"), &search.SearchOptions{Size: 10})
	if result.Total != 1 {
		t.Fatalf("MatchIn title: want 1 hit, got total=%d", result.Total)
	}
	if result.Hits[0].ID != "1" {
		t.Fatalf("MatchIn title: want doc 1, got %s", result.Hits[0].ID)
	}
}
