package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FSResolver resolves contexts from a local filesystem directory.
type FSResolver struct {
	Root string // absolute path to the registry root
}

// Resolve reads {Root}/{contextPath}/context.yaml and returns a
// ResolvedContext whose Dir points to the context directory on disk.
// Values files are already local, so no downloads are needed.
func (r *FSResolver) Resolve(_ context.Context, contextPath string) (*ResolvedContext, error) {
	dir := filepath.Join(r.Root, contextPath)
	data, err := os.ReadFile(filepath.Join(dir, "context.yaml"))
	if err != nil {
		return nil, fmt.Errorf("reading context file: %w", err)
	}

	var ctx Context
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("parsing context.yaml: %w", err)
	}

	return &ResolvedContext{
		Repos:      ctx.Repos,
		Components: ctx.Components,
		Dir:        dir,
	}, nil
}
