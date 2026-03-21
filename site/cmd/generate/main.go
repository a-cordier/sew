package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type sewConfig struct {
	Abstract   bool           `yaml:"abstract"`
	From       []string       `yaml:"from"`
	Components []sewComponent `yaml:"components"`
}

type sewComponent struct {
	Name string `yaml:"name"`
}

type readmeFrontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

type componentPage struct {
	Title       string   `yaml:"title"`
	Layout      string   `yaml:"layout"`
	Path        string   `yaml:"path"`
	Context     bool     `yaml:"context"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	From        []string `yaml:"from,omitempty"`
	Components  []string `yaml:"components,omitempty"`
	NotesCreate string   `yaml:"notes_create,omitempty"`
	Type        string   `yaml:"type"`
}

type sectionPage struct {
	Title          string `yaml:"title"`
	DefaultVariant string `yaml:"default_variant,omitempty"`
	Type           string `yaml:"type"`
}

func main() {
	registryDir := "registry"
	contentDir := "site/content/registry"
	staticDir := "site/static"

	if err := os.RemoveAll(contentDir); err != nil {
		fatalf("clean %s: %v", contentDir, err)
	}
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		fatalf("create %s: %v", contentDir, err)
	}

	writeRegistryRoot(contentDir)

	configs := map[string]*sewConfig{}
	componentDirs := map[string]bool{}
	intermediateDirs := map[string]bool{}

	err := filepath.WalkDir(registryDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "sew.yaml" {
			return nil
		}

		dir := filepath.Dir(path)
		relDir, err := filepath.Rel(registryDir, dir)
		if err != nil {
			return fmt.Errorf("relative path for %s: %w", dir, err)
		}

		config, err := parseSewConfig(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		configs[relDir] = config

		if config.Abstract {
			fmt.Printf("skip abstract: %s\n", relDir)
			return nil
		}

		componentDirs[relDir] = true
		collectIntermediateDirs(relDir, intermediateDirs)
		return nil
	})
	if err != nil {
		fatalf("walk registry: %v", err)
	}

	for relDir := range componentDirs {
		config := configs[relDir]
		dir := filepath.Join(registryDir, relDir)

		readmeTitle, description, tags, body := parseReadme(filepath.Join(dir, "README.md"))

		title := relDir
		if readmeTitle != "" {
			title = readmeTitle
		}
		notesCreate := readOptionalFile(filepath.Join(dir, "notes.create"))
		notesCreate = strings.ReplaceAll(notesCreate, "{{ .Kind.Name }}", relDir)

		page := componentPage{
			Title:       title,
			Layout:      "detail",
			Path:        relDir,
			Context:     true,
			Description: description,
			Tags:        tags,
			From:        config.From,
			Components:  resolveComponents(relDir, configs),
			NotesCreate: notesCreate,
			Type:        "registry",
		}

		outDir := filepath.Join(contentDir, relDir)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fatalf("mkdir %s: %v", outDir, err)
		}
		if err := writePage(filepath.Join(outDir, "_index.md"), page, body); err != nil {
			fatalf("write page %s: %v", relDir, err)
		}

		fmt.Printf("generated: %s\n", relDir)
	}

	for dir := range intermediateDirs {
		if componentDirs[dir] {
			continue
		}

		defaultVariant := readOptionalFile(filepath.Join(registryDir, dir, ".default"))

		page := sectionPage{
			Title:          dir,
			DefaultVariant: defaultVariant,
			Type:           "registry",
		}

		outDir := filepath.Join(contentDir, dir)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fatalf("mkdir intermediate %s: %v", outDir, err)
		}
		if err := writePage(filepath.Join(outDir, "_index.md"), page, ""); err != nil {
			fatalf("write intermediate %s: %v", dir, err)
		}

		fmt.Printf("generated section: %s\n", dir)
	}

	if err := copyToStatic(registryDir, staticDir); err != nil {
		fatalf("copy to static: %v", err)
	}

	generateSchemaDoc("schema/sew.schema.yaml", "site/content/docs/reference/configuration.md")

	fmt.Println("done")
}

func resolveComponents(relDir string, configs map[string]*sewConfig) []string {
	seen := map[string]bool{}
	var result []string
	var walk func(string)
	walk = func(dir string) {
		config, ok := configs[dir]
		if !ok {
			return
		}
		for _, parent := range config.From {
			walk(parent)
		}
		for _, c := range config.Components {
			if !seen[c.Name] {
				seen[c.Name] = true
				result = append(result, c.Name)
			}
		}
	}
	walk(relDir)
	return result
}

func writeRegistryRoot(contentDir string) {
	page := sectionPage{
		Title: "Registry",
		Type:  "registry",
	}
	if err := writePage(filepath.Join(contentDir, "_index.md"), page, "Browse available components and application stacks."); err != nil {
		fatalf("write registry root: %v", err)
	}
}

func parseSewConfig(path string) (*sewConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config sewConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func parseReadme(path string) (title string, description string, tags []string, body string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", nil, ""
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		return "", "", nil, strings.TrimSpace(content)
	}

	rest := content[4:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return "", "", nil, strings.TrimSpace(content)
	}

	var fm readmeFrontmatter
	if err := yaml.Unmarshal([]byte(rest[:idx]), &fm); err != nil {
		return "", "", nil, strings.TrimSpace(content)
	}

	bodyStart := idx + 4
	if bodyStart < len(rest) {
		body = strings.TrimSpace(rest[bodyStart:])
	}

	body = stripLeadingH1(body)

	return fm.Title, fm.Description, fm.Tags, body
}

func stripLeadingH1(body string) string {
	if !strings.HasPrefix(body, "# ") {
		return body
	}
	if nl := strings.Index(body, "\n"); nl >= 0 {
		return strings.TrimSpace(body[nl+1:])
	}
	return ""
}

func readOptionalFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func collectIntermediateDirs(relDir string, intermediates map[string]bool) {
	parts := strings.Split(relDir, string(filepath.Separator))
	for i := 1; i < len(parts); i++ {
		intermediates[filepath.Join(parts[:i]...)] = true
	}
}

func writePage(path string, frontmatter any, body string) error {
	var buf bytes.Buffer
	buf.WriteString("---\n")

	fmBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}
	buf.Write(fmBytes)
	buf.WriteString("---\n")

	if body != "" {
		buf.WriteString("\n")
		buf.WriteString(body)
		buf.WriteString("\n")
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

func copyToStatic(registryDir, staticDir string) error {
	return filepath.WalkDir(registryDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(registryDir, path)
		if err != nil {
			return err
		}

		dest := filepath.Join(staticDir, relPath)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fmt.Printf("static: %s\n", relPath)
		return os.WriteFile(dest, data, 0644)
	})
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
