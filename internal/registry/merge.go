package registry

import (
	"path/filepath"

	"github.com/a-cordier/sew/internal/config"
)

// ApplyOverrides applies user overrides from config onto the resolved context.
// Override values file paths are resolved relative to configDir.
// Mutates resolved.Components in place.
func ApplyOverrides(resolved *ResolvedContext, overrides map[string]config.ComponentOverride, configDir string) {
	if overrides == nil {
		return
	}
	for i := range resolved.Components {
		comp := &resolved.Components[i]
		o, ok := overrides[comp.Name]
		if !ok {
			continue
		}
		if o.Helm != nil && comp.Helm != nil {
			if o.Helm.Version != "" {
				comp.Helm.Version = o.Helm.Version
			}
			for _, v := range o.Helm.Values {
				path := v
				if !filepath.IsAbs(path) && configDir != "" {
					path = filepath.Join(configDir, v)
				}
				comp.Helm.Values = append(comp.Helm.Values, path)
			}
		}
	}
}
