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

// MergeImages merges context defaults (base) with user overrides. Preload
// refs are deduplicated (union); non-nil Preload/Mirrors from the override
// wins.
func MergeImages(base, override ImagesConfig) ImagesConfig {
	result := base
	if override.Mirrors != nil {
		result.Mirrors = override.Mirrors
	}
	result.Preload = mergePreload(base.Preload, override.Preload)
	return result
}

func mergePreload(base, override *PreloadConfig) *PreloadConfig {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}
	return &PreloadConfig{
		Refs: mergeRefs(base.Refs, override.Refs),
	}
}

func mergeRefs(base, override []string) []string {
	if len(base) == 0 {
		return override
	}
	if len(override) == 0 {
		return base
	}
	seen := make(map[string]bool, len(base))
	merged := make([]string, 0, len(base)+len(override))
	for _, img := range base {
		if !seen[img] {
			seen[img] = true
			merged = append(merged, img)
		}
	}
	for _, img := range override {
		if !seen[img] {
			seen[img] = true
			merged = append(merged, img)
		}
	}
	return merged
}
