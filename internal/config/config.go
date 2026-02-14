package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Home      string                `yaml:"home"`
	Registry  string                `yaml:"registry"`
	Context   string                `yaml:"context"`
	Kind      KindConfig            `yaml:"kind"`
	Overrides map[string]ComponentOverride `yaml:"overrides,omitempty"`

	// Dir is the directory containing the config file (set by Load).
	// Used to resolve relative paths in overrides.
	Dir string `yaml:"-"`
}

// ComponentOverride holds user overrides for a component, keyed by component name.
type ComponentOverride struct {
	Helm *HelmOverride `yaml:"helm,omitempty"`
}

// HelmOverride allows overriding version and appending values files for a Helm component.
type HelmOverride struct {
	Version string   `yaml:"version,omitempty"`
	Values  []string `yaml:"values,omitempty"`
}

type KindConfig struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Name       string     `yaml:"name"`
	Nodes      []KindNode `yaml:"nodes"`
}

type KindNode struct {
	Role              string            `yaml:"role"`
	ExtraPortMappings []PortMapping     `yaml:"extraPortMappings,omitempty"`
	ExtraMounts       []Mount           `yaml:"extraMounts,omitempty"`
	Labels            map[string]string `yaml:"labels,omitempty"`
}

type PortMapping struct {
	ContainerPort int32  `yaml:"containerPort"`
	HostPort      int32  `yaml:"hostPort"`
	Protocol      string `yaml:"protocol,omitempty"`
}

type Mount struct {
	HostPath      string `yaml:"hostPath"`
	ContainerPath string `yaml:"containerPath"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	cfg.Dir = filepath.Dir(path)
	return &cfg, nil
}

func (k *KindConfig) RawYAML() []byte {
	data, _ := yaml.Marshal(k)
	return data
}
