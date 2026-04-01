package build

import (
	"os"
	"testing"

	"github.com/a-cordier/sew/internal/config"
	v1 "k8s.io/api/core/v1"
	"gopkg.in/yaml.v3"
)

func TestPodSpecReferencesImage_MatchInContainer(t *testing.T) {
	spec := v1.PodSpec{
		Containers: []v1.Container{
			{Name: "gateway", Image: "graviteeio/apim-gateway:latest"},
		},
	}
	if !podSpecReferencesImage(spec, "graviteeio/apim-gateway:latest") {
		t.Fatal("expected match for container image")
	}
}

func TestPodSpecReferencesImage_MatchInInitContainer(t *testing.T) {
	spec := v1.PodSpec{
		InitContainers: []v1.Container{
			{Name: "init", Image: "busybox:latest"},
		},
		Containers: []v1.Container{
			{Name: "app", Image: "myapp:v1"},
		},
	}
	if !podSpecReferencesImage(spec, "busybox:latest") {
		t.Fatal("expected match for init container image")
	}
}

func TestPodSpecReferencesImage_NoMatch(t *testing.T) {
	spec := v1.PodSpec{
		Containers: []v1.Container{
			{Name: "app", Image: "myapp:v1"},
		},
	}
	if podSpecReferencesImage(spec, "other:v2") {
		t.Fatal("expected no match for unrelated image")
	}
}

func TestPodSpecReferencesImage_EmptySpec(t *testing.T) {
	if podSpecReferencesImage(v1.PodSpec{}, "any:image") {
		t.Fatal("expected no match for empty pod spec")
	}
}

func TestPodSpecReferencesImage_MultipleContainersOneMatches(t *testing.T) {
	spec := v1.PodSpec{
		Containers: []v1.Container{
			{Name: "sidecar", Image: "envoy:v1"},
			{Name: "app", Image: "graviteeio/apim-gateway:latest"},
			{Name: "logger", Image: "fluentd:v1"},
		},
	}
	if !podSpecReferencesImage(spec, "graviteeio/apim-gateway:latest") {
		t.Fatal("expected match when one of multiple containers uses the image")
	}
}

func TestPodSpecReferencesImage_ExactMatchRequired(t *testing.T) {
	spec := v1.PodSpec{
		Containers: []v1.Container{
			{Name: "app", Image: "graviteeio/apim-gateway:latest-debian"},
		},
	}
	if podSpecReferencesImage(spec, "graviteeio/apim-gateway:latest") {
		t.Fatal("expected no match when tag differs")
	}
	if podSpecReferencesImage(spec, "graviteeio/apim-gateway") {
		t.Fatal("expected no match when tag is missing from search")
	}
}

func TestPodSpecReferencesImage_BothContainersAndInitContainers(t *testing.T) {
	spec := v1.PodSpec{
		InitContainers: []v1.Container{
			{Name: "init-db", Image: "flyway:v1"},
		},
		Containers: []v1.Container{
			{Name: "app", Image: "myapp:v1"},
		},
	}
	if !podSpecReferencesImage(spec, "flyway:v1") {
		t.Fatal("expected match in init containers")
	}
	if !podSpecReferencesImage(spec, "myapp:v1") {
		t.Fatal("expected match in regular containers")
	}
	if podSpecReferencesImage(spec, "unknown:v1") {
		t.Fatal("expected no match for image not in any container")
	}
}


func TestExpandBuildArgs_Nil(t *testing.T) {
	result := expandBuildArgs(nil)
	if result != nil {
		t.Fatalf("expected nil for nil input, got %v", result)
	}
}

func TestExpandBuildArgs_Empty(t *testing.T) {
	result := expandBuildArgs(map[string]string{})
	if result != nil {
		t.Fatalf("expected nil for empty input, got %v", result)
	}
}

func TestExpandBuildArgs_ExpandsEnvVars(t *testing.T) {
	t.Setenv("SEW_TEST_VAR", "expanded-value")

	args := map[string]string{
		"LITERAL":  "plain",
		"FROM_ENV": "$SEW_TEST_VAR",
	}
	result := expandBuildArgs(args)

	if len(result) != 2 {
		t.Fatalf("expected 2 build args, got %d", len(result))
	}
	if result["LITERAL"] == nil || *result["LITERAL"] != "plain" {
		t.Fatalf("expected LITERAL %q, got %v", "plain", result["LITERAL"])
	}
	if result["FROM_ENV"] == nil || *result["FROM_ENV"] != "expanded-value" {
		t.Fatalf("expected FROM_ENV %q, got %v", "expanded-value", result["FROM_ENV"])
	}
}

