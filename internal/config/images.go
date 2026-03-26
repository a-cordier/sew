package config

const PreloadModeReplace = "replace"

type ImagesConfig struct {
	Mirrors *MirrorsConfig `yaml:"mirrors,omitempty"`
	Preload *PreloadConfig `yaml:"preload,omitempty"`
}

type PreloadConfig struct {
	Mode string   `yaml:"mode,omitempty"`
	Refs []string `yaml:"refs,omitempty"`
	Skip []string `yaml:"skip,omitempty"`
}

type MirrorsConfig struct {
	Data      string   `yaml:"data,omitempty"`
	Upstreams []string `yaml:"upstreams,omitempty"`
}

// EffectiveRefs returns Refs minus any entries listed in Skip.
func (p *PreloadConfig) EffectiveRefs() []string {
	if p == nil || len(p.Refs) == 0 {
		return nil
	}
	if len(p.Skip) == 0 {
		return p.Refs
	}
	skipSet := make(map[string]bool, len(p.Skip))
	for _, s := range p.Skip {
		skipSet[s] = true
	}
	var result []string
	for _, r := range p.Refs {
		if !skipSet[r] {
			result = append(result, r)
		}
	}
	return result
}

// MergeImages merges context defaults (base) with user overrides. Mirrors use
// last-writer-wins. Preload merging is governed by the override's Mode field.
func MergeImages(base, override ImagesConfig) ImagesConfig {
	result := base
	if override.Mirrors != nil {
		result.Mirrors = override.Mirrors
	}
	result.Preload = mergePreload(base.Preload, override.Preload)
	return result
}

func mergePreload(base, override *PreloadConfig) *PreloadConfig {
	if override == nil {
		return base
	}
	if base == nil {
		return &PreloadConfig{Refs: override.Refs, Skip: override.Skip}
	}
	if override.Mode == PreloadModeReplace {
		return &PreloadConfig{Refs: override.Refs}
	}
	return &PreloadConfig{
		Refs: unionStrings(base.Refs, override.Refs),
		Skip: unionStrings(base.Skip, override.Skip),
	}
}

func unionStrings(a, b []string) []string {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	seen := make(map[string]bool, len(a)+len(b))
	var result []string
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
