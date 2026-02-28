package registry

import (
	"path/filepath"

	"github.com/a-cordier/sew/core"
)

// MergeComponents merges user-level component customizations into the resolved
// context. Components are matched by name. For each match the merge applies:
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
			continue
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
		for _, v := range patch.Helm.ValueFiles {
			path := v
			if !filepath.IsAbs(path) && configDir != "" {
				path = filepath.Join(configDir, v)
			}
			comp.Helm.ValueFiles = append(comp.Helm.ValueFiles, path)
		}
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
