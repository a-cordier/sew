package core

import (
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestMergeFeatures_OverrideReplacesBase(t *testing.T) {
	base := FeaturesConfig{
		LoadBalancer: &LoadBalancerConfig{Enabled: true},
		Gateway:      &GatewayConfig{Enabled: true, Channel: GatewayChannelStandard},
		LocalDNS:     &LocalDNSConfig{Enabled: true, Domain: "ctx.local", Port: 5353},
	}
	override := FeaturesConfig{
		LoadBalancer: &LoadBalancerConfig{Enabled: false},
	}
	result := MergeFeatures(base, override)

	if result.LoadBalancer == nil || result.LoadBalancer.Enabled {
		t.Fatal("expected LoadBalancer to be overridden to disabled")
	}
	if result.Gateway == nil || !result.Gateway.Enabled {
		t.Fatal("expected Gateway to be preserved from base")
	}
	if result.LocalDNS == nil || result.LocalDNS.Domain != "ctx.local" {
		t.Fatal("expected LocalDNS to be preserved from base")
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
	if result.LoadBalancer != nil {
		t.Fatal("expected LoadBalancer to remain nil")
	}
}

func TestMergeFeatures_BothNil(t *testing.T) {
	result := MergeFeatures(FeaturesConfig{}, FeaturesConfig{})
	if result.LoadBalancer != nil || result.Gateway != nil || result.LocalDNS != nil {
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
	if f.LoadBalancer == nil || !f.LoadBalancer.Enabled {
		t.Fatal("expected LoadBalancer to be auto-enabled")
	}
}

func TestResolveFeatureDependencies_GatewayDefaultsChannelToStandard(t *testing.T) {
	f := FeaturesConfig{
		Gateway: &GatewayConfig{Enabled: true},
	}
	if _, err := ResolveFeatureDependencies(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Gateway.Channel != GatewayChannelStandard {
		t.Fatalf("expected channel %q, got %q", GatewayChannelStandard, f.Gateway.Channel)
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
		LoadBalancer: &LoadBalancerConfig{Enabled: false},
		Gateway:      &GatewayConfig{Enabled: true},
	}
	_, err := ResolveFeatureDependencies(&f)
	if err == nil {
		t.Fatal("expected error when gateway is enabled but load-balancer is explicitly disabled")
	}
}

func TestResolveFeatureDependencies_DNSWithoutGatewayWarns(t *testing.T) {
	f := FeaturesConfig{
		LocalDNS: &LocalDNSConfig{Enabled: true},
	}
	warnings, err := ResolveFeatureDependencies(&f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
}

func TestResolveFeatureDependencies_DNSAppliesDefaults(t *testing.T) {
	f := FeaturesConfig{
		LocalDNS: &LocalDNSConfig{Enabled: true},
	}
	if _, err := ResolveFeatureDependencies(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.LocalDNS.Domain != LocalDNSDefaultDomain {
		t.Fatalf("expected domain %q, got %q", LocalDNSDefaultDomain, f.LocalDNS.Domain)
	}
	if f.LocalDNS.Port != LocalDNSDefaultPort {
		t.Fatalf("expected port %d, got %d", LocalDNSDefaultPort, f.LocalDNS.Port)
	}
}

func TestResolveFeatureDependencies_DNSPreservesExplicitValues(t *testing.T) {
	f := FeaturesConfig{
		Gateway:  &GatewayConfig{Enabled: true},
		LocalDNS: &LocalDNSConfig{Enabled: true, Domain: "custom.dev", Port: 9053},
	}
	warnings, err := ResolveFeatureDependencies(&f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if f.LocalDNS.Domain != "custom.dev" {
		t.Fatalf("expected domain %q, got %q", "custom.dev", f.LocalDNS.Domain)
	}
	if f.LocalDNS.Port != 9053 {
		t.Fatalf("expected port %d, got %d", 9053, f.LocalDNS.Port)
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
	if f.LoadBalancer != nil {
		t.Fatal("expected LoadBalancer to remain nil when gateway is disabled")
	}
}
