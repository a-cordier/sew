package config

type ImagesConfig struct {
	Mirrors *MirrorsConfig `yaml:"mirrors,omitempty"`
	Preload *PreloadConfig `yaml:"preload,omitempty"`
}

type PreloadConfig struct {
	Refs []string `yaml:"refs,omitempty"`
}

type MirrorsConfig struct {
	Data      string   `yaml:"data,omitempty"`
	Upstreams []string `yaml:"upstreams,omitempty"`
}

// MergeImages merges context defaults (base) with user overrides. For both
// Preload and Mirrors, a non-nil override replaces the base entirely.
func MergeImages(base, override ImagesConfig) ImagesConfig {
	result := base
	if override.Mirrors != nil {
		result.Mirrors = override.Mirrors
	}
	result.Preload = mergePreload(base.Preload, override.Preload)
	return result
}

func mergePreload(base, override *PreloadConfig) *PreloadConfig {
	if override != nil {
		return override
	}
	return base
}
