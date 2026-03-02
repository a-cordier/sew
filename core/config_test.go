package core

import (
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestMergeFeatures_OverrideReplacesBase(t *testing.T) {
	base := FeaturesConfig{
		LB: &LBConfig{Enabled: true},
		Gateway:      &GatewayConfig{Enabled: true, Channel: GatewayChannelStandard},
		DNS:          &DNSConfig{Enabled: true, Domain: "ctx.local", Port: 15353},
	}
	override := FeaturesConfig{
		LB: &LBConfig{Enabled: false},
	}
	result := MergeFeatures(base, override)

	if result.LB == nil || result.LB.Enabled {
		t.Fatal("expected LB to be overridden to disabled")
	}
	if result.Gateway == nil || !result.Gateway.Enabled {
		t.Fatal("expected Gateway to be preserved from base")
	}
	if result.DNS == nil || result.DNS.Domain != "ctx.local" {
		t.Fatal("expected DNS to be preserved from base")
	}
}

func TestMergeFeatures_NilOverrideKeepsBase(t *testing.T) {
	base := FeaturesConfig{
		Gateway: &GatewayConfig{Enabled: true, Channel: GatewayChannelExperimental},
	}
	override := FeaturesConfig{}
	result := MergeFeatures(base, override)

	if result.Gateway == nil || result.Gateway.Channel != GatewayChannelExperimental {
		t.Fatal("expected Gateway to be preserved when override is nil")
	}
	if result.LB != nil {
		t.Fatal("expected LB to remain nil")
	}
}

func TestMergeFeatures_BothNil(t *testing.T) {
	result := MergeFeatures(FeaturesConfig{}, FeaturesConfig{})
	if result.LB != nil || result.Gateway != nil || result.DNS != nil {
		t.Fatal("expected all features to remain nil")
	}
}

func TestResolveFeatureDependencies_GatewayAutoEnablesLB(t *testing.T) {
	f := FeaturesConfig{
		Gateway: &GatewayConfig{Enabled: true},
	}
	warnings, err := ResolveFeatureDependencies(&f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if f.LB == nil || !f.LB.Enabled {
		t.Fatal("expected LB to be auto-enabled")
	}
}

func TestResolveFeatureDependencies_GatewayDoesNotDefaultChannel(t *testing.T) {
	f := FeaturesConfig{
		Gateway: &GatewayConfig{Enabled: true},
	}
	if _, err := ResolveFeatureDependencies(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Gateway.Channel != "" {
		t.Fatalf("expected channel to remain empty, got %q", f.Gateway.Channel)
	}
}

func TestResolveFeatureDependencies_GatewayPreservesExplicitChannel(t *testing.T) {
	f := FeaturesConfig{
		Gateway: &GatewayConfig{Enabled: true, Channel: GatewayChannelExperimental},
	}
	if _, err := ResolveFeatureDependencies(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Gateway.Channel != GatewayChannelExperimental {
		t.Fatalf("expected channel %q, got %q", GatewayChannelExperimental, f.Gateway.Channel)
	}
}

func TestResolveFeatureDependencies_GatewayWithLBDisabledErrors(t *testing.T) {
	f := FeaturesConfig{
		LB:      &LBConfig{Enabled: false},
		Gateway: &GatewayConfig{Enabled: true},
	}
	_, err := ResolveFeatureDependencies(&f)
	if err == nil {
		t.Fatal("expected error when gateway is enabled but lb is explicitly disabled")
	}
}

func TestResolveFeatureDependencies_DNSWithoutGatewayNoWarning(t *testing.T) {
	f := FeaturesConfig{
		DNS:     &DNSConfig{Enabled: true},
	}
	warnings, err := ResolveFeatureDependencies(&f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestResolveFeatureDependencies_DNSAppliesDefaults(t *testing.T) {
	f := FeaturesConfig{
		DNS:     &DNSConfig{Enabled: true},
	}
	if _, err := ResolveFeatureDependencies(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.DNS.Domain != DNSDefaultDomain {
		t.Fatalf("expected domain %q, got %q", DNSDefaultDomain, f.DNS.Domain)
	}
	if f.DNS.Port != DNSDefaultPort {
		t.Fatalf("expected port %d, got %d", DNSDefaultPort, f.DNS.Port)
	}
}

func TestResolveFeatureDependencies_DNSPreservesExplicitValues(t *testing.T) {
	f := FeaturesConfig{
		Gateway:  &GatewayConfig{Enabled: true},
		DNS:     &DNSConfig{Enabled: true, Domain: "custom.dev", Port: 9053},
	}
	warnings, err := ResolveFeatureDependencies(&f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if f.DNS.Domain != "custom.dev" {
		t.Fatalf("expected domain %q, got %q", "custom.dev", f.DNS.Domain)
	}
	if f.DNS.Port != 9053 {
		t.Fatalf("expected port %d, got %d", 9053, f.DNS.Port)
	}
}

func TestResolveFeatureDependencies_AllNilNoOp(t *testing.T) {
	f := FeaturesConfig{}
	warnings, err := ResolveFeatureDependencies(&f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}

func TestMergeWithContext_UserPortsReplaceContext(t *testing.T) {
	k := KindConfig{
		Name: KindDefaultName,
		Nodes: []KindNode{{
			Role: "control-plane",
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 9090, HostPort: 9090},
			},
		}},
	}
	ctx := &ContextKindConfig{
		ExtraPortMappings: []PortMapping{
			{ContainerPort: 80, HostPort: 80},
			{ContainerPort: 443, HostPort: 443},
		},
	}
	k.MergeWithContext(ctx)

	if len(k.Nodes[0].ExtraPortMappings) != 1 {
		t.Fatalf("expected 1 port mapping, got %d", len(k.Nodes[0].ExtraPortMappings))
	}
	if k.Nodes[0].ExtraPortMappings[0].ContainerPort != 9090 {
		t.Fatalf("expected containerPort 9090, got %d", k.Nodes[0].ExtraPortMappings[0].ContainerPort)
	}
}

func TestMergeWithContext_NoUserPortsFallsBackToContext(t *testing.T) {
	k := KindConfig{
		Name:  KindDefaultName,
		Nodes: []KindNode{{Role: "control-plane"}},
	}
	ctx := &ContextKindConfig{
		ExtraPortMappings: []PortMapping{
			{ContainerPort: 80, HostPort: 80},
			{ContainerPort: 443, HostPort: 443},
		},
	}
	k.MergeWithContext(ctx)

	if len(k.Nodes[0].ExtraPortMappings) != 2 {
		t.Fatalf("expected 2 port mappings, got %d", len(k.Nodes[0].ExtraPortMappings))
	}
	if k.Nodes[0].ExtraPortMappings[0].ContainerPort != 80 {
		t.Fatalf("expected containerPort 80, got %d", k.Nodes[0].ExtraPortMappings[0].ContainerPort)
	}
	if k.Nodes[0].ExtraPortMappings[1].ContainerPort != 443 {
		t.Fatalf("expected containerPort 443, got %d", k.Nodes[0].ExtraPortMappings[1].ContainerPort)
	}
}

func TestMergeWithDefaults_UserPortsReplaceDefaults(t *testing.T) {
	k := KindConfig{
		Name: KindDefaultName,
		Nodes: []KindNode{{
			Role: "control-plane",
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 9090, HostPort: 9090},
			},
		}},
	}
	defaults := &KindConfig{
		Nodes: []KindNode{{
			Role: "control-plane",
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 80, HostPort: 80},
				{ContainerPort: 443, HostPort: 443},
			},
		}},
	}
	k.MergeWithDefaults(defaults)

	if len(k.Nodes[0].ExtraPortMappings) != 1 {
		t.Fatalf("expected 1 port mapping, got %d", len(k.Nodes[0].ExtraPortMappings))
	}
	if k.Nodes[0].ExtraPortMappings[0].ContainerPort != 9090 {
		t.Fatalf("expected containerPort 9090, got %d", k.Nodes[0].ExtraPortMappings[0].ContainerPort)
	}
}

func TestMergeWithDefaults_NoUserPortsFallsBackToDefaults(t *testing.T) {
	k := KindConfig{
		Name:  KindDefaultName,
		Nodes: []KindNode{{Role: "control-plane"}},
	}
	defaults := &KindConfig{
		Nodes: []KindNode{{
			Role: "control-plane",
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 80, HostPort: 80},
				{ContainerPort: 443, HostPort: 443},
			},
		}},
	}
	k.MergeWithDefaults(defaults)

	if len(k.Nodes[0].ExtraPortMappings) != 2 {
		t.Fatalf("expected 2 port mappings, got %d", len(k.Nodes[0].ExtraPortMappings))
	}
	if k.Nodes[0].ExtraPortMappings[0].ContainerPort != 80 {
		t.Fatalf("expected containerPort 80, got %d", k.Nodes[0].ExtraPortMappings[0].ContainerPort)
	}
	if k.Nodes[0].ExtraPortMappings[1].ContainerPort != 443 {
		t.Fatalf("expected containerPort 443, got %d", k.Nodes[0].ExtraPortMappings[1].ContainerPort)
	}
}

func TestMergeWithContext_NilContextIsNoOp(t *testing.T) {
	k := KindConfig{
		Name:  KindDefaultName,
		Nodes: []KindNode{{Role: "control-plane", ExtraPortMappings: []PortMapping{{ContainerPort: 8080, HostPort: 8080}}}},
	}
	k.MergeWithContext(nil)

	if len(k.Nodes[0].ExtraPortMappings) != 1 || k.Nodes[0].ExtraPortMappings[0].ContainerPort != 8080 {
		t.Fatal("expected user ports to be untouched when context is nil")
	}
}

func TestMergeWithContext_NoUserNodesIsNoOp(t *testing.T) {
	k := KindConfig{Name: KindDefaultName}
	ctx := &ContextKindConfig{
		ExtraPortMappings: []PortMapping{{ContainerPort: 80, HostPort: 80}},
	}
	k.MergeWithContext(ctx)

	if len(k.Nodes) != 0 {
		t.Fatal("expected nodes to remain empty")
	}
}

func TestMergeWithContext_ContextTopLevelAndNodePortsMergedWhenUserHasNone(t *testing.T) {
	k := KindConfig{
		Name:  KindDefaultName,
		Nodes: []KindNode{{Role: "control-plane"}},
	}
	ctx := &ContextKindConfig{
		ExtraPortMappings: []PortMapping{
			{ContainerPort: 80, HostPort: 80},
			{ContainerPort: 443, HostPort: 443},
		},
		Nodes: []ContextKindNode{{
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 443, HostPort: 8443, Protocol: "TCP"},
				{ContainerPort: 8080, HostPort: 8080},
			},
		}},
	}
	k.MergeWithContext(ctx)

	if len(k.Nodes[0].ExtraPortMappings) != 3 {
		t.Fatalf("expected 3 merged port mappings (80, 443 overridden, 8080), got %d", len(k.Nodes[0].ExtraPortMappings))
	}
	portMap := make(map[int32]PortMapping)
	for _, p := range k.Nodes[0].ExtraPortMappings {
		portMap[p.ContainerPort] = p
	}
	if p, ok := portMap[443]; !ok || p.HostPort != 8443 {
		t.Fatal("expected node-level port 443 to override top-level (hostPort 8443)")
	}
	if _, ok := portMap[80]; !ok {
		t.Fatal("expected top-level port 80 to be present")
	}
	if _, ok := portMap[8080]; !ok {
		t.Fatal("expected node-level port 8080 to be present")
	}
}

