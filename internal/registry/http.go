package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/core"
	"gopkg.in/yaml.v3"
)

// HTTPResolver resolves contexts by fetching files from an HTTP registry.
type HTTPResolver struct {
	BaseURL    string
	CacheRoot  string
	HTTPClient *http.Client
}

// Resolve fetches {BaseURL}/{contextPath}/context.yaml, downloads referenced
// values files to a local cache directory, and returns a ResolvedContext
// whose Dir points to that cache.
func (r *HTTPResolver) Resolve(ctx context.Context, contextPath string) (*core.ResolvedContext, error) {
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
	if resp.StatusCode == http.StatusNotFound {
		variant, defaultErr := r.fetchDefault(ctx, client, baseURL, contextPath)
		if defaultErr != nil {
			return nil, fmt.Errorf("fetching context: %s", resp.Status)
		}
		return r.Resolve(ctx, contextPath+"/"+variant)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching context: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading context: %w", err)
	}

	var parsed core.Context
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parsing context.yaml: %w", err)
	}

	cacheDir := filepath.Join(r.CacheRoot, contextPath)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	var filesToFetch []string
	seen := make(map[string]bool)
	for _, comp := range parsed.Components {
		if comp.Helm != nil {
			for _, v := range comp.Helm.ValueFiles {
				if !seen[v] {
					seen[v] = true
					filesToFetch = append(filesToFetch, v)
				}
			}
		}
		if comp.K8s != nil {
			for _, f := range comp.K8s.ManifestFiles {
				if !seen[f] {
					seen[f] = true
					filesToFetch = append(filesToFetch, f)
				}
			}
		}
	}

	for _, f := range filesToFetch {
		u := baseURL + "/" + contextPath + "/" + f
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, fmt.Errorf("building request for %s: %w", f, err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetching %s: %w", f, err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("fetching %s: %s", f, resp.Status)
		}
		outPath := filepath.Join(cacheDir, f)
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
			return nil, fmt.Errorf("writing %s: %w", f, err)
		}
	}

	return &core.ResolvedContext{
		Repos:      parsed.Repos,
		Components: parsed.Components,
		Dir:        cacheDir,
		Kind:       parsed.Kind,
		Features:   parsed.Features,
	}, nil
}

// fetchDefault fetches {baseURL}/{contextPath}/.default and returns the
// trimmed variant name. It returns an error if the file does not exist,
// the server returns a non-200 status, or the file content is empty.
func (r *HTTPResolver) fetchDefault(ctx context.Context, client *http.Client, baseURL, contextPath string) (string, error) {
	u := baseURL + "/" + contextPath + "/.default"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("building request for .default: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching .default: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching .default: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading .default: %w", err)
	}
	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", fmt.Errorf("empty .default file at %s", u)
	}
	return name, nil
}
