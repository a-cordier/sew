package core

import (
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Registry   string      `yaml:"registry"`
	Context    string      `yaml:"context"`
	Kind       KindConfig  `yaml:"kind"`
	Images     ImagesConfig `yaml:"images,omitempty"`
	Components []Component `yaml:"components,omitempty"`

	// Dir is set by Load to resolve relative paths in component value files.
	Dir string `yaml:"-"`
}

type ImagesConfig struct {
	Mirrors *MirrorsConfig `yaml:"mirrors,omitempty"`
}

type MirrorsConfig struct {
	Data      string   `yaml:"data,omitempty"`
	Upstreams []string `yaml:"upstreams,omitempty"`
}

const (
	KindDefaultAPIVersion = "kind.x-k8s.io/v1alpha4"
	KindDefaultKind       = "Cluster"
	KindDefaultName       = "sew"
)

type KindConfig struct {
	APIVersion              string     `yaml:"apiVersion"`
	Kind                    string     `yaml:"kind"`
	Name                    string     `yaml:"name"`
	Nodes                   []KindNode `yaml:"nodes"`
	ContainerdConfigPatches []string   `yaml:"containerdConfigPatches,omitempty"`
}

func (k *KindConfig) ApplyDefaults() {
	if k.APIVersion == "" {
		k.APIVersion = KindDefaultAPIVersion
	}
	if k.Kind == "" {
		k.Kind = KindDefaultKind
	}
	if k.Name == "" {
		k.Name = KindDefaultName
	}
	if len(k.Nodes) == 0 {
		k.Nodes = []KindNode{{Role: "control-plane"}}
	}
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

// RawYAML returns the KindConfig serialized as YAML.
func (k *KindConfig) RawYAML() ([]byte, error) {
	return yaml.Marshal(k)
}

// MergeWithContext merges context Kind requirements into the config. Context
// provides the base (e.g. name, extra port mappings); user config overrides. On
// conflict (same containerPort), user wins. When sew.yaml does not set a custom
// cluster name (still the default "sew"), the context's kind.name is used.
func (k *KindConfig) MergeWithContext(ctx *ContextKindConfig) {
	if ctx == nil {
		return
	}
	if ctx.Name != "" && k.Name == KindDefaultName {
		k.Name = ctx.Name
	}
	if len(k.Nodes) == 0 {
		return
	}
	contextPorts := ctx.ExtraPortMappings
	if len(ctx.Nodes) > 0 && len(ctx.Nodes[0].ExtraPortMappings) > 0 {
		contextPorts = mergePortMappings(contextPorts, ctx.Nodes[0].ExtraPortMappings)
	}
	node := &k.Nodes[0]
	node.ExtraPortMappings = mergePortMappings(contextPorts, node.ExtraPortMappings)
}

// MergeWithDefaults fills zero-value Kind fields from the embedded defaults,
// preserving any user-specified values. Default port mappings are used as a
// base; user mappings override on the same containerPort.
func (k *KindConfig) MergeWithDefaults(defaults *KindConfig) {
	if defaults == nil {
		return
	}
	if k.Name == KindDefaultName && defaults.Name != "" {
		k.Name = defaults.Name
	}
	if len(k.Nodes) == 0 && len(defaults.Nodes) > 0 {
		k.Nodes = make([]KindNode, len(defaults.Nodes))
		copy(k.Nodes, defaults.Nodes)
		return
	}
	if len(k.Nodes) > 0 && len(defaults.Nodes) > 0 {
		node := &k.Nodes[0]
		node.ExtraPortMappings = mergePortMappings(defaults.Nodes[0].ExtraPortMappings, node.ExtraPortMappings)
	}
}

func mergePortMappings(base, override []PortMapping) []PortMapping {
	byPort := make(map[int32]PortMapping)
	for _, p := range base {
		byPort[p.ContainerPort] = p
	}
	for _, p := range override {
		byPort[p.ContainerPort] = p
	}
	if len(byPort) == 0 {
		return nil
	}
	ports := make([]int32, 0, len(byPort))
	for cp := range byPort {
		ports = append(ports, cp)
	}
	sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })
	out := make([]PortMapping, 0, len(ports))
	for _, cp := range ports {
		out = append(out, byPort[cp])
	}
	return out
}

