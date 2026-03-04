package core

type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type HelmSpec struct {
	Chart      string                 `yaml:"chart"`
	Version    string                 `yaml:"version,omitempty"`
	ValueFiles []string               `yaml:"valueFiles,omitempty"`
	Values     map[string]interface{} `yaml:"values,omitempty"`
}

type K8sSpec struct {
	ManifestFiles []string                 `yaml:"manifestFiles,omitempty"`
	Manifests     []map[string]interface{} `yaml:"manifests,omitempty"`
}

type Conditions struct {
	Ready bool `yaml:"ready,omitempty"`
}

type Selector struct {
	MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
}

// Requirement declares a dependency on another component.
type Requirement struct {
	Component  string     `yaml:"component"`
	Conditions Conditions `yaml:"conditions,omitempty"`
	Selector   *Selector  `yaml:"selector,omitempty"`
	Timeout    string     `yaml:"timeout,omitempty"`
}

type Component struct {
	Name      string        `yaml:"name"`
	Type      string        `yaml:"type,omitempty"`
	Namespace string        `yaml:"namespace,omitempty"`
	Requires  []Requirement `yaml:"requires,omitempty"`
	Helm      *HelmSpec     `yaml:"helm,omitempty"`
	K8s       *K8sSpec      `yaml:"k8s,omitempty"`
}

// EffectiveType returns Type, defaulting to "helm".
func (c *Component) EffectiveType() string {
	if c.Type == "" {
		return "helm"
	}
	return c.Type
}

// ResolvedContext is a fully resolved context with all referenced files in Dir.
type ResolvedContext struct {
	Repos      []Repo
	Components []Component
	Dir        string
	Kind       KindConfig
	Features   FeaturesConfig
	Images     ImagesConfig
}

