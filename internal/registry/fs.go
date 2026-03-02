package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/core"
	"gopkg.in/yaml.v3"
)

// FSResolver resolves contexts from a local filesystem directory.
type FSResolver struct {
	Root string // absolute path to the registry root
}

// Resolve reads {Root}/{contextPath}/sew.yaml and returns a
// ResolvedContext whose Dir points to the context directory on disk.
// Values files are already local, so no downloads are needed.
//
// For backward compatibility, if sew.yaml does not exist, Resolve
// falls back to context.yaml. If neither exists, it looks for a
// .default file containing the name of a default variant
// sub-directory. When found, it appends the variant to contextPath
// and resolves again.
func (r *FSResolver) Resolve(ctx context.Context, contextPath string) (*core.ResolvedContext, error) {
	dir := filepath.Join(r.Root, contextPath)

	data, err := r.readContextFile(dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("reading context file: %w", err)
		}
		variant, defaultErr := os.ReadFile(filepath.Join(dir, ".default"))
		if defaultErr != nil {
			return nil, fmt.Errorf("reading context file: %w", err)
		}
		name := strings.TrimSpace(string(variant))
		if name == "" {
			return nil, fmt.Errorf("empty .default file in %s", dir)
		}
		return r.Resolve(ctx, filepath.Join(contextPath, name))
	}

	var ctxFile core.Context
	if err := yaml.Unmarshal(data, &ctxFile); err != nil {
		return nil, fmt.Errorf("parsing context file: %w", err)
	}

	return &core.ResolvedContext{
		Repos:      ctxFile.Repos,
		Components: ctxFile.Components,
		Dir:        dir,
		Kind:       ctxFile.Kind,
		Features:   ctxFile.Features,
	}, nil
}

// readContextFile tries sew.yaml first, falling back to context.yaml
// for backward compatibility.
func (r *FSResolver) readContextFile(dir string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(dir, "sew.yaml"))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return os.ReadFile(filepath.Join(dir, "context.yaml"))
	}
	return data, err
}
