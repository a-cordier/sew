package cache

import "testing"

func TestStripRegistryHost(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"graviteeio/apim-gateway:latest", "graviteeio/apim-gateway:latest"},
		{"mongo:7", "library/mongo:7"},
		{"mongo", "library/mongo"},
		{"ghcr.io/org/repo:v1.2", "org/repo:v1.2"},
		{"docker.io/library/nginx:1.25", "library/nginx:1.25"},
		{"my-registry.example.com:5000/team/app:sha-abc", "team/app:sha-abc"},
		{"localhost:5000/myimg:dev", "myimg:dev"},
	}
	for _, tc := range tests {
		t.Run(tc.ref, func(t *testing.T) {
			got := stripRegistryHost(tc.ref)
			if got != tc.want {
				t.Fatalf("stripRegistryHost(%q) = %q, want %q", tc.ref, got, tc.want)
			}
		})
	}
}

func TestRegistryHost(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"graviteeio/apim-gateway:latest", "docker.io"},
		{"mongo:7", "docker.io"},
		{"ghcr.io/org/repo:v1.2", "ghcr.io"},
		{"my-registry.example.com:5000/team/app:sha-abc", "my-registry.example.com:5000"},
		{"localhost:5000/myimg:dev", "localhost:5000"},
	}
	for _, tc := range tests {
		t.Run(tc.ref, func(t *testing.T) {
			got := registryHost(tc.ref)
			if got != tc.want {
				t.Fatalf("registryHost(%q) = %q, want %q", tc.ref, got, tc.want)
			}
		})
	}
}

func TestPreloadUpstreams(t *testing.T) {
	images := []string{
		"graviteeio/apim-gateway:latest",
		"mongo:7",
		"ghcr.io/org/repo:v1.2",
		"graviteeio/apim-management-api:latest",
	}
	upstreams := PreloadUpstreams(images)

	want := []string{"docker.io", "ghcr.io"}
	if len(upstreams) != len(want) {
		t.Fatalf("expected %d upstreams, got %d: %v", len(want), len(upstreams), upstreams)
	}
	for i, w := range want {
		if upstreams[i] != w {
			t.Fatalf("upstreams[%d] = %q, want %q", i, upstreams[i], w)
		}
	}
}
