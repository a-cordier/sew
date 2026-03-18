package installer

import (
	"testing"
)

func TestInstallOpts_ZeroValueHasDryRunFalse(t *testing.T) {
	var opts InstallOpts
	if opts.DryRun {
		t.Fatal("expected zero-value InstallOpts to have DryRun=false")
	}
}

func TestForType_Helm(t *testing.T) {
	inst, err := ForType("helm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := inst.(*HelmInstaller); !ok {
		t.Fatalf("expected *HelmInstaller, got %T", inst)
	}
}

func TestForType_K8s(t *testing.T) {
	inst, err := ForType("k8s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := inst.(*ManifestInstaller); !ok {
		t.Fatalf("expected *ManifestInstaller, got %T", inst)
	}
}

func TestForType_Unknown(t *testing.T) {
	_, err := ForType("terraform")
	if err == nil {
		t.Fatal("expected error for unknown component type")
	}
}
