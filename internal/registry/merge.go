package registry

import (
	"path/filepath"

	"github.com/a-cordier/sew/internal/config"
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
func MergeComponents(resolved *config.ResolvedContext, components []config.Component, configDir string) {
	if len(components) == 0 {
		return
	}
	byName := make(map[string]*config.Component, len(resolved.Components))
	for i := range resolved.Components {
		byName[resolved.Components[i].Name] = &resolved.Components[i]
	}
	for _, patch := range components {
		comp, ok := byName[patch.Name]
		if !ok {
			resolveValueFilePaths(&patch, configDir)
			resolveManifestFilePaths(&patch, configDir)
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
		if patch.Helm != nil && comp.Helm != nil {
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
		if patch.K8s != nil {
			if comp.K8s == nil {
				comp.K8s = &config.K8sSpec{}
			}
			resolveManifestFilePaths(&patch, configDir)
			comp.K8s.ManifestFiles = append(comp.K8s.ManifestFiles, patch.K8s.ManifestFiles...)
			comp.K8s.Manifests = append(comp.K8s.Manifests, patch.K8s.Manifests...)
		}
	}
}

// resolveValueFilePaths resolves relative value file paths in a component's
// HelmSpec to absolute paths based on configDir.
func resolveValueFilePaths(c *config.Component, configDir string) {
	if c.Helm == nil || configDir == "" {
		return
	}
	for i, v := range c.Helm.ValueFiles {
		if !filepath.IsAbs(v) {
			c.Helm.ValueFiles[i] = filepath.Join(configDir, v)
		}
	}
}

// resolveManifestFilePaths resolves relative manifest file paths to absolute
// paths based on configDir. This allows user configs to reference manifest
// files relative to their own location rather than the registry context dir.
func resolveManifestFilePaths(c *config.Component, configDir string) {
	if c.K8s == nil || configDir == "" {
		return
	}
	for i, f := range c.K8s.ManifestFiles {
		if !filepath.IsAbs(f) {
			c.K8s.ManifestFiles[i] = filepath.Join(configDir, f)
		}
	}
}

// MergeRepos merges local repos into context repos, deduplicating by name.
// When both lists contain a repo with the same name, the local entry wins.
func MergeRepos(contextRepos, localRepos []config.Repo) []config.Repo {
	if len(localRepos) == 0 {
		return contextRepos
	}
	localByName := make(map[string]config.Repo, len(localRepos))
	for _, r := range localRepos {
		localByName[r.Name] = r
	}
	out := make([]config.Repo, 0, len(contextRepos)+len(localRepos))
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
