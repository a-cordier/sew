package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/config"
	"gopkg.in/yaml.v3"
)

// FSResolver resolves contexts from a local filesystem directory.
type FSResolver struct {
	Root    string // absolute path to the registry root
	SewHome string
}

// Resolve reads {Root}/{contextPath}/sew.yaml and returns a
// ResolvedContext whose Dir points to the context directory on disk.
// Values files are already local, so no downloads are needed.
//
// If the parsed config declares a parent context (via the context field),
// the parent is resolved first and the child's overrides are merged on top.
//
// For backward compatibility, if sew.yaml does not exist, Resolve
// falls back to context.yaml. If neither exists, it looks for a
// .default file containing the name of a default variant
// sub-directory. When found, it appends the variant to contextPath
// and resolves again.
func (r *FSResolver) Resolve(ctx context.Context, contextPath string) (*config.ResolvedContext, error) {
	selfRegistry := "file://" + r.Root
	ctx, err := withVisited(ctx, contextRef{Registry: selfRegistry, Context: contextPath})
	if err != nil {
		return nil, err
	}

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

	var ctxCfg config.Config
	if err := yaml.Unmarshal(data, &ctxCfg); err != nil {
		return nil, fmt.Errorf("parsing context file: %w", err)
	}

	if len(ctxCfg.From) > 0 {
		return resolveFrom(ctx, ctxCfg, dir, selfRegistry, r.SewHome)
	}

	return &config.ResolvedContext{
		Repos:      ctxCfg.Repos,
		Components: ctxCfg.Components,
		Dir:        dir,
		Kind:       ctxCfg.Kind,
		Features:   ctxCfg.Features,
		Images:     ctxCfg.Images,
		Notes:      readNotes(dir),
		Abstract:   ctxCfg.Abstract,
	}, nil
}

func readNotes(dir string) config.ResolvedNotes {
	var notes config.ResolvedNotes
	if data, err := os.ReadFile(filepath.Join(dir, "notes.create")); err == nil {
		notes.Create = string(data)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "notes.delete")); err == nil {
		notes.Delete = string(data)
	}
	return notes
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
