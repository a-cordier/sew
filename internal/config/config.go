package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a-cordier/sew/core"
	"gopkg.in/yaml.v3"
)

// ApplyDefaults merges embedded default values into cfg. Only fields that are
// still at their zero value are populated. The context field is intentionally
// not defaulted so that omitting it yields a plain Kind cluster.
func ApplyDefaults(cfg *core.Config, defaultData []byte) {
	if len(defaultData) == 0 {
		return
	}
	var defaults core.Config
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
func Merge(base, override *core.Config) {
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
	if override.Images.Mirrors != nil {
		base.Images = override.Images
	}
	override.Kind.MergeWithDefaults(&base.Kind)
	base.Kind = override.Kind
}

func Load(path string) (*core.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var cfg core.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	cfg.Kind.ApplyDefaults()
	cfg.Dir = filepath.Dir(path)
	return &cfg, nil
}
