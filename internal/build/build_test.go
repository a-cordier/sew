package build

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
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

// drainBuildOutput tests

func TestDrainBuildOutput_Success(t *testing.T) {
	input := strings.NewReader(`{"stream":"Step 1/3 : FROM alpine"}
{"stream":"Step 2/3 : RUN echo hello"}
{"stream":"Successfully built abc123"}
`)
	if err := drainBuildOutput(input, io.Discard); err != nil {
		t.Fatalf("expected no error for successful build output, got: %v", err)
	}
}

func TestDrainBuildOutput_DetectsError(t *testing.T) {
	input := strings.NewReader(`{"stream":"Step 1/2 : FROM nonexistent"}
{"error":"pull access denied for nonexistent, repository does not exist"}
`)
	err := drainBuildOutput(input, io.Discard)
	if err == nil {
		t.Fatal("expected error when build output contains an error message")
	}
	if !strings.Contains(err.Error(), "pull access denied") {
		t.Fatalf("expected error to contain the Docker error message, got: %v", err)
	}
}

func TestDrainBuildOutput_MalformedJSON(t *testing.T) {
	input := strings.NewReader("not json at all\n{\"stream\":\"ok\"}\n")
	if err := drainBuildOutput(input, io.Discard); err != nil {
		t.Fatalf("expected malformed lines to be skipped, got: %v", err)
	}
}

func TestDrainBuildOutput_Empty(t *testing.T) {
	input := strings.NewReader("")
	if err := drainBuildOutput(input, io.Discard); err != nil {
		t.Fatalf("expected no error for empty output, got: %v", err)
	}
}

// createBuildContext tests

func TestCreateBuildContext_DockerfileInsideContext(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r, dockerfileInTar, err := createBuildContext(dir, filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dockerfileInTar != "Dockerfile" {
		t.Fatalf("expected dockerfileInTar %q, got %q", "Dockerfile", dockerfileInTar)
	}

	files := tarEntries(t, r)
	if !files["Dockerfile"] {
		t.Fatal("expected Dockerfile in tar archive")
	}
	if !files["main.go"] {
		t.Fatal("expected main.go in tar archive")
	}
}

func TestCreateBuildContext_DockerfileOutsideContext(t *testing.T) {
	root := t.TempDir()
	contextDir := filepath.Join(root, "src")
	if err := os.MkdirAll(contextDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contextDir, "app.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dockerfilePath := filepath.Join(root, "docker", "Dockerfile")
	if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dockerfilePath, []byte("FROM golang\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r, dockerfileInTar, err := createBuildContext(contextDir, dockerfilePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dockerfileInTar != ".sew.Dockerfile" {
		t.Fatalf("expected synthetic dockerfile name %q, got %q", ".sew.Dockerfile", dockerfileInTar)
	}

	files := tarEntries(t, r)
	if !files["app.go"] {
		t.Fatal("expected app.go in tar archive")
	}
	if !files[".sew.Dockerfile"] {
		t.Fatal("expected .sew.Dockerfile injected into tar archive")
	}
}

func TestCreateBuildContext_SubdirectoryDockerfile(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "docker")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "Dockerfile"), []byte("FROM alpine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r, dockerfileInTar, err := createBuildContext(dir, filepath.Join(subdir, "Dockerfile"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dockerfileInTar != "docker/Dockerfile" {
		t.Fatalf("expected dockerfileInTar %q, got %q", "docker/Dockerfile", dockerfileInTar)
	}

	files := tarEntries(t, r)
	if !files["docker/Dockerfile"] {
		t.Fatal("expected docker/Dockerfile in tar archive")
	}
	if !files["main.go"] {
		t.Fatal("expected main.go in tar archive")
	}
}

func tarEntries(t *testing.T, r io.Reader) map[string]bool {
	t.Helper()
	tr := tar.NewReader(r)
	files := make(map[string]bool)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading tar: %v", err)
		}
		files[hdr.Name] = true
	}
	return files
}
