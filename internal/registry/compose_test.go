package registry

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-cordier/sew/core"
)

// writeFile is a test helper that creates parent dirs and writes data.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFSResolver_ParentComposition(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "parent", "sew.yaml"), `
kind:
  name: parent-cluster
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 80
      hostPort: 80

repos:
  - name: base-repo
    url: https://example.com/base

components:
  - name: base-comp
    helm:
      chart: base/chart
`)

	writeFile(t, filepath.Join(root, "child", "sew.yaml"), `
context: parent

kind:
  name: child-cluster

components:
  - name: child-comp
    helm:
      chart: child/chart
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Kind.Name != "child-cluster" {
		t.Fatalf("expected kind name %q, got %q", "child-cluster", resolved.Kind.Name)
	}
	if len(resolved.Kind.Nodes) != 1 || len(resolved.Kind.Nodes[0].ExtraPortMappings) != 1 {
		t.Fatal("expected child to inherit parent port mappings (child has nodes but no ports)")
	}
	if resolved.Kind.Nodes[0].ExtraPortMappings[0].ContainerPort != 80 {
		t.Fatalf("expected inherited port 80, got %d", resolved.Kind.Nodes[0].ExtraPortMappings[0].ContainerPort)
	}
	if len(resolved.Repos) != 1 || resolved.Repos[0].Name != "base-repo" {
		t.Fatalf("expected parent repo to be inherited, got %v", resolved.Repos)
	}
	if len(resolved.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resolved.Components))
	}
	// Dir should point to the parent context (base)
	if resolved.Dir != filepath.Join(root, "parent") {
		t.Fatalf("expected Dir to be parent dir, got %q", resolved.Dir)
	}
}

func TestFSResolver_ParentComposition_ChildOverridesComponent(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "parent", "sew.yaml"), `
repos:
  - name: repo1
    url: https://example.com/repo1

components:
  - name: app
    helm:
      chart: repo1/app
      version: "1.0"
      values:
        key1: val1
        key2: val2
`)

	writeFile(t, filepath.Join(root, "child", "sew.yaml"), `
context: parent

components:
  - name: app
    helm:
      version: "2.0"
      values:
        key2: overridden
        key3: new
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resolved.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(resolved.Components))
	}
	comp := resolved.Components[0]
	if comp.Helm.Chart != "repo1/app" {
		t.Fatalf("expected chart preserved from parent, got %q", comp.Helm.Chart)
	}
	if comp.Helm.Version != "2.0" {
		t.Fatalf("expected version overridden to 2.0, got %q", comp.Helm.Version)
	}
	if comp.Helm.Values["key1"] != "val1" {
		t.Fatal("expected key1 preserved from parent")
	}
	if comp.Helm.Values["key2"] != "overridden" {
		t.Fatal("expected key2 overridden by child")
	}
	if comp.Helm.Values["key3"] != "new" {
		t.Fatal("expected key3 added by child")
	}
}

func TestFSResolver_ParentComposition_FeaturesMerge(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "parent", "sew.yaml"), `
features:
  lb:
    enabled: true
  dns:
    enabled: true
    domain: parent.local

components: []
`)

	writeFile(t, filepath.Join(root, "child", "sew.yaml"), `
context: parent

features:
  dns:
    enabled: true
    domain: child.local
    port: 5353

components: []
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Features.LB == nil || !resolved.Features.LB.Enabled {
		t.Fatal("expected LB inherited from parent")
	}
	if resolved.Features.DNS == nil || resolved.Features.DNS.Domain != "child.local" {
		t.Fatal("expected DNS domain overridden by child")
	}
	if resolved.Features.DNS.Port != 5353 {
		t.Fatalf("expected DNS port 5353, got %d", resolved.Features.DNS.Port)
	}
}

func TestFSResolver_CycleDetection(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "a", "sew.yaml"), `
context: b
components: []
`)
	writeFile(t, filepath.Join(root, "b", "sew.yaml"), `
context: a
components: []
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	_, err := resolver.Resolve(context.Background(), "a")
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error message, got: %v", err)
	}
}

func TestFSResolver_SelfCycleDetection(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "self", "sew.yaml"), `
context: self
components: []
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	_, err := resolver.Resolve(context.Background(), "self")
	if err == nil {
		t.Fatal("expected cycle detection error for self-reference")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error message, got: %v", err)
	}
}

