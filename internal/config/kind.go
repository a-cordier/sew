package config

import (
	"gopkg.in/yaml.v3"
)

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
	Role                 string            `yaml:"role"`
	ExtraPortMappings    []PortMapping     `yaml:"extraPortMappings,omitempty"`
	ExtraMounts          []Mount           `yaml:"extraMounts,omitempty"`
	Labels               map[string]string `yaml:"labels,omitempty"`
	KubeadmConfigPatches []string          `yaml:"kubeadmConfigPatches,omitempty"`
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

// MergeWithContext merges context Kind requirements into the config. When
// sew.yaml does not set a custom cluster name (still the default "sew"), the
// context's kind.name is used.
//
// Port mappings use full-replacement semantics: if the user defines any
// extraPortMappings on the node, those are kept as-is and all context ports
// are ignored. If the user defines none, the context ports are used.
func (k *KindConfig) MergeWithContext(ctx *KindConfig) {
	if ctx == nil {
		return
	}
	if ctx.Name != "" && k.Name == KindDefaultName {
		k.Name = ctx.Name
	}
	if len(k.Nodes) == 0 || len(ctx.Nodes) == 0 {
		return
	}
	node := &k.Nodes[0]
	if len(node.ExtraPortMappings) == 0 {
		node.ExtraPortMappings = ctx.Nodes[0].ExtraPortMappings
	}
}

// MergeWithDefaults fills zero-value Kind fields from the embedded defaults,
// preserving any user-specified values.
//
// Port mappings use full-replacement semantics: if the user defines any
// extraPortMappings on the node, those are kept as-is and the default port
// mappings are ignored. If the user defines none, the defaults are used.
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
		if len(node.ExtraPortMappings) == 0 {
			node.ExtraPortMappings = defaults.Nodes[0].ExtraPortMappings
		}
	}
}
