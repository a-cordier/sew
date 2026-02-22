package api

import (
	"context"
)

type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type HelmSpec struct {
	Chart   string   `yaml:"chart"`
	Version string   `yaml:"version,omitempty"`
	Values  []string `yaml:"values,omitempty"`
}

type ManifestSpec struct {
	Files []string `yaml:"files"`
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
	Manifest  *ManifestSpec `yaml:"manifest,omitempty"`
}

// EffectiveType returns Type, defaulting to "helm".
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

// ResolvedContext is a fully resolved context with all referenced files in Dir.
type ResolvedContext struct {
	Repos      []Repo
	Components []Component
	Dir        string
}

// Resolver resolves a context path against a registry into a ResolvedContext.
type Resolver interface {
	Resolve(ctx context.Context, contextPath string) (*ResolvedContext, error)
}
