package search_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/muzidudu/go_utils/search"
)

func ExampleBleveEngine() {
	path := "testdata_example.bleve"
	defer os.RemoveAll(path)

	engine, err := search.NewUsing(search.Config{Path: path})
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	// 索引
	_ = engine.IndexDoc("1", map[string]string{
		"title":   "Go 语言",
		"content": "Go 是高效的编程语言",
	})

	// 搜索
	result, _ := engine.Search(search.Match("Go"), &search.SearchOptions{Size: 10})
	fmt.Printf("Total: %d\n", result.Total)
	for _, h := range result.Hits {
		fmt.Printf("ID=%s Score=%.2f\n", h.ID, h.Score)
	}
}

func TestMatchIn(t *testing.T) {
	path := "testdata_matchin.bleve"
	defer os.RemoveAll(path)

	engine, err := search.NewUsing(search.Config{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	_ = engine.IndexDoc("1", map[string]string{"title": "Go 语言", "content": "Python 入门"})
	_ = engine.IndexDoc("2", map[string]string{"title": "Python 入门", "content": "Go 语言"})

	// 仅在 title 搜索 "Go"，应只命中 doc 1
	result, _ := engine.Search(search.MatchIn("Go", "title"), &search.SearchOptions{Size: 10})
	if result.Total != 1 || result.Hits[0].ID != "1" {
		t.Errorf("MatchIn title: expected 1 hit with ID=1, got %d hits", result.Total)
	}

	// 仅在 content 搜索 "Go"，应只命中 doc 2
	result, _ = engine.Search(search.MatchIn("Go", "content"), &search.SearchOptions{Size: 10})
	if result.Total != 1 || result.Hits[0].ID != "2" {
		t.Errorf("MatchIn content: expected 1 hit with ID=2, got %d hits", result.Total)
	}
}

func TestHighlight(t *testing.T) {
	path := "testdata_highlight.bleve"
	defer os.RemoveAll(path)

	engine, err := search.NewUsing(search.Config{
		Path:    path,
		Mapping: search.NewHighlightableMapping(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	_ = engine.IndexDoc("1", map[string]string{
		"title":   "Go 语言",
		"content": "Go 是一门简洁高效的编程语言",
	})

	result, err := engine.Search(search.Match("Go"), &search.SearchOptions{
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
		t.Fatal("expected at least 1 hit")
	}
	hit := result.Hits[0]
	if len(hit.Fragments) == 0 {
		t.Error("expected highlight fragments")
	}
	for field, frags := range hit.Fragments {
		if len(frags) > 0 && field != "" {
			t.Logf("Fragment %s: %s", field, frags[0])
		}
	}
}

func TestEngine(t *testing.T) {
	path := "testdata_engine.bleve"
	defer os.RemoveAll(path)

	engine, err := search.NewUsing(search.Config{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	// IndexDoc
	if err := engine.IndexDoc("1", map[string]string{"title": "A", "content": "hello"}); err != nil {
		t.Fatal(err)
	}

	// Batch
	if err := engine.Batch(map[string]interface{}{
		"2": map[string]string{"title": "B", "content": "world"},
		"3": map[string]string{"title": "C", "content": "hello world"},
	}); err != nil {
		t.Fatal(err)
	}

	// Search
	result, err := engine.Search(search.Match("hello"), &search.SearchOptions{Size: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total < 2 {
		t.Errorf("expected at least 2 hits, got %d", result.Total)
	}

	// Count
	n, err := engine.Count()
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("expected 3 docs, got %d", n)
	}

	// Delete
	if err := engine.Delete("3"); err != nil {
		t.Fatal(err)
	}
	n, _ = engine.Count()
	if n != 2 {
		t.Errorf("after delete expected 2 docs, got %d", n)
	}
}
