package registry

// Repo is a Helm chart repository.
type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// HelmSpec is the Helm configuration for a component.
type HelmSpec struct {
	Chart   string   `yaml:"chart"`
	Version string   `yaml:"version,omitempty"`
	Values  []string `yaml:"values,omitempty"`
}

// ManifestSpec is the manifest configuration for a component.
type ManifestSpec struct {
	Files []string `yaml:"files"`
}

// Conditions describes the state a required component must be in.
type Conditions struct {
	Ready bool `yaml:"ready,omitempty"`
}

// Selector is a Kubernetes label selector for pod readiness.
type Selector struct {
	MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
}

// Requirement references another component that must be satisfied before installation.
type Requirement struct {
	Component  string     `yaml:"component"`
	Conditions Conditions `yaml:"conditions,omitempty"`
	Selector   *Selector  `yaml:"selector,omitempty"`
	Timeout    string     `yaml:"timeout,omitempty"`
}

// Component is a single installable unit (Helm chart, manifest, etc.).
type Component struct {
	Name      string        `yaml:"name"`
	Type      string        `yaml:"type,omitempty"`
	Namespace string        `yaml:"namespace,omitempty"`
	Requires  []Requirement `yaml:"requires,omitempty"`
	Helm      *HelmSpec     `yaml:"helm,omitempty"`
	Manifest  *ManifestSpec `yaml:"manifest,omitempty"`
}

// EffectiveType returns the component type, defaulting to "helm".
func (c *Component) EffectiveType() string {
	if c.Type == "" {
		return "helm"
	}
	return c.Type
}

// Context is the parsed content of a context.yaml file.
type Context struct {
	Includes   []string    `yaml:"includes,omitempty"`
	Repos      []Repo      `yaml:"repos,omitempty"`
	Components []Component `yaml:"components"`
}

// ResolvedContext is a context with all referenced files available in Dir.
type ResolvedContext struct {
	Repos      []Repo
	Components []Component
	Dir        string
}
