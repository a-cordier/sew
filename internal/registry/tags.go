package registry

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// TagDefinition describes a single allowed tag in a registry vocabulary.
type TagDefinition struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// LoadTags reads a tags vocabulary file and returns the allowed tag set.
func LoadTags(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading tags file: %w", err)
	}

	var defs []TagDefinition
	if err := yaml.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("parsing tags file: %w", err)
	}
	if len(defs) == 0 {
		return nil, fmt.Errorf("tags file %s is empty", path)
	}

	allowed := make(map[string]bool, len(defs))
	for _, d := range defs {
		if d.Name == "" {
			return nil, fmt.Errorf("tags file contains entry with empty name")
		}
		allowed[d.Name] = true
	}
	return allowed, nil
}

// readmeFrontmatter holds the fields we care about from README YAML frontmatter.
type readmeFrontmatter struct {
	Tags []string `yaml:"tags"`
}

// ValidateReadmeTags checks that every tag in a README's YAML frontmatter
// belongs to the allowed set. Returns an error listing all invalid tags.
func ValidateReadmeTags(readmePath string, allowed map[string]bool) error {
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("reading README: %w", err)
	}

	fm, err := extractFrontmatter(data)
	if err != nil {
		return err
	}
	if fm == nil {
		return nil
	}

	var invalid []string
	for _, tag := range fm.Tags {
		if !allowed[tag] {
			invalid = append(invalid, tag)
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("invalid tag(s): %s", strings.Join(invalid, ", "))
	}
	return nil
}

// extractFrontmatter parses YAML frontmatter delimited by "---" lines.
// Returns nil (no error) when no frontmatter is present.
func extractFrontmatter(data []byte) (*readmeFrontmatter, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	if !scanner.Scan() {
		return nil, nil
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return nil, nil
	}

	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			var fm readmeFrontmatter
			if err := yaml.Unmarshal(buf.Bytes(), &fm); err != nil {
				return nil, fmt.Errorf("parsing frontmatter: %w", err)
			}
			return &fm, nil
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	return nil, fmt.Errorf("unterminated frontmatter (missing closing ---)")
}
