package core

import (
	"fmt"
	"sort"

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

type GatewayChannel string

const (
	GatewayChannelStandard     GatewayChannel = "standard"
	GatewayChannelExperimental GatewayChannel = "experimental"

	DNSDefaultDomain = "sew.local"
	DNSDefaultPort   = 15353
)

type LBConfig struct {
	Enabled bool `yaml:"enabled"`
}

type GatewayConfig struct {
	Enabled bool           `yaml:"enabled"`
	Channel GatewayChannel `yaml:"channel,omitempty"`
}

type DNSRecord struct {
	Hostname  string `yaml:"hostname"`
	Service   string `yaml:"service"`
	Namespace string `yaml:"namespace"`
}

type DNSConfig struct {
	Enabled bool        `yaml:"enabled"`
	Domain  string      `yaml:"domain,omitempty"`
	Port    int         `yaml:"port,omitempty"`
	Records []DNSRecord `yaml:"records,omitempty"`
}

// FeaturesConfig groups optional networking features. Pointer semantics: a
// non-nil sub-config means the user explicitly set the feature; nil means
// "use the context default."
type FeaturesConfig struct {
	LB *LBConfig `yaml:"lb,omitempty"`
	Gateway      *GatewayConfig     `yaml:"gateway,omitempty"`
	DNS          *DNSConfig         `yaml:"dns,omitempty"`
}

// MergeFeatures merges context defaults (base) with user overrides. For each
// feature key, a non-nil override replaces the base; otherwise the base stands.
func MergeFeatures(base, override FeaturesConfig) FeaturesConfig {
	result := base
	if override.LB != nil {
		result.LB = override.LB
	}
	if override.Gateway != nil {
		result.Gateway = override.Gateway
	}
	if override.DNS != nil {
		result.DNS = override.DNS
	}
	return result
}

// ResolveFeatureDependencies validates inter-feature constraints and
// auto-enables implied features. It mutates f in place (e.g. enabling
// lb when gateway requires it, filling defaults for DNS).
// It returns warnings for non-fatal issues and an error for conflicts.
func ResolveFeatureDependencies(f *FeaturesConfig) (warnings []string, err error) {
	gwEnabled := f.Gateway != nil && f.Gateway.Enabled
	if gwEnabled {
		if f.LB == nil {
			f.LB = &LBConfig{Enabled: true}
		} else if !f.LB.Enabled {
			return nil, fmt.Errorf("gateway requires lb, but lb is explicitly disabled")
		}
	}

	dnsEnabled := f.DNS != nil && f.DNS.Enabled
	if dnsEnabled {
		if f.DNS.Domain == "" {
			f.DNS.Domain = DNSDefaultDomain
		}
		if f.DNS.Port == 0 {
			f.DNS.Port = DNSDefaultPort
		}
	}

	return warnings, nil
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
// are ignored. If the user defines none, the context ports are used. Within
// the context itself, top-level and per-node port mappings are merged by
// containerPort (per-node wins on conflict) before being applied.
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
	if len(node.ExtraPortMappings) == 0 {
		node.ExtraPortMappings = contextPorts
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

// mergePortMappings combines two port-mapping slices keyed by containerPort,
// with override taking precedence on conflict. It is used internally to
// reconcile top-level and per-node port mappings within a single context file.
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

