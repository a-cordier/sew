package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FSResolver resolves contexts from a local filesystem directory.
type FSResolver struct {
	Root string // absolute path to the registry root
}

// Resolve reads {Root}/{contextPath}/context.yaml and returns a
// ResolvedContext whose Dir points to the context directory on disk.
// Values files are already local, so no downloads are needed.
//
// If context.yaml does not exist, Resolve looks for a .default file
// containing the name of a default variant sub-directory. When found,
// it appends the variant to contextPath and resolves again.
func (r *FSResolver) Resolve(ctx context.Context, contextPath string) (*ResolvedContext, error) {
	dir := filepath.Join(r.Root, contextPath)
	data, err := os.ReadFile(filepath.Join(dir, "context.yaml"))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("reading context file: %w", err)
		}
		// context.yaml not found – try the .default fallback.
		variant, defaultErr := os.ReadFile(filepath.Join(dir, ".default"))
		if defaultErr != nil {
			// No .default either – report the original missing context.yaml.
			return nil, fmt.Errorf("reading context file: %w", err)
		}
		name := strings.TrimSpace(string(variant))
		if name == "" {
			return nil, fmt.Errorf("empty .default file in %s", dir)
		}
		return r.Resolve(ctx, filepath.Join(contextPath, name))
	}

	var ctxFile Context
	if err := yaml.Unmarshal(data, &ctxFile); err != nil {
		return nil, fmt.Errorf("parsing context.yaml: %w", err)
	}

	return &ResolvedContext{
		Repos:      ctxFile.Repos,
		Components: ctxFile.Components,
		Dir:        dir,
	}, nil
}
