package installer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestRenderDiff_NoChange(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderDiff("test", "a\n", "a\n", &buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output for identical inputs, got %q", buf.String())
	}
}

func TestRenderDiff_Added(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderDiff("comp", "", "line1\nline2\n", &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "comp (current)") {
		t.Error("expected header with component name (current)")
	}
	if !strings.Contains(out, "comp (proposed)") {
		t.Error("expected header with component name (proposed)")
	}
	if !strings.Contains(out, "+line1") {
		t.Error("expected added line1")
	}
	if !strings.Contains(out, "+line2") {
		t.Error("expected added line2")
	}
}

func TestRenderDiff_Removed(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderDiff("comp", "old\n", "", &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "-old") {
		t.Error("expected removed line")
	}
}

func TestRenderDiff_Changed(t *testing.T) {
	before := "line1\nline2\nline3\n"
	after := "line1\nchanged\nline3\n"
	var buf bytes.Buffer
	if err := RenderDiff("svc", before, after, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "-line2") {
		t.Error("expected removed line2")
	}
	if !strings.Contains(out, "+changed") {
		t.Error("expected added changed line")
	}
	if !strings.Contains(out, "@@") {
		t.Error("expected hunk header")
	}
}

func TestRenderDiff_Colored(t *testing.T) {
	orig := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = orig }()

	for _, c := range []*color.Color{headerColor, hunkColor, addColor, delColor} {
		c.EnableColor()
		defer c.DisableColor()
	}

	before := "a\nb\n"
	after := "a\nc\n"
	var buf bytes.Buffer
	if err := RenderDiff("x", before, after, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "\033[") {
		t.Error("expected ANSI color codes in output")
	}
}
