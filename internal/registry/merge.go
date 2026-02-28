package registry

import (
	"path/filepath"

	"github.com/a-cordier/sew/core"
)

// MergeComponents merges user-level component customizations into the resolved
// context. Components are matched by name. For each match the merge applies:
//   - requires: user requirements are appended (deduplicated by component name)
//   - helm.chart: user wins if non-empty
//   - helm.version: user wins if non-empty
//   - helm.valueFiles: user files are appended (higher precedence in Helm)
//   - helm.values: shallow-merged on top of context values (user wins per key)
//
// Value file paths from the user config are resolved relative to configDir.
func MergeComponents(resolved *core.ResolvedContext, components []core.Component, configDir string) {
	if len(components) == 0 {
		return
	}
	byName := make(map[string]*core.Component, len(resolved.Components))
	for i := range resolved.Components {
		byName[resolved.Components[i].Name] = &resolved.Components[i]
	}
	for _, patch := range components {
		comp, ok := byName[patch.Name]
		if !ok {
			resolveValueFilePaths(&patch, configDir)
			resolved.Components = append(resolved.Components, patch)
			continue
		}
		if len(patch.Requires) > 0 {
			existing := make(map[string]bool, len(comp.Requires))
			for _, r := range comp.Requires {
				existing[r.Component] = true
			}
			for _, r := range patch.Requires {
				if !existing[r.Component] {
					comp.Requires = append(comp.Requires, r)
				}
			}
		}
		if patch.Helm == nil || comp.Helm == nil {
			continue
		}
		if patch.Helm.Chart != "" {
			comp.Helm.Chart = patch.Helm.Chart
		}
		if patch.Helm.Version != "" {
			comp.Helm.Version = patch.Helm.Version
		}
		resolveValueFilePaths(&patch, configDir)
		comp.Helm.ValueFiles = append(comp.Helm.ValueFiles, patch.Helm.ValueFiles...)
		if len(patch.Helm.Values) > 0 {
			if comp.Helm.Values == nil {
				comp.Helm.Values = make(map[string]interface{})
			}
			for k, v := range patch.Helm.Values {
				comp.Helm.Values[k] = v
			}
		}
	}
}

// resolveValueFilePaths resolves relative value file paths in a component's
// HelmSpec to absolute paths based on configDir.
func resolveValueFilePaths(c *core.Component, configDir string) {
	if c.Helm == nil || configDir == "" {
		return
	}
	for i, v := range c.Helm.ValueFiles {
		if !filepath.IsAbs(v) {
			c.Helm.ValueFiles[i] = filepath.Join(configDir, v)
		}
	}
}

// MergeRepos merges local repos into context repos, deduplicating by name.
// When both lists contain a repo with the same name, the local entry wins.
func MergeRepos(contextRepos, localRepos []core.Repo) []core.Repo {
	if len(localRepos) == 0 {
		return contextRepos
	}
	localByName := make(map[string]core.Repo, len(localRepos))
	for _, r := range localRepos {
		localByName[r.Name] = r
	}
	out := make([]core.Repo, 0, len(contextRepos)+len(localRepos))
	for _, r := range contextRepos {
		if local, ok := localByName[r.Name]; ok {
			out = append(out, local)
			delete(localByName, r.Name)
		} else {
			out = append(out, r)
		}
	}
	for _, r := range localRepos {
		if _, ok := localByName[r.Name]; ok {
			out = append(out, r)
		}
	}
	return out
}
