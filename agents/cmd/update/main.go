package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type agentFrontmatter struct {
	Product string   `yaml:"product"`
	Paths   []string `yaml:"paths"`
}

const (
	beginMarker = "<!-- agents:begin -->"
	endMarker   = "<!-- agents:end -->"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "agents:update: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	entries, err := filepath.Glob("agents/*-agent.md")
	if err != nil {
		return err
	}
	sort.Strings(entries)

	var rows []string
	for _, path := range entries {
		fm, err := parseFrontmatter(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		paths := "`" + strings.Join(fm.Paths, "`, `") + "`"
		rows = append(rows, fmt.Sprintf("| %s | [%s](%s) | %s |", fm.Product, path, path, paths))
	}

	var section bytes.Buffer
	section.WriteString("| Product | Instructions | Applies to |\n")
	section.WriteString("|---|---|---|\n")
	for _, row := range rows {
		section.WriteString(row)
		section.WriteByte('\n')
	}

	agentsFile := "AGENTS.md"
	content, err := os.ReadFile(agentsFile)
	if err != nil {
		return err
	}

	text := string(content)
	beginIdx := strings.Index(text, beginMarker)
	endIdx := strings.Index(text, endMarker)
	if beginIdx < 0 || endIdx < 0 || endIdx <= beginIdx {
		return fmt.Errorf("missing %s / %s markers in %s", beginMarker, endMarker, agentsFile)
	}

	var out strings.Builder
	out.WriteString(text[:beginIdx+len(beginMarker)])
	out.WriteByte('\n')
	out.WriteString(section.String())
	out.WriteString(text[endIdx:])

	return os.WriteFile(agentsFile, []byte(out.String()), 0644)
}

func parseFrontmatter(path string) (agentFrontmatter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return agentFrontmatter{}, err
	}

	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return agentFrontmatter{}, fmt.Errorf("no YAML frontmatter found")
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return agentFrontmatter{}, fmt.Errorf("unterminated frontmatter")
	}
	fmBlock := text[4 : 4+end]

	var fm agentFrontmatter
	if err := yaml.Unmarshal([]byte(fmBlock), &fm); err != nil {
		return agentFrontmatter{}, err
	}
	if fm.Product == "" {
		return agentFrontmatter{}, fmt.Errorf("missing 'product' in frontmatter")
	}
	if len(fm.Paths) == 0 {
		return agentFrontmatter{}, fmt.Errorf("missing 'paths' in frontmatter")
	}
	return fm, nil
}