func TestFSResolver_ThreeLevelComposition(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "grandparent", "sew.yaml"), `
kind:
  name: gp-cluster
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 80
      hostPort: 80

repos:
  - name: gp-repo
    url: https://example.com/gp

components:
  - name: gp-comp
    helm:
      chart: gp/chart
      values:
        gp-key: gp-val
`)

	writeFile(t, filepath.Join(root, "mid", "sew.yaml"), `
context: grandparent

repos:
  - name: mid-repo
    url: https://example.com/mid

components:
  - name: gp-comp
    helm:
      values:
        mid-key: mid-val
  - name: mid-comp
    helm:
      chart: mid/chart
`)

	writeFile(t, filepath.Join(root, "leaf", "sew.yaml"), `
context: mid

kind:
  name: leaf-cluster

components:
  - name: leaf-comp
    helm:
      chart: leaf/chart
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "leaf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Kind.Name != "leaf-cluster" {
		t.Fatalf("expected kind name %q, got %q", "leaf-cluster", resolved.Kind.Name)
	}
	if len(resolved.Kind.Nodes[0].ExtraPortMappings) != 1 {
		t.Fatal("expected grandparent port mappings inherited through chain")
	}

	repoNames := make(map[string]bool)
	for _, r := range resolved.Repos {
		repoNames[r.Name] = true
	}
	if !repoNames["gp-repo"] || !repoNames["mid-repo"] {
		t.Fatalf("expected both gp-repo and mid-repo, got %v", resolved.Repos)
	}

	compNames := make(map[string]bool)
	for _, c := range resolved.Components {
		compNames[c.Name] = true
	}
	if !compNames["gp-comp"] || !compNames["mid-comp"] || !compNames["leaf-comp"] {
		t.Fatalf("expected all three components, got names %v", compNames)
	}
}

func TestFSResolver_SameRegistryImplicit(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "base", "sew.yaml"), `
components:
  - name: shared
    helm:
      chart: shared/chart
`)

	// Child omits registry → uses the same FS registry
	writeFile(t, filepath.Join(root, "derived", "sew.yaml"), `
context: base

components:
  - name: extra
    helm:
      chart: extra/chart
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "derived")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resolved.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resolved.Components))
	}
}

func TestFSResolver_CrossRegistryComposition(t *testing.T) {
	regA := t.TempDir()
	regB := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(regA, "parent", "sew.yaml"), `
components:
  - name: from-a
    helm:
      chart: a/chart
`)

	writeFile(t, filepath.Join(regB, "child", "sew.yaml"), `
registry: file://`+regA+`
context: parent

components:
  - name: from-b
    helm:
      chart: b/chart
`)

	resolver := &FSResolver{Root: regB, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resolved.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resolved.Components))
	}
	names := map[string]bool{}
	for _, c := range resolved.Components {
		names[c.Name] = true
	}
	if !names["from-a"] || !names["from-b"] {
		t.Fatalf("expected both from-a and from-b, got %v", names)
	}
}

func TestFSResolver_RelativeFileRegistryURL(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "org", "base", "sew.yaml"), `
kind:
  name: base-cluster

components:
  - name: base-comp
    helm:
      chart: base/chart
`)

	writeFile(t, filepath.Join(root, "org", "variants", "custom", "sew.yaml"), `
registry: file://../..
context: base

kind:
  name: custom-cluster

components: []
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "org/variants/custom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Kind.Name != "custom-cluster" {
		t.Fatalf("expected kind name %q, got %q", "custom-cluster", resolved.Kind.Name)
	}
	if len(resolved.Components) != 1 {
		t.Fatalf("expected 1 component from parent, got %d", len(resolved.Components))
	}
}

func TestFSResolver_NoParent(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "standalone", "sew.yaml"), `
kind:
  name: standalone

components:
  - name: app
    helm:
      chart: app/chart
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "standalone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Kind.Name != "standalone" {
		t.Fatalf("expected kind name %q, got %q", "standalone", resolved.Kind.Name)
	}
	if len(resolved.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(resolved.Components))
	}
}