func TestMergeWithContext_UserPortsReplaceContextEvenWithNodePorts(t *testing.T) {
	k := KindConfig{
		Name: KindDefaultName,
		Nodes: []KindNode{{
			Role:              "control-plane",
			ExtraPortMappings: []PortMapping{{ContainerPort: 9090, HostPort: 9090}},
		}},
	}
	ctx := &ContextKindConfig{
		ExtraPortMappings: []PortMapping{{ContainerPort: 80, HostPort: 80}},
		Nodes: []ContextKindNode{{
			ExtraPortMappings: []PortMapping{{ContainerPort: 443, HostPort: 443}},
		}},
	}
	k.MergeWithContext(ctx)

	if len(k.Nodes[0].ExtraPortMappings) != 1 {
		t.Fatalf("expected 1 port mapping, got %d", len(k.Nodes[0].ExtraPortMappings))
	}
	if k.Nodes[0].ExtraPortMappings[0].ContainerPort != 9090 {
		t.Fatalf("expected user port 9090, got %d", k.Nodes[0].ExtraPortMappings[0].ContainerPort)
	}
}

func TestMergeWithContext_ContextNameOverridesDefault(t *testing.T) {
	k := KindConfig{
		Name:  KindDefaultName,
		Nodes: []KindNode{{Role: "control-plane"}},
	}
	ctx := &ContextKindConfig{Name: "my-cluster"}
	k.MergeWithContext(ctx)

	if k.Name != "my-cluster" {
		t.Fatalf("expected name %q, got %q", "my-cluster", k.Name)
	}
}

