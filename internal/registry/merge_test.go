package registry

import (
	"path/filepath"
	"testing"

	"github.com/a-cordier/sew/internal/config"
)

func TestMergeComponents_MatchingHelmComponent(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name: "app",
				Helm: &config.HelmSpec{
					Chart:   "repo/app",
					Version: "1.0",
					Values: map[string]interface{}{
						"key1": "base-val",
					},
				},
			},
		},
	}
	overrides := []config.Component{
		{
			Name: "app",
			Helm: &config.HelmSpec{
				Version: "2.0",
				Values: map[string]interface{}{
					"key1": "override-val",
					"key2": "new-val",
				},
			},
		},
	}

	MergeComponents(resolved, overrides, "")

	if len(resolved.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(resolved.Components))
	}
	comp := resolved.Components[0]
	if comp.Helm.Chart != "repo/app" {
		t.Fatalf("expected chart preserved, got %q", comp.Helm.Chart)
	}
	if comp.Helm.Version != "2.0" {
		t.Fatalf("expected version overridden, got %q", comp.Helm.Version)
	}
	if comp.Helm.Values["key1"] != "override-val" {
		t.Fatal("expected key1 overridden")
	}
	if comp.Helm.Values["key2"] != "new-val" {
		t.Fatal("expected key2 added")
	}
}

func TestMergeComponents_NewComponent(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{Name: "existing", Helm: &config.HelmSpec{Chart: "existing/chart"}},
		},
	}
	overrides := []config.Component{
		{Name: "added", Helm: &config.HelmSpec{Chart: "added/chart"}},
	}

	MergeComponents(resolved, overrides, "")

	if len(resolved.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resolved.Components))
	}
	names := make(map[string]bool)
	for _, c := range resolved.Components {
		names[c.Name] = true
	}
	if !names["existing"] || !names["added"] {
		t.Fatalf("expected both existing and added, got %v", names)
	}
}

func TestMergeComponents_EmptyOverrides(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{Name: "app", Helm: &config.HelmSpec{Chart: "app/chart"}},
		},
	}

	MergeComponents(resolved, nil, "")

	if len(resolved.Components) != 1 {
		t.Fatalf("expected 1 component unchanged, got %d", len(resolved.Components))
	}
}

func TestMergeComponents_RequirementsDedup(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name: "app",
				Helm: &config.HelmSpec{Chart: "app/chart"},
				Requires: []config.Requirement{
					{Component: "db"},
				},
			},
		},
	}
	overrides := []config.Component{
		{
			Name: "app",
			Helm: &config.HelmSpec{},
			Requires: []config.Requirement{
				{Component: "db"},
				{Component: "cache"},
			},
		},
	}

	MergeComponents(resolved, overrides, "")

	comp := resolved.Components[0]
	if len(comp.Requires) != 2 {
		t.Fatalf("expected 2 requirements (deduped), got %d", len(comp.Requires))
	}
	reqNames := make(map[string]bool)
	for _, r := range comp.Requires {
		reqNames[r.Component] = true
	}
	if !reqNames["db"] || !reqNames["cache"] {
		t.Fatalf("expected db and cache, got %v", reqNames)
	}
}

func TestMergeComponents_K8sManifestsMerge(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name: "routes",
				K8s: &config.K8sSpec{
					ManifestFiles: []string{"/abs/path/base.yaml"},
				},
			},
		},
	}
	overrides := []config.Component{
		{
			Name: "routes",
			K8s: &config.K8sSpec{
				ManifestFiles: []string{"extra.yaml"},
			},
		},
	}

	MergeComponents(resolved, overrides, "/config/dir")

	comp := resolved.Components[0]
	if len(comp.K8s.ManifestFiles) != 2 {
		t.Fatalf("expected 2 manifest files, got %d", len(comp.K8s.ManifestFiles))
	}
	if comp.K8s.ManifestFiles[0] != "/abs/path/base.yaml" {
		t.Fatalf("expected base manifest preserved, got %q", comp.K8s.ManifestFiles[0])
	}
	expected := filepath.Join("/config/dir", "extra.yaml")
	if comp.K8s.ManifestFiles[1] != expected {
		t.Fatalf("expected extra manifest resolved to %q, got %q", expected, comp.K8s.ManifestFiles[1])
	}
}

