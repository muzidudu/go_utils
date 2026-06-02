package template

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestSitesEngine_RegisterFilter_FilterExists(t *testing.T) {
	e := New(t.TempDir(), ".django")
	if e.FilterExists("test_suffix") {
		t.Fatal("should not exist yet")
	}
	if err := e.RegisterFilter("test_suffix", filterTestSuffix); err != nil {
		t.Fatal(err)
	}
	if !e.FilterExists("test_suffix") {
		t.Fatal("pending filter should exist")
	}
	if err := e.RegisterFilter("test_suffix", filterTestSuffix); err == nil {
		t.Fatal("expected duplicate register error")
	}
}

func TestSitesEngine_ReplaceFilter(t *testing.T) {
	e := New(t.TempDir(), ".django")
	_ = e.RegisterFilter("test_suffix", filterTestSuffix)
	if err := e.ReplaceFilter("test_suffix", filterTestSuffixExclaim); err != nil {
		t.Fatal(err)
	}
	if !e.FilterExists("test_suffix") {
		t.Fatal("filter should still exist")
	}
}

func TestSitesEngine_RegisterFilterAfterLoad(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "template", "default")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "index.django"), []byte(`{{ "hi"|test_suffix }}`), 0o644); err != nil {
		t.Fatal(err)
	}
	e := New(dir, ".django")
	if err := e.RegisterFilter("test_suffix", filterTestSuffix); err != nil {
		t.Fatal(err)
	}
	if err := e.Load(); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := e.Render(&out, "index", nil); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "hi!" {
		t.Errorf("got %q", got)
	}
}

func filterTestSuffix(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.String() + "!"), nil
}

func filterTestSuffixExclaim(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.String() + "!!"), nil
}
