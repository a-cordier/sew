package notes

import (
	"bytes"
	"text/template"
)

// Render executes templateContent as a Go text/template with data as the
// dot context. It returns the rendered string or an error if parsing or
// execution fails.
func Render(templateContent string, data any) (string, error) {
	tmpl, err := template.New("notes").Parse(templateContent)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderWithFlags is like Render but registers a hasFlag template function
// that returns true when the named context flag was activated on the CLI.
// This lets notes.create templates conditionally show content based on
// which flags the user passed (e.g. {{ if not (hasFlag "no-portal") }}).
func RenderWithFlags(templateContent string, data any, activeFlags []string) (string, error) {
	flagSet := make(map[string]bool, len(activeFlags))
	for _, f := range activeFlags {
		flagSet[f] = true
	}
	funcMap := template.FuncMap{
		"hasFlag": func(name string) bool { return flagSet[name] },
	}
	tmpl, err := template.New("notes").Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
