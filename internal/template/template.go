package sewtmpl

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Render templates raw sew.yaml bytes by extracting the vars block,
// merging with --set overrides, and executing the document as a Go
// text/template. The returned bytes are ready for YAML unmarshaling.
func Render(raw []byte, setOverrides map[string]string) ([]byte, error) {
	vars, err := extractVars(raw)
	if err != nil {
		return nil, fmt.Errorf("extracting vars: %w", err)
	}

	for k, v := range setOverrides {
		vars[k] = v
	}

	tmpl, err := template.New("sew").
		Option("missingkey=error").
		Funcs(funcMap()).
		Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}
	return buf.Bytes(), nil
}

// VarDef describes a template variable with its default value and
// optional documentation.
type VarDef struct {
	Name        string
	Default     string
	Description string
}

// ExtractVarDefs scans raw sew.yaml bytes for a top-level vars block and
// returns structured metadata for each variable. Supports both simple
// (key: "value") and extended (key: {default: "value", description: "..."})
// forms.
func ExtractVarDefs(raw []byte) ([]VarDef, error) {
	varsYAML, ok := isolateVarsBlock(raw)
	if !ok {
		return nil, nil
	}

	var wrapper struct {
		Vars yaml.Node `yaml:"vars"`
	}
	if err := yaml.Unmarshal([]byte(varsYAML), &wrapper); err != nil {
		return nil, fmt.Errorf("parsing vars block: %w", err)
	}

	node := &wrapper.Vars
	if node.Kind == 0 || node.Kind != yaml.MappingNode {
		return nil, nil
	}

	var defs []VarDef
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		def := VarDef{Name: key.Value}
		switch val.Kind {
		case yaml.ScalarNode:
			def.Default = val.Value
		case yaml.MappingNode:
			var ext struct {
				Default     string `yaml:"default"`
				Description string `yaml:"description"`
			}
			if err := val.Decode(&ext); err != nil {
				return nil, fmt.Errorf("parsing var %q: %w", key.Value, err)
			}
			def.Default = ext.Default
			def.Description = ext.Description
		}
		defs = append(defs, def)
	}
	return defs, nil
}

// extractVars scans raw YAML bytes for a top-level vars block and returns
// a flat string map suitable for template execution. Supports both simple
// (key: "value") and extended (key: {default: "value", description: "..."})
// forms — the extended form is normalized to just key → default.
func extractVars(raw []byte) (map[string]string, error) {
	defs, err := ExtractVarDefs(raw)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(defs))
	for _, d := range defs {
		m[d.Name] = d.Default
	}
	return m, nil
}

// isolateVarsBlock finds the top-level vars: block boundaries in raw YAML
// that may contain template expressions elsewhere. Returns the slice of
// text containing only the vars block (valid YAML), and whether one was found.
func isolateVarsBlock(raw []byte) (string, bool) {
	lines := strings.Split(string(raw), "\n")

	start := -1
	end := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if start == -1 {
			if trimmed == "vars:" {
				start = i
				continue
			}
			if strings.HasPrefix(trimmed, "vars:") {
				start = i
				end = i + 1
				break
			}
			continue
		}
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && !strings.HasPrefix(trimmed, "#") {
			end = i
			break
		}
	}

	if start == -1 {
		return "", false
	}
	return strings.Join(lines[start:end], "\n"), true
}

// funcMap returns the template.FuncMap shared by all sew template
// rendering. It provides:
//
//   - env: returns the value of an environment variable.
//   - default: returns a fallback when the pipeline value is empty.
//   - required: returns the pipeline value or an error with the given message.
func funcMap() template.FuncMap {
	return template.FuncMap{
		"env": func(key string) string {
			return os.Getenv(key)
		},
		"default": func(fallback, val string) string {
			if val == "" {
				return fallback
			}
			return val
		},
		"required": func(msg, val string) (string, error) {
			if val == "" {
				return "", fmt.Errorf("%s", msg)
			}
			return val, nil
		},
	}
}
