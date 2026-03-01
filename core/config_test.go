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
		LB:      &LBConfig{Enabled: false},
		Gateway: &GatewayConfig{Enabled: true},
	}
	_, err := ResolveFeatureDependencies(&f)
	if err == nil {
		t.Fatal("expected error when gateway is enabled but lb is explicitly disabled")
	}
}

func TestResolveFeatureDependencies_DNSWithoutGatewayWarns(t *testing.T) {
	f := FeaturesConfig{
		DNS:     &DNSConfig{Enabled: true},
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
