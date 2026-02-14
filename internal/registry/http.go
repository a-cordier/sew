package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// HTTPResolver resolves contexts by fetching files from an HTTP registry.
type HTTPResolver struct {
	BaseURL    string
	CacheRoot  string // default ~/.sew/cache
	HTTPClient *http.Client
}

// Resolve fetches {BaseURL}/{contextPath}/context.yaml, downloads referenced
// values files to a local cache directory, and returns a ResolvedContext
// whose Dir points to that cache.
func (r *HTTPResolver) Resolve(ctx context.Context, contextPath string) (*ResolvedContext, error) {
	baseURL := strings.TrimSuffix(r.BaseURL, "/")
	contextURL := baseURL + "/" + contextPath + "/context.yaml"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, contextURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	client := r.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching context: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching context: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading context: %w", err)
	}

	var parsed Context
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parsing context.yaml: %w", err)
	}

	cacheRoot := r.CacheRoot
	if cacheRoot == "" {
		home, _ := os.UserHomeDir()
		cacheRoot = filepath.Join(home, ".sew", "cache")
	}
	cacheDir := filepath.Join(cacheRoot, contextPath)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	// Download each referenced values file
	seen := make(map[string]bool)
	for _, comp := range parsed.Components {
		if comp.Helm == nil {
			continue
		}
		for _, v := range comp.Helm.Values {
			if seen[v] {
				continue
			}
			seen[v] = true
			u := baseURL + "/" + contextPath + "/" + v
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				return nil, fmt.Errorf("building request for %s: %w", v, err)
			}
			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("fetching %s: %w", v, err)
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, fmt.Errorf("fetching %s: %s", v, resp.Status)
			}
			outPath := filepath.Join(cacheDir, v)
			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
				resp.Body.Close()
				return nil, err
			}
			out, err := os.Create(outPath)
			if err != nil {
				resp.Body.Close()
				return nil, err
			}
			_, err = io.Copy(out, resp.Body)
			out.Close()
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("writing %s: %w", v, err)
			}
		}
	}

	return &ResolvedContext{
		Repos:      parsed.Repos,
		Components: parsed.Components,
		Dir:        cacheDir,
	}, nil
}
