package config

// Build describes a local Docker image build for the inner dev loop.
type Build struct {
	Name       string   `yaml:"name"`
	Image      string   `yaml:"image"`
	Dir        string   `yaml:"dir,omitempty"`
	Pre        []string `yaml:"pre,omitempty"`
	Context    string   `yaml:"context,omitempty"`
	Dockerfile string   `yaml:"dockerfile,omitempty"`
}

// BuildImageRefs returns the image references from a list of builds,
// suitable for adding to the preload skip list.
func BuildImageRefs(builds []Build) []string {
	if len(builds) == 0 {
		return nil
	}
	refs := make([]string, 0, len(builds))
	for _, b := range builds {
		if b.Image != "" {
			refs = append(refs, b.Image)
		}
	}
	return refs
}
