package cache

import (
	"strings"
	"testing"
)

func TestBuildKey(t *testing.T) {
	if got := BuildKey("app", 1, "x"); got != "app:1:x" {
		t.Errorf("BuildKey = %q", got)
	}
	if got := BuildKey("ns"); got != "ns" {
		t.Errorf("BuildKey no parts = %q", got)
	}
	if got := BuildKey("", "a"); got != ":a" {
		t.Errorf("BuildKey empty prefix = %q", got)
	}
}

func TestBuildQueryKey(t *testing.T) {
	k1 := BuildQueryKey("list", map[string]string{"a": "1", "b": "2"})
	k2 := BuildQueryKey("list", map[string]string{"b": "2", "a": "1"})
	if k1 != k2 {
		t.Errorf("same params different map order should match: %q vs %q", k1, k2)
	}
	if !strings.HasPrefix(k1, "list:query:") {
		t.Errorf("expected prefix list:query:, got %q", k1)
	}
}

func TestBuildQueryKey_MarshalError(t *testing.T) {
	ch := make(chan int)
	k := BuildQueryKey("p", ch)
	if !strings.HasPrefix(k, "p:query:") {
		t.Errorf("expected fallback key, got %q", k)
	}
}
