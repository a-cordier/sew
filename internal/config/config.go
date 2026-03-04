package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Registry   string         `yaml:"registry"`
	Context    string         `yaml:"context"`
	Kind       KindConfig     `yaml:"kind"`
	Features   FeaturesConfig `yaml:"features,omitempty"`
	Images     ImagesConfig   `yaml:"images,omitempty"`
	Repos      []Repo         `yaml:"repos,omitempty"`
	Components []Component    `yaml:"components,omitempty"`

	// Dir is set by Load to resolve relative paths in component value files.
	Dir string `yaml:"-"`
}

// ApplyDefaults merges embedded default values into cfg. Only fields that are
// still at their zero value are populated. The context field is intentionally
// not defaulted so that omitting it yields a plain Kind cluster.
func ApplyDefaults(cfg *Config, defaultData []byte) {
	if len(defaultData) == 0 {
		return
	}
	var defaults Config
	if err := yaml.Unmarshal(defaultData, &defaults); err != nil {
		return
	}
	if cfg.Registry == "" {
		cfg.Registry = defaults.Registry
	}
	cfg.Kind.MergeWithDefaults(&defaults.Kind)
}

// Merge applies non-zero fields from override onto base. For each top-level
// Config field, the override value wins when set.
func Merge(base, override *Config) {
	if override.Registry != "" {
		base.Registry = override.Registry
	}
	if override.Context != "" {
		base.Context = override.Context
	}
	if override.Dir != "" {
		base.Dir = override.Dir
	}
	if len(override.Repos) > 0 {
		base.Repos = override.Repos
	}
	if len(override.Components) > 0 {
		base.Components = override.Components
	}
	base.Images = MergeImages(base.Images, override.Images)
	base.Features = MergeFeatures(base.Features, override.Features)
	override.Kind.MergeWithDefaults(&base.Kind)
	base.Kind = override.Kind
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

	cfg.Kind.ApplyDefaults()
	cfg.Dir = filepath.Dir(path)
	return &cfg, nil
}
