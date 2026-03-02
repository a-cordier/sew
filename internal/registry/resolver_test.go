package registry

import (
	"path/filepath"
	"testing"
)

func TestNewResolver_FileURL(t *testing.T) {
	r := NewResolver("file:///some/path", "/sew-home")
	fs, ok := r.(*FSResolver)
	if !ok {
		t.Fatal("expected FSResolver for file:// URL")
	}
	if fs.Root != "/some/path" {
		t.Fatalf("expected root %q, got %q", "/some/path", fs.Root)
	}
	if fs.SewHome != "/sew-home" {
		t.Fatalf("expected SewHome %q, got %q", "/sew-home", fs.SewHome)
	}
}

func TestNewResolver_HTTPURL(t *testing.T) {
	r := NewResolver("https://example.com/registry", "/sew-home")
	h, ok := r.(*HTTPResolver)
	if !ok {
		t.Fatal("expected HTTPResolver for HTTP URL")
	}
	if h.BaseURL != "https://example.com/registry" {
		t.Fatalf("expected base URL %q, got %q", "https://example.com/registry", h.BaseURL)
	}
	expectedCache := filepath.Join("/sew-home", "cache")
	if h.CacheRoot != expectedCache {
		t.Fatalf("expected cache root %q, got %q", expectedCache, h.CacheRoot)
	}
	if h.SewHome != "/sew-home" {
		t.Fatalf("expected SewHome %q, got %q", "/sew-home", h.SewHome)
	}
}
