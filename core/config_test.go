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
	ctx := &KindConfig{
		Nodes: []KindNode{{
			Role: "control-plane",
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 80, HostPort: 80},
				{ContainerPort: 443, HostPort: 443},
			},
		}},
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
	ctx := &KindConfig{
		Nodes: []KindNode{{
			Role: "control-plane",
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 80, HostPort: 80},
				{ContainerPort: 443, HostPort: 443},
			},
		}},
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

func TestMergeImages_BothEmpty(t *testing.T) {
	result := MergeImages(ImagesConfig{}, ImagesConfig{})
	if result.Preload != nil {
		t.Fatal("expected Preload to be nil")
	}
	if result.Mirrors != nil {
		t.Fatal("expected Mirrors to be nil")
	}
}

func TestMergeImages_BasePreloadPreserved(t *testing.T) {
	base := ImagesConfig{
		Preload: &PreloadConfig{Refs: []string{"img-a", "img-b"}},
	}
	result := MergeImages(base, ImagesConfig{})
	if result.Preload == nil {
		t.Fatal("expected Preload preserved from base")
	}
	if len(result.Preload.Refs) != 2 || result.Preload.Refs[0] != "img-a" || result.Preload.Refs[1] != "img-b" {
		t.Fatalf("expected base refs preserved, got %v", result.Preload.Refs)
	}
}

func TestMergeImages_OverrideRefsAppended(t *testing.T) {
	base := ImagesConfig{
		Preload: &PreloadConfig{Refs: []string{"img-a", "img-b"}},
	}
	override := ImagesConfig{
		Preload: &PreloadConfig{Refs: []string{"img-c"}},
	}
	result := MergeImages(base, override)
	if len(result.Preload.Refs) != 3 {
		t.Fatalf("expected 3 refs, got %d: %v", len(result.Preload.Refs), result.Preload.Refs)
	}
}

func TestMergeImages_RefsDeduplicated(t *testing.T) {
	base := ImagesConfig{
		Preload: &PreloadConfig{Refs: []string{"img-a", "img-b"}},
	}
	override := ImagesConfig{
		Preload: &PreloadConfig{Refs: []string{"img-b", "img-c"}},
	}
	result := MergeImages(base, override)
	if len(result.Preload.Refs) != 3 {
		t.Fatalf("expected 3 refs (deduped), got %d: %v", len(result.Preload.Refs), result.Preload.Refs)
	}
	expected := []string{"img-a", "img-b", "img-c"}
	for i, v := range expected {
		if result.Preload.Refs[i] != v {
			t.Fatalf("expected ref[%d]=%q, got %q", i, v, result.Preload.Refs[i])
		}
	}
}

func TestMergeImages_PreloadFromEitherSide(t *testing.T) {
	base := ImagesConfig{Preload: &PreloadConfig{Refs: []string{"img-a"}}}
	result := MergeImages(base, ImagesConfig{})
	if result.Preload == nil {
		t.Fatal("expected Preload preserved from base")
	}

	result = MergeImages(ImagesConfig{}, ImagesConfig{Preload: &PreloadConfig{Refs: []string{"img-b"}}})
	if result.Preload == nil || result.Preload.Refs[0] != "img-b" {
		t.Fatal("expected Preload from override")
	}
}

func TestMergeImages_MirrorsOverrideWins(t *testing.T) {
	base := ImagesConfig{
		Mirrors: &MirrorsConfig{Upstreams: []string{"docker.io"}},
	}
	override := ImagesConfig{
		Mirrors: &MirrorsConfig{Upstreams: []string{"ghcr.io"}},
	}
	result := MergeImages(base, override)
	if len(result.Mirrors.Upstreams) != 1 || result.Mirrors.Upstreams[0] != "ghcr.io" {
		t.Fatalf("expected override mirrors to win, got %v", result.Mirrors.Upstreams)
	}
}

func TestMergeImages_NilMirrorsOverrideKeepsBase(t *testing.T) {
	base := ImagesConfig{
		Mirrors: &MirrorsConfig{Upstreams: []string{"docker.io"}},
	}
	result := MergeImages(base, ImagesConfig{})
	if result.Mirrors == nil || result.Mirrors.Upstreams[0] != "docker.io" {
		t.Fatal("expected base mirrors preserved when override is nil")
	}
}

func TestMergeImages_NilBasePreloadOverrideWins(t *testing.T) {
	override := ImagesConfig{
		Preload: &PreloadConfig{Refs: []string{"img-x"}},
	}
	result := MergeImages(ImagesConfig{}, override)
	if result.Preload == nil || len(result.Preload.Refs) != 1 || result.Preload.Refs[0] != "img-x" {
		t.Fatalf("expected override preload to be used, got %v", result.Preload)
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
	ctx := &KindConfig{
		Nodes: []KindNode{{
			Role:              "control-plane",
			ExtraPortMappings: []PortMapping{{ContainerPort: 80, HostPort: 80}},
		}},
	}
	k.MergeWithContext(ctx)

	if len(k.Nodes) != 0 {
		t.Fatal("expected nodes to remain empty")
	}
}

func TestMergeWithContext_ContextNodePortsAppliedWhenUserHasNone(t *testing.T) {
	k := KindConfig{
		Name:  KindDefaultName,
		Nodes: []KindNode{{Role: "control-plane"}},
	}
	ctx := &KindConfig{
		Nodes: []KindNode{{
			Role: "control-plane",
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 80, HostPort: 80},
				{ContainerPort: 443, HostPort: 8443, Protocol: "TCP"},
				{ContainerPort: 8080, HostPort: 8080},
			},
		}},
	}
	k.MergeWithContext(ctx)

	if len(k.Nodes[0].ExtraPortMappings) != 3 {
		t.Fatalf("expected 3 port mappings, got %d", len(k.Nodes[0].ExtraPortMappings))
	}
	portMap := make(map[int32]PortMapping)
	for _, p := range k.Nodes[0].ExtraPortMappings {
		portMap[p.ContainerPort] = p
	}
	if p, ok := portMap[443]; !ok || p.HostPort != 8443 {
		t.Fatal("expected port 443 with hostPort 8443")
	}
	if _, ok := portMap[80]; !ok {
		t.Fatal("expected port 80 to be present")
	}
	if _, ok := portMap[8080]; !ok {
		t.Fatal("expected port 8080 to be present")
	}
}

func TestMergeWithContext_UserPortsReplaceContextNodePorts(t *testing.T) {
	k := KindConfig{
		Name: KindDefaultName,
		Nodes: []KindNode{{
			Role:              "control-plane",
			ExtraPortMappings: []PortMapping{{ContainerPort: 9090, HostPort: 9090}},
		}},
	}
	ctx := &KindConfig{
		Nodes: []KindNode{{
			Role: "control-plane",
			ExtraPortMappings: []PortMapping{
				{ContainerPort: 80, HostPort: 80},
				{ContainerPort: 443, HostPort: 443},
			},
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
	ctx := &KindConfig{Name: "my-cluster"}
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
	ctx := &KindConfig{Name: "ctx-cluster"}
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
