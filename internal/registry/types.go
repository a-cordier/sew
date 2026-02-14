package registry

// Repo represents a Helm chart repository.
type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// HelmSpec holds the Helm-specific configuration for a component.
type HelmSpec struct {
	Chart   string   `yaml:"chart"`
	Version string   `yaml:"version,omitempty"`
	Values  []string `yaml:"values,omitempty"`
}

// ManifestSpec holds the manifest-specific configuration for a component.
// Files can be local paths or URLs.
type ManifestSpec struct {
	Files []string `yaml:"files"`
}

// Component represents a single installable unit.
// The Type field determines which installer handles it
// and which nested spec is populated.
type Component struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type,omitempty"`      // "helm" (default), "manifest", "kustomize", ...
	Namespace string `yaml:"namespace,omitempty"`

	// Installer-specific specs -- only one is populated per component.
	Helm     *HelmSpec     `yaml:"helm,omitempty"`
	Manifest *ManifestSpec `yaml:"manifest,omitempty"`
	// Kustomize *KustomizeSpec `yaml:"kustomize,omitempty"` // future
}

// EffectiveType returns the component type, defaulting to "helm".
func (c *Component) EffectiveType() string {
	if c.Type == "" {
		return "helm"
	}
	return c.Type
}

// Context is the parsed content of a context.yaml file within the registry.
type Context struct {
	Includes   []string    `yaml:"includes,omitempty"`
	Repos      []Repo      `yaml:"repos,omitempty"`
	Components []Component `yaml:"components"`
}

// ResolvedContext is a fully resolved context with all referenced files
// available in a local directory.
type ResolvedContext struct {
	Repos      []Repo
	Components []Component
	// Dir is the local directory containing fetched files.
	// All relative file paths in Components are relative to this dir.
	Dir string
}
