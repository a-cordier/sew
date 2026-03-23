package notes

import (
	"strings"
	"testing"
)

func TestRender_BasicTemplate(t *testing.T) {
	type data struct {
		Name string
	}
	out, err := Render("Hello {{ .Name }}", data{Name: "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Hello world" {
		t.Fatalf("expected %q, got %q", "Hello world", out)
	}
}

func TestRenderWithFlags_FlagPresent(t *testing.T) {
	tmpl := `{{ if hasFlag "no-portal" }}hidden{{ else }}visible{{ end }}`
	out, err := RenderWithFlags(tmpl, nil, []string{"no-portal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hidden" {
		t.Fatalf("expected %q, got %q", "hidden", out)
	}
}

func TestRenderWithFlags_FlagAbsent(t *testing.T) {
	tmpl := `{{ if hasFlag "no-portal" }}hidden{{ else }}visible{{ end }}`
	out, err := RenderWithFlags(tmpl, nil, []string{"no-es"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "visible" {
		t.Fatalf("expected %q, got %q", "visible", out)
	}
}

func TestRenderWithFlags_NoFlags(t *testing.T) {
	tmpl := `{{ if hasFlag "no-portal" }}hidden{{ else }}visible{{ end }}`
	out, err := RenderWithFlags(tmpl, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "visible" {
		t.Fatalf("expected %q, got %q", "visible", out)
	}
}

func TestRenderWithFlags_MultipleFlags(t *testing.T) {
	tmpl := `{{ if not (hasFlag "no-ui") }}Console{{ end }}{{ if not (hasFlag "no-portal") }}Portal{{ end }}`
	out, err := RenderWithFlags(tmpl, nil, []string{"no-ui"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "Console") {
		t.Fatal("expected Console hidden when no-ui active")
	}
	if !strings.Contains(out, "Portal") {
		t.Fatal("expected Portal visible when only no-ui active")
	}
}

func TestRenderWithFlags_UnknownFlagReturnsFalse(t *testing.T) {
	tmpl := `{{ if hasFlag "nonexistent" }}yes{{ else }}no{{ end }}`
	out, err := RenderWithFlags(tmpl, nil, []string{"no-portal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "no" {
		t.Fatalf("expected %q, got %q", "no", out)
	}
}

func TestRenderWithFlags_WithDotData(t *testing.T) {
	type cfg struct {
		Name string
	}
	tmpl := `{{ .Name }}{{ if not (hasFlag "no-portal") }} with portal{{ end }}`
	out, err := RenderWithFlags(tmpl, cfg{Name: "cluster"}, []string{"no-portal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "cluster" {
		t.Fatalf("expected %q, got %q", "cluster", out)
	}
}

func TestRenderWithFlags_RealisticNotesTemplate(t *testing.T) {
	tmpl := `Your cluster is ready.

{{ if not (hasFlag "no-ui") -}}
Console     http://localhost:30080
{{ end -}}
{{ if and (not (hasFlag "no-portal")) (not (hasFlag "no-ui")) -}}
Portal      http://localhost:30081
{{ end -}}
Gateway     http://localhost:30082
`

	tests := []struct {
		name     string
		flags    []string
		wantIn   []string
		wantOut  []string
	}{
		{
			name:    "no flags",
			flags:   nil,
			wantIn:  []string{"Console", "Portal", "Gateway"},
			wantOut: nil,
		},
		{
			name:    "no-portal",
			flags:   []string{"no-portal"},
			wantIn:  []string{"Console", "Gateway"},
			wantOut: []string{"Portal"},
		},
		{
			name:    "no-ui",
			flags:   []string{"no-ui"},
			wantIn:  []string{"Gateway"},
			wantOut: []string{"Console", "Portal"},
		},
		{
			name:    "no-portal and no-ui",
			flags:   []string{"no-portal", "no-ui"},
			wantIn:  []string{"Gateway"},
			wantOut: []string{"Console", "Portal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := RenderWithFlags(tmpl, nil, tt.flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, s := range tt.wantIn {
				if !strings.Contains(out, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, out)
				}
			}
			for _, s := range tt.wantOut {
				if strings.Contains(out, s) {
					t.Errorf("expected output NOT to contain %q, got:\n%s", s, out)
				}
			}
		})
	}
}