func TestMergeComponents_ValueFilePathResolution(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name: "app",
				Helm: &config.HelmSpec{
					Chart:      "app/chart",
					ValueFiles: []string{"/abs/values.yaml"},
				},
			},
		},
	}
	overrides := []config.Component{
		{
			Name: "app",
			Helm: &config.HelmSpec{
				ValueFiles: []string{"local-values.yaml"},
			},
		},
	}

	MergeComponents(resolved, overrides, "/user/config")

	comp := resolved.Components[0]
	if len(comp.Helm.ValueFiles) != 2 {
		t.Fatalf("expected 2 value files, got %d", len(comp.Helm.ValueFiles))
	}
	if comp.Helm.ValueFiles[0] != "/abs/values.yaml" {
		t.Fatalf("expected base value file preserved, got %q", comp.Helm.ValueFiles[0])
	}
	expected := filepath.Join("/user/config", "local-values.yaml")
	if comp.Helm.ValueFiles[1] != expected {
		t.Fatalf("expected local value file resolved to %q, got %q", expected, comp.Helm.ValueFiles[1])
	}
}

func TestMergeComponents_ChartOverride(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name: "app",
				Helm: &config.HelmSpec{
					Chart:   "original/chart",
					Version: "1.0",
				},
			},
		},
	}
	overrides := []config.Component{
		{
			Name: "app",
			Helm: &config.HelmSpec{
				Chart: "custom/chart",
			},
		},
	}

	MergeComponents(resolved, overrides, "")

	comp := resolved.Components[0]
	if comp.Helm.Chart != "custom/chart" {
		t.Fatalf("expected chart overridden, got %q", comp.Helm.Chart)
	}
	if comp.Helm.Version != "1.0" {
		t.Fatalf("expected version preserved, got %q", comp.Helm.Version)
	}
}

func TestMergeComponents_K8sOnNewComponent(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{Name: "app", Helm: &config.HelmSpec{Chart: "app/chart"}},
		},
	}
	overrides := []config.Component{
		{
			Name: "routes",
			K8s: &config.K8sSpec{
				ManifestFiles: []string{"gateway.yaml"},
			},
		},
	}

	MergeComponents(resolved, overrides, "/config")

	if len(resolved.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resolved.Components))
	}
	routes := resolved.Components[1]
	if routes.Name != "routes" {
		t.Fatalf("expected routes component, got %q", routes.Name)
	}
	expected := filepath.Join("/config", "gateway.yaml")
	if routes.K8s.ManifestFiles[0] != expected {
		t.Fatalf("expected manifest path resolved to %q, got %q", expected, routes.K8s.ManifestFiles[0])
	}
}

func TestMergeComponents_ConditionsSelectorTimeout(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name: "apim",
				Helm: &config.HelmSpec{Chart: "apim/chart"},
			},
		},
	}
	overrides := []config.Component{
		{
			Name:       "apim",
			Conditions: config.Conditions{Ready: true},
			Selector: &config.Selector{
				MatchLabels: map[string]string{"app": "apim"},
			},
			Timeout: "10m",
			Helm:    &config.HelmSpec{},
		},
	}

	MergeComponents(resolved, overrides, "")

	comp := resolved.Components[0]
	if !comp.Conditions.Ready {
		t.Fatal("expected conditions.ready to be true")
	}
	if comp.Selector == nil || comp.Selector.MatchLabels["app"] != "apim" {
		t.Fatalf("expected selector overridden, got %v", comp.Selector)
	}
	if comp.Timeout != "10m" {
		t.Fatalf("expected timeout overridden, got %q", comp.Timeout)
	}
}

func TestMergeComponents_NamespaceOverride(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name: "mongodb",
				K8s:  &config.K8sSpec{},
			},
		},
	}
	overrides := []config.Component{
		{
			Name:      "mongodb",
			Namespace: "gravitee",
		},
	}

	MergeComponents(resolved, overrides, "")

	comp := resolved.Components[0]
	if comp.Namespace != "gravitee" {
		t.Fatalf("expected namespace overridden to %q, got %q", "gravitee", comp.Namespace)
	}
}

