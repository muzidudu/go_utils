package template

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestNew_BuiltinFilters(t *testing.T) {
	e := New(t.TempDir(), ".django")
	for _, name := range []string{"suffix", "default_empty", "truncate_chars", "filesize", "number"} {
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

func applyBuiltinFilter(t *testing.T, fn pongo2.FilterFunction, in any, param any) string {
	t.Helper()
	var p *pongo2.Value
	if param == nil {
		p = pongo2.AsValue(nil)
	} else {
		p = pongo2.AsValue(param)
	}
	v, err := fn(pongo2.AsValue(in), p)
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	return v.String()
}

func TestFilterFilesize(t *testing.T) {
	cases := []struct {
		in    any
		param any
		want  string
	}{
		{0, nil, "0 B"},
		{500, nil, "500 B"},
		{1024, nil, "1 KB"},
		{1536, nil, "1.5 KB"},
		{1 << 20, nil, "1 MB"},
		{int64(1) << 30, nil, "1 GB"},
		{int64(1) << 40, nil, "1 TB"},
		{int64(1) << 50, nil, "1 PB"},
		{float64(1 << 60), nil, "1 EB"},
		{"2048", nil, "2 KB"},
		{-1024, nil, "-1 KB"},
		{1536, 2, "1.5 KB"},
		{1234, 2, "1.21 KB"},
	}
	for _, tc := range cases {
		got := applyBuiltinFilter(t, filterFilesize, tc.in, tc.param)
		if got != tc.want {
			t.Errorf("filesize(%v, %v) = %q, want %q", tc.in, tc.param, got, tc.want)
		}
	}
}

func TestFilterNumber(t *testing.T) {
	cases := []struct {
		in    any
		param any
		want  string
	}{
		{0, nil, "0"},
		{999, nil, "999"},
		{1000, nil, "1k"},
		{1500, nil, "1.5k"},
		{1_000_000, nil, "1m"},
		{1_500_000, nil, "1.5m"},
		{1_234_567, nil, "1.2m"},
		{1_000_000_000, nil, "1b"},
		{1_000_000_000_000, nil, "1t"},
		{"2500", nil, "2.5k"},
		{-1500, nil, "-1.5k"},
		{1500, 2, "1.5k"},
		{1_234_567, 2, "1.23m"},
	}
	for _, tc := range cases {
		got := applyBuiltinFilter(t, filterNumber, tc.in, tc.param)
		if got != tc.want {
			t.Errorf("number(%v, %v) = %q, want %q", tc.in, tc.param, got, tc.want)
		}
	}
}

func TestBuiltinFilters_FilesizeNumberRender(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "template", "default")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{{ size|filesize }} {{ views|number }}`
	if err := os.WriteFile(filepath.Join(templateDir, "index.django"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	e := New(dir, ".django")
	if err := e.Load(); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := e.Render(&out, "index", map[string]any{"size": 1536, "views": 1500}); err != nil {
		t.Fatal(err)
	}
	want := "1.5 KB 1.5k"
	if got := out.String(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
