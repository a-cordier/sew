package registry

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTags_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tags.yaml")
	writeFile(t, path, `
- name: database
  description: Persistent data stores
- name: networking
  description: API gateways
`)

	allowed, err := LoadTags(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(allowed) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(allowed))
	}
	if !allowed["database"] {
		t.Fatal("expected database to be allowed")
	}
	if !allowed["networking"] {
		t.Fatal("expected networking to be allowed")
	}
}

func TestLoadTags_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tags.yaml")
	writeFile(t, path, `[]`)

	_, err := LoadTags(path)
	if err == nil {
		t.Fatal("expected error for empty tags file")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected 'empty' in error, got: %v", err)
	}
}

func TestLoadTags_MissingFile(t *testing.T) {
	_, err := LoadTags("/nonexistent/tags.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadTags_EmptyName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tags.yaml")
	writeFile(t, path, `
- name: ""
  description: Bad entry
`)

	_, err := LoadTags(path)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !strings.Contains(err.Error(), "empty name") {
		t.Fatalf("expected 'empty name' in error, got: %v", err)
	}
}

func TestValidateReadmeTags_AllValid(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, `---
title: "Test"
tags: [database, networking]
---

# Test
`)

	allowed := map[string]bool{"database": true, "networking": true, "security": true}
	if err := ValidateReadmeTags(readme, allowed); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateReadmeTags_InvalidTag(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, `---
title: "Test"
tags: [database, cache, gateway]
---

# Test
`)

	allowed := map[string]bool{"database": true, "networking": true}
	err := ValidateReadmeTags(readme, allowed)
	if err == nil {
		t.Fatal("expected error for invalid tags")
	}
	if !strings.Contains(err.Error(), "cache") {
		t.Fatalf("expected 'cache' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "gateway") {
		t.Fatalf("expected 'gateway' in error, got: %v", err)
	}
}

func TestValidateReadmeTags_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, `# Just a heading

No frontmatter here.
`)

	allowed := map[string]bool{"database": true}
	if err := ValidateReadmeTags(readme, allowed); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateReadmeTags_NoTags(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, `---
title: "Test"
---

# Test
`)

	allowed := map[string]bool{"database": true}
	if err := ValidateReadmeTags(readme, allowed); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateReadmeTags_UnterminatedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, `---
title: "Test"
tags: [database]
`)

	allowed := map[string]bool{"database": true}
	err := ValidateReadmeTags(readme, allowed)
	if err == nil {
		t.Fatal("expected error for unterminated frontmatter")
	}
	if !strings.Contains(err.Error(), "unterminated") {
		t.Fatalf("expected 'unterminated' in error, got: %v", err)
	}
}

func TestExtractFrontmatter_EmptyFile(t *testing.T) {
	fm, err := extractFrontmatter([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm != nil {
		t.Fatal("expected nil frontmatter for empty file")
	}
}