func TestMergeComponents_NamespacePreservedWhenUnset(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name:      "mongodb",
				Namespace: "infra",
				K8s:       &config.K8sSpec{},
			},
		},
	}
	overrides := []config.Component{
		{
			Name: "mongodb",
		},
	}

	MergeComponents(resolved, overrides, "")

	comp := resolved.Components[0]
	if comp.Namespace != "infra" {
		t.Fatalf("expected namespace preserved as %q, got %q", "infra", comp.Namespace)
	}
}

func TestMergeComponents_ConditionsNotOverriddenWhenUnset(t *testing.T) {
	resolved := &config.ResolvedContext{
		Components: []config.Component{
			{
				Name:       "apim",
				Conditions: config.Conditions{Ready: true},
				Selector: &config.Selector{
					MatchLabels: map[string]string{"app": "apim"},
				},
				Timeout: "5m",
				Helm:    &config.HelmSpec{Chart: "apim/chart"},
			},
		},
	}
	overrides := []config.Component{
		{
			Name: "apim",
			Helm: &config.HelmSpec{},
		},
	}

	MergeComponents(resolved, overrides, "")

	comp := resolved.Components[0]
	if !comp.Conditions.Ready {
		t.Fatal("expected conditions.ready preserved")
	}
	if comp.Selector == nil || comp.Selector.MatchLabels["app"] != "apim" {
		t.Fatal("expected selector preserved")
	}
	if comp.Timeout != "5m" {
		t.Fatalf("expected timeout preserved, got %q", comp.Timeout)
	}
}

func TestMergeRepos_NoOverlap(t *testing.T) {
	ctx := []config.Repo{
		{Name: "repo-a", URL: "https://a.example.com"},
	}
	local := []config.Repo{
		{Name: "repo-b", URL: "https://b.example.com"},
	}

	result := MergeRepos(ctx, local)

	if len(result) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(result))
	}
}

func TestMergeRepos_OverlapLocalWins(t *testing.T) {
	ctx := []config.Repo{
		{Name: "shared", URL: "https://old.example.com"},
		{Name: "ctx-only", URL: "https://ctx.example.com"},
	}
	local := []config.Repo{
		{Name: "shared", URL: "https://new.example.com"},
	}

	result := MergeRepos(ctx, local)

	if len(result) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(result))
	}
	repoByName := make(map[string]string)
	for _, r := range result {
		repoByName[r.Name] = r.URL
	}
	if repoByName["shared"] != "https://new.example.com" {
		t.Fatalf("expected local URL to win, got %q", repoByName["shared"])
	}
	if repoByName["ctx-only"] != "https://ctx.example.com" {
		t.Fatalf("expected ctx-only preserved, got %q", repoByName["ctx-only"])
	}
}

func TestMergeRepos_EmptyLocal(t *testing.T) {
	ctx := []config.Repo{
		{Name: "repo", URL: "https://example.com"},
	}

	result := MergeRepos(ctx, nil)

	if len(result) != 1 || result[0].Name != "repo" {
		t.Fatalf("expected context repos preserved, got %v", result)
	}
}

func TestMergeRepos_EmptyContext(t *testing.T) {
	local := []config.Repo{
		{Name: "repo", URL: "https://example.com"},
	}

	result := MergeRepos(nil, local)

	if len(result) != 1 || result[0].Name != "repo" {
		t.Fatalf("expected local repos returned, got %v", result)
	}
}

func TestMergeRepos_OrderPreserved(t *testing.T) {
	ctx := []config.Repo{
		{Name: "alpha", URL: "https://alpha.example.com"},
		{Name: "beta", URL: "https://beta.example.com"},
		{Name: "gamma", URL: "https://gamma.example.com"},
	}
	local := []config.Repo{
		{Name: "beta", URL: "https://beta-override.example.com"},
		{Name: "delta", URL: "https://delta.example.com"},
	}

	result := MergeRepos(ctx, local)

	if len(result) != 4 {
		t.Fatalf("expected 4 repos, got %d", len(result))
	}
	if result[0].Name != "alpha" {
		t.Fatalf("expected alpha first, got %q", result[0].Name)
	}
	if result[1].Name != "beta" || result[1].URL != "https://beta-override.example.com" {
		t.Fatalf("expected beta (overridden) second, got %v", result[1])
	}
	if result[2].Name != "gamma" {
		t.Fatalf("expected gamma third, got %q", result[2].Name)
	}
	if result[3].Name != "delta" {
		t.Fatalf("expected delta fourth, got %q", result[3].Name)
	}
}