func TestExpandBuildArgs_BraceSyntax(t *testing.T) {
	t.Setenv("SEW_BRACE_VAR", "braced")

	result := expandBuildArgs(map[string]string{
		"ARG": "${SEW_BRACE_VAR}",
	})
	if result["ARG"] == nil || *result["ARG"] != "braced" {
		t.Fatalf("expected brace syntax expanded to %q, got %v", "braced", result["ARG"])
	}
}

func TestExpandBuildArgs_UnsetVarExpandsEmpty(t *testing.T) {
	t.Setenv("SEW_UNSET_GUARD", "")
	os.Unsetenv("SEW_DEFINITELY_UNSET_VAR_12345")

	result := expandBuildArgs(map[string]string{
		"MISSING": "$SEW_DEFINITELY_UNSET_VAR_12345",
	})
	if result["MISSING"] == nil || *result["MISSING"] != "" {
		t.Fatalf("expected unset var to expand to empty string, got %v", result["MISSING"])
	}
}

func TestExpandBuildArgs_CompoundValue(t *testing.T) {
	t.Setenv("SEW_TAG", "3.12.7")

	result := expandBuildArgs(map[string]string{
		"IMAGE": "docker.io/datawire/aes:$SEW_TAG",
	})
	if result["IMAGE"] == nil || *result["IMAGE"] != "docker.io/datawire/aes:3.12.7" {
		t.Fatalf("expected compound expansion, got %v", result["IMAGE"])
	}
}

func TestExpandBuildArgs_MultipleVarsInValue(t *testing.T) {
	t.Setenv("SEW_REPO", "datawire")
	t.Setenv("SEW_NAME", "aes")
	t.Setenv("SEW_VER", "3.12.7")

	result := expandBuildArgs(map[string]string{
		"FULL_REF": "docker.io/$SEW_REPO/$SEW_NAME:$SEW_VER",
	})
	want := "docker.io/datawire/aes:3.12.7"
	if result["FULL_REF"] == nil || *result["FULL_REF"] != want {
		t.Fatalf("expected %q, got %v", want, result["FULL_REF"])
	}
}

func TestExpandBuildArgs_AllKeysPreserved(t *testing.T) {
	args := map[string]string{
		"A": "1",
		"B": "2",
		"C": "3",
	}
	result := expandBuildArgs(args)
	if len(result) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(result))
	}
	for k := range args {
		if _, ok := result[k]; !ok {
			t.Fatalf("expected key %q to be present in result", k)
		}
		if result[k] == nil {
			t.Fatalf("expected non-nil pointer for key %q", k)
		}
	}
}

func TestExpandBuildArgs_PointersAreDistinct(t *testing.T) {
	args := map[string]string{
		"A": "same",
		"B": "same",
	}
	result := expandBuildArgs(args)
	if result["A"] == result["B"] {
		t.Fatal("expected distinct pointers for each key even with identical values")
	}
}

func TestExpandBuildArgs_IntegrationWithConfig(t *testing.T) {
	t.Setenv("HOME", "/home/dev")

	input := `
builds:
  - name: aes
    image: docker.io/datawire/aes:3.12.7
    dir: $HOME/src/edge-stack
    buildArgs:
      EMISSARY_BASE: docker.io/datawire/aes:3.12.7
      BUILD_VERSION: "1.0.0"
      SRC_DIR: $HOME/src
`
	var cfg config.Config
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	b := cfg.Builds[0]

	result := expandBuildArgs(b.BuildArgs)
	if len(result) != 3 {
		t.Fatalf("expected 3 build args, got %d", len(result))
	}
	if v := *result["EMISSARY_BASE"]; v != "docker.io/datawire/aes:3.12.7" {
		t.Fatalf("expected literal value preserved, got %q", v)
	}
	if v := *result["BUILD_VERSION"]; v != "1.0.0" {
		t.Fatalf("expected BUILD_VERSION %q, got %q", "1.0.0", v)
	}
	if v := *result["SRC_DIR"]; v != "/home/dev/src" {
		t.Fatalf("expected SRC_DIR expanded to %q, got %q", "/home/dev/src", v)
	}
}

