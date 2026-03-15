package registry

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/config"
)

type contextRef struct {
	Registry string
	Context  string
}

type visitedKey struct{}

// withVisited adds ref to the visited set carried in ctx. It returns an error
// if ref was already present (cycle detected).
func withVisited(ctx context.Context, ref contextRef) (context.Context, error) {
	visited, _ := ctx.Value(visitedKey{}).(map[contextRef]bool)
	if visited[ref] {
		return ctx, fmt.Errorf("context cycle detected: %q in registry %q", ref.Context, ref.Registry)
	}
	next := make(map[contextRef]bool, len(visited)+1)
	for k, v := range visited {
		next[k] = v
	}
	next[ref] = true
	return context.WithValue(ctx, visitedKey{}, next), nil
}

// resolveRegistryURL resolves a registry URL that may contain a relative
// file:// path against contextDir.
func resolveRegistryURL(rawURL, contextDir string) string {
	if !strings.HasPrefix(rawURL, "file://") {
		return rawURL
	}
	p := strings.TrimPrefix(rawURL, "file://")
	if !filepath.IsAbs(p) {
		p = filepath.Join(contextDir, p)
	}
	return "file://" + filepath.Clean(p)
}

// resolveFrom resolves all context refs declared in childCfg.From, merges them
// left-to-right into an accumulator, then applies the child's local overrides
// on top. selfRegistryURL is used as the parent registry when childCfg.Registry
// is empty.
func resolveFrom(ctx context.Context, childCfg config.Config, childDir, selfRegistryURL, sewHome string) (*config.ResolvedContext, error) {
	registryURL := selfRegistryURL
	if childCfg.Registry != "" {
		registryURL = resolveRegistryURL(childCfg.Registry, childDir)
	}

	acc := &config.ResolvedContext{}
	for _, ref := range childCfg.From {
		resolver := NewResolver(registryURL, sewHome)
		parent, err := resolver.Resolve(ctx, ref)
		if err != nil {
			return nil, fmt.Errorf("resolving from %q: %w", ref, err)
		}
		absolutizeComponentPaths(parent)
		mergeInto(acc, parent)
	}

	MergeComponents(acc, childCfg.Components, childDir)
	acc.Repos = MergeRepos(acc.Repos, childCfg.Repos)
	acc.Features = config.MergeFeatures(acc.Features, childCfg.Features)
	acc.Kind = mergeKind(acc.Kind, childCfg.Kind)
	acc.Images = config.MergeImages(acc.Images, childCfg.Images)
	acc.Notes = mergeNotes(acc.Notes, readNotes(childDir))

	return acc, nil
}

// mergeInto merges a resolved context (src) into an accumulator (acc).
// Later sources override earlier ones for matching fields.
func mergeInto(acc, src *config.ResolvedContext) {
	MergeComponents(acc, src.Components, "")
	acc.Repos = MergeRepos(acc.Repos, src.Repos)
	acc.Features = config.MergeFeatures(acc.Features, src.Features)
	acc.Kind = mergeKind(acc.Kind, src.Kind)
	acc.Images = config.MergeImages(acc.Images, src.Images)
	acc.Notes = mergeNotes(acc.Notes, src.Notes)
}

// mergeNotes merges child notes on top of base notes.
// Non-empty child fields win; empty fields inherit from base.
func mergeNotes(base, child config.ResolvedNotes) config.ResolvedNotes {
	result := base
	if child.Create != "" {
		result.Create = child.Create
	}
	if child.Delete != "" {
		result.Delete = child.Delete
	}
	return result
}

// mergeKind merges child Kind overrides on top of a base KindConfig.
// Non-zero child fields win; zero-value fields inherit from base.
func mergeKind(base, child config.KindConfig) config.KindConfig {
	result := base
	if child.Name != "" {
		result.Name = child.Name
	}
	if child.APIVersion != "" {
		result.APIVersion = child.APIVersion
	}
	if child.Kind != "" {
		result.Kind = child.Kind
	}
	if len(child.ContainerdConfigPatches) > 0 {
		result.ContainerdConfigPatches = child.ContainerdConfigPatches
	}
	if len(child.Nodes) > 0 {
		result.Nodes = child.Nodes
		if len(base.Nodes) > 0 && len(result.Nodes[0].ExtraPortMappings) == 0 {
			result.Nodes[0].ExtraPortMappings = base.Nodes[0].ExtraPortMappings
		}
	}
	return result
}