func TestMergeKind_ChildNameWins(t *testing.T) {
	base := core.KindConfig{Name: "base-cluster"}
	child := core.KindConfig{Name: "child-cluster"}
	result := mergeKind(base, child)
	if result.Name != "child-cluster" {
		t.Fatalf("expected %q, got %q", "child-cluster", result.Name)
	}
}

func TestMergeKind_EmptyChildInheritsBase(t *testing.T) {
	base := core.KindConfig{
		Name:       "base-cluster",
		APIVersion: "v1alpha4",
		Nodes: []core.KindNode{{
			Role:              "control-plane",
			ExtraPortMappings: []core.PortMapping{{ContainerPort: 80, HostPort: 80}},
		}},
	}
	child := core.KindConfig{}
	result := mergeKind(base, child)

	if result.Name != "base-cluster" {
		t.Fatalf("expected name inherited, got %q", result.Name)
	}
	if result.APIVersion != "v1alpha4" {
		t.Fatalf("expected apiVersion inherited, got %q", result.APIVersion)
	}
	if len(result.Nodes) != 1 || result.Nodes[0].ExtraPortMappings[0].ContainerPort != 80 {
		t.Fatal("expected nodes inherited from base")
	}
}

func TestMergeKind_ChildNodesInheritParentPorts(t *testing.T) {
	base := core.KindConfig{
		Nodes: []core.KindNode{{
			Role:              "control-plane",
			ExtraPortMappings: []core.PortMapping{{ContainerPort: 80, HostPort: 80}},
		}},
	}
	child := core.KindConfig{
		Nodes: []core.KindNode{{Role: "control-plane"}},
	}
	result := mergeKind(base, child)

	if len(result.Nodes[0].ExtraPortMappings) != 1 {
		t.Fatalf("expected child to inherit parent port mappings, got %d", len(result.Nodes[0].ExtraPortMappings))
	}
}

func TestMergeKind_ChildPortsWin(t *testing.T) {
	base := core.KindConfig{
		Nodes: []core.KindNode{{
			Role:              "control-plane",
			ExtraPortMappings: []core.PortMapping{{ContainerPort: 80, HostPort: 80}},
		}},
	}
	child := core.KindConfig{
		Nodes: []core.KindNode{{
			Role:              "control-plane",
			ExtraPortMappings: []core.PortMapping{{ContainerPort: 9090, HostPort: 9090}},
		}},
	}
	result := mergeKind(base, child)

	if len(result.Nodes[0].ExtraPortMappings) != 1 {
		t.Fatalf("expected 1 port mapping, got %d", len(result.Nodes[0].ExtraPortMappings))
	}
	if result.Nodes[0].ExtraPortMappings[0].ContainerPort != 9090 {
		t.Fatalf("expected child port 9090, got %d", result.Nodes[0].ExtraPortMappings[0].ContainerPort)
	}
}

func TestWithVisited_DetectsCycle(t *testing.T) {
	ctx := context.Background()
	ref := contextRef{Registry: "file:///reg", Context: "path/a"}

	ctx, err := withVisited(ctx, ref)
	if err != nil {
		t.Fatalf("first visit should not error: %v", err)
	}

	_, err = withVisited(ctx, ref)
	if err == nil {
		t.Fatal("expected cycle error on second visit")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error message, got: %v", err)
	}
}

