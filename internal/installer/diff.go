package installer

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	headerColor = color.New(color.Bold)
	hunkColor   = color.New(color.FgCyan)
	addColor    = color.New(color.FgGreen)
	delColor    = color.New(color.FgRed)
)

// RenderDiff writes a colored unified diff of before/after to w.
// It silently returns if before and after are identical.
func RenderDiff(name, before, after string, w io.Writer) error {
	if before == after {
		return nil
	}

	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(before),
		B:        difflib.SplitLines(after),
		FromFile: name + " (current)",
		ToFile:   name + " (proposed)",
		Context:  3,
	})
	if err != nil {
		return err
	}

	for _, line := range strings.SplitAfter(diff, "\n") {
		if line == "" {
			continue
		}
		var c *color.Color
		switch {
		case strings.HasPrefix(line, "---"), strings.HasPrefix(line, "+++"):
			c = headerColor
		case strings.HasPrefix(line, "@@"):
			c = hunkColor
		case strings.HasPrefix(line, "-"):
			c = delColor
		case strings.HasPrefix(line, "+"):
			c = addColor
		}
		if c != nil {
			fmt.Fprint(w, c.Sprint(line))
		} else {
			fmt.Fprint(w, line)
		}
	}
	return nil
}
