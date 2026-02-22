package core

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

// ContextKindConfig declares Kind cluster requirements from a context (e.g. name, extra port mappings).
type ContextKindConfig struct {
	Name               string         `yaml:"name,omitempty"`
	ExtraPortMappings  []PortMapping  `yaml:"extraPortMappings,omitempty"`
	Nodes              []ContextKindNode `yaml:"nodes,omitempty"`
}

// ContextKindNode is a node entry under context kind (for port mappings under nodes[0]).
type ContextKindNode struct {
	ExtraPortMappings []PortMapping `yaml:"extraPortMappings,omitempty"`
}

// Context is the parsed content of a context.yaml file.
type Context struct {
	Kind       *ContextKindConfig `yaml:"kind,omitempty"`
	Includes   []string           `yaml:"includes,omitempty"`
	Repos      []Repo             `yaml:"repos,omitempty"`
	Components []Component        `yaml:"components"`
}

// ResolvedContext is a fully resolved context with all referenced files in Dir.
type ResolvedContext struct {
	Repos      []Repo
	Components []Component
	Dir        string
	Kind       *ContextKindConfig
}

// Resolver resolves a context path against a registry into a ResolvedContext.
type Resolver interface {
	Resolve(ctx context.Context, contextPath string) (*ResolvedContext, error)
}