func TestWithVisited_DifferentRefsOK(t *testing.T) {
	ctx := context.Background()

	ctx, err := withVisited(ctx, contextRef{Registry: "file:///reg", Context: "a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = withVisited(ctx, contextRef{Registry: "file:///reg", Context: "b"})
	if err != nil {
		t.Fatalf("different context should not error: %v", err)
	}
}

func TestResolveRegistryURL_AbsoluteFileURL(t *testing.T) {
	result := resolveRegistryURL("file:///absolute/path", "/some/dir")
	if result != "file:///absolute/path" {
		t.Fatalf("expected absolute path preserved, got %q", result)
	}
}

func TestResolveRegistryURL_RelativeFileURL(t *testing.T) {
	result := resolveRegistryURL("file://./..", "/some/context/dir")
	if result != "file:///some/context" {
		t.Fatalf("expected resolved to %q, got %q", "file:///some/context", result)
	}
}

func TestResolveRegistryURL_HTTPPassthrough(t *testing.T) {
	result := resolveRegistryURL("https://example.com/registry", "/some/dir")
	if result != "https://example.com/registry" {
		t.Fatalf("expected HTTP URL unchanged, got %q", result)
	}
}

func TestFSResolver_ParentComposition_ImagesMerge(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "parent", "sew.yaml"), `
images:
  preload:
    refs:
      - img-a
      - img-b

components:
  - name: app
    helm:
      chart: app/chart
`)

	writeFile(t, filepath.Join(root, "child", "sew.yaml"), `
context: parent

images:
  preload:
    refs:
      - img-b
      - img-c

components: []
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Images.Preload == nil {
		t.Fatal("expected Preload inherited from parent")
	}
	if len(resolved.Images.Preload.Refs) != 3 {
		t.Fatalf("expected 3 refs (deduped union), got %d: %v", len(resolved.Images.Preload.Refs), resolved.Images.Preload.Refs)
	}
	expected := map[string]bool{"img-a": true, "img-b": true, "img-c": true}
	for _, ref := range resolved.Images.Preload.Refs {
		if !expected[ref] {
			t.Fatalf("unexpected ref %q", ref)
		}
	}
}

func TestFSResolver_NoParent_ImagesPreserved(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "standalone", "sew.yaml"), `
images:
  preload:
    refs:
      - my-image:latest

components:
  - name: app
    helm:
      chart: app/chart
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "standalone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Images.Preload == nil {
		t.Fatal("expected Preload to be non-nil")
	}
	if len(resolved.Images.Preload.Refs) != 1 || resolved.Images.Preload.Refs[0] != "my-image:latest" {
		t.Fatalf("expected [my-image:latest], got %v", resolved.Images.Preload.Refs)
	}
}

func TestFSResolver_ContextYAMLFallback(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "legacy", "context.yaml"), `
kind:
  name: legacy-cluster

components:
  - name: legacy-app
    helm:
      chart: legacy/chart
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "legacy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Kind.Name != "legacy-cluster" {
		t.Fatalf("expected kind name %q, got %q", "legacy-cluster", resolved.Kind.Name)
	}
	if len(resolved.Components) != 1 || resolved.Components[0].Name != "legacy-app" {
		t.Fatal("expected legacy-app component")
	}
}

func TestFSResolver_DefaultVariant(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "product", ".default"), `standard`)
	writeFile(t, filepath.Join(root, "product", "standard", "sew.yaml"), `
kind:
  name: standard-cluster

components:
  - name: standard-app
    helm:
      chart: std/chart
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "product")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Kind.Name != "standard-cluster" {
		t.Fatalf("expected kind name %q, got %q", "standard-cluster", resolved.Kind.Name)
	}
	if len(resolved.Components) != 1 || resolved.Components[0].Name != "standard-app" {
		t.Fatal("expected standard-app component")
	}
}

func TestFSResolver_DefaultVariantWithComposition(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "base", "sew.yaml"), `
repos:
  - name: base-repo
    url: https://example.com/base

components:
  - name: base-comp
    helm:
      chart: base/chart
`)
	writeFile(t, filepath.Join(root, "product", ".default"), `standard`)
	writeFile(t, filepath.Join(root, "product", "standard", "sew.yaml"), `
context: base

components:
  - name: extra
    helm:
      chart: extra/chart
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	resolved, err := resolver.Resolve(context.Background(), "product")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resolved.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resolved.Components))
	}
	if len(resolved.Repos) != 1 || resolved.Repos[0].Name != "base-repo" {
		t.Fatalf("expected base repo inherited, got %v", resolved.Repos)
	}
}

func TestFSResolver_ThreeLevelCycleDetection(t *testing.T) {
	root := t.TempDir()
	sewHome := t.TempDir()

	writeFile(t, filepath.Join(root, "a", "sew.yaml"), `
context: b
components: []
`)
	writeFile(t, filepath.Join(root, "b", "sew.yaml"), `
context: c
components: []
`)
	writeFile(t, filepath.Join(root, "c", "sew.yaml"), `
context: a
components: []
`)

	resolver := &FSResolver{Root: root, SewHome: sewHome}
	_, err := resolver.Resolve(context.Background(), "a")
	if err == nil {
		t.Fatal("expected cycle detection error for a→b→c→a")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error message, got: %v", err)
	}
}
