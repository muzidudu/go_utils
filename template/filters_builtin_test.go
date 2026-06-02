package template

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNew_BuiltinFilters(t *testing.T) {
	e := New(t.TempDir(), ".django")
	for _, name := range []string{"suffix", "default_empty", "truncate_chars"} {
		if !e.FilterExists(name) {
			t.Errorf("builtin filter %q should exist after New", name)
		}
	}
}

func TestBuiltinFilters_Render(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "template", "default")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{{ "hi"|suffix }} {{ ""|default_empty:"匿名" }} {{ "abcdefghij"|truncate_chars:5 }}`
	if err := os.WriteFile(filepath.Join(templateDir, "index.django"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	e := New(dir, ".django")
	if err := e.Load(); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := e.Render(&out, "index", nil); err != nil {
		t.Fatal(err)
	}
	want := "hi! 匿名 ab..."
	if got := out.String(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