func TestMergeWithContext_CustomNameNotOverriddenByContext(t *testing.T) {
	k := KindConfig{
		Name:  "user-cluster",
		Nodes: []KindNode{{Role: "control-plane"}},
	}
	ctx := &ContextKindConfig{Name: "ctx-cluster"}
	k.MergeWithContext(ctx)

	if k.Name != "user-cluster" {
		t.Fatalf("expected name %q preserved, got %q", "user-cluster", k.Name)
	}
}

func TestMergeWithDefaults_NilDefaultsIsNoOp(t *testing.T) {
	k := KindConfig{
		Name:  KindDefaultName,
		Nodes: []KindNode{{Role: "control-plane", ExtraPortMappings: []PortMapping{{ContainerPort: 8080, HostPort: 8080}}}},
	}
	k.MergeWithDefaults(nil)

	if len(k.Nodes[0].ExtraPortMappings) != 1 || k.Nodes[0].ExtraPortMappings[0].ContainerPort != 8080 {
		t.Fatal("expected user ports to be untouched when defaults is nil")
	}
}

func TestMergeWithDefaults_NoUserNodesCopiesDefaults(t *testing.T) {
	k := KindConfig{Name: KindDefaultName}
	defaults := &KindConfig{
		Nodes: []KindNode{{
			Role:              "control-plane",
			ExtraPortMappings: []PortMapping{{ContainerPort: 80, HostPort: 80}},
		}},
	}
	k.MergeWithDefaults(defaults)

	if len(k.Nodes) != 1 {
		t.Fatalf("expected 1 node copied from defaults, got %d", len(k.Nodes))
	}
	if len(k.Nodes[0].ExtraPortMappings) != 1 || k.Nodes[0].ExtraPortMappings[0].ContainerPort != 80 {
		t.Fatal("expected default port mappings to be copied")
	}
}

func TestResolveFeatureDependencies_DisabledGatewayNoAutoLB(t *testing.T) {
	f := FeaturesConfig{
		Gateway: &GatewayConfig{Enabled: false},
	}
	warnings, err := ResolveFeatureDependencies(&f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if f.LB != nil {
		t.Fatal("expected LB to remain nil when gateway is disabled")
	}
}
