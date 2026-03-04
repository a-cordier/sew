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
