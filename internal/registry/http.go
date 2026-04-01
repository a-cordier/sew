package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/config"
	"gopkg.in/yaml.v3"
)

// HTTPResolver resolves contexts by fetching files from an HTTP registry.
type HTTPResolver struct {
	BaseURL    string
	CacheRoot  string
	SewHome    string
	HTTPClient *http.Client
}

// Resolve fetches {BaseURL}/{contextPath}/sew.yaml, downloads referenced
// values files to a local cache directory, and returns a ResolvedContext
// whose Dir points to that cache.
//
// If sew.yaml returns 404, Resolve tries the .default variant lookup.
func (r *HTTPResolver) Resolve(ctx context.Context, contextPath string) (*config.ResolvedContext, error) {
	ctx, err := withVisited(ctx, contextRef{Registry: r.BaseURL, Context: contextPath})
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimSuffix(r.BaseURL, "/")
	client := r.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	data, status, err := r.fetchContextFile(ctx, client, baseURL, contextPath)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		variant, defaultErr := r.fetchDefault(ctx, client, baseURL, contextPath)
		if defaultErr != nil {
			return nil, fmt.Errorf("fetching context: 404 Not Found")
		}
		return r.Resolve(ctx, contextPath+"/"+variant)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fetching context: %d", status)
	}

	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parsing context file: %w", err)
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
		if err := r.fetchAndCache(ctx, client, baseURL, contextPath, cacheDir, f, false); err != nil {
			return nil, err
		}
	}

	for _, notesFile := range []string{"notes.create", "notes.delete"} {
		if err := r.fetchAndCache(ctx, client, baseURL, contextPath, cacheDir, notesFile, true); err != nil {
			return nil, err
		}
	}

	flags, err := r.fetchFlags(ctx, client, baseURL, contextPath, cacheDir)
	if err != nil {
		return nil, err
	}

	if len(parsed.From) > 0 {
		return resolveFrom(ctx, parsed, cacheDir, r.BaseURL, r.SewHome)
	}

	return &config.ResolvedContext{
		Repos:      parsed.Helm.Repos,
		Components: parsed.Components,
		Dir:        cacheDir,
		Kind:       parsed.Kind,
		Features:   parsed.Features,
		Images:     parsed.Images,
		Notes:      readNotes(cacheDir),
		Abstract:   parsed.Abstract,
		Flags:      flags,
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

// fetchAndCache fetches a single file from the remote registry and writes it
// to the local cache directory. When ignoreNotFound is true, a 404 response
// is silently skipped instead of treated as an error.
func (r *HTTPResolver) fetchAndCache(ctx context.Context, client *http.Client, baseURL, contextPath, cacheDir, filename string, ignoreNotFound bool) error {
	u := baseURL + "/" + contextPath + "/" + filename
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("building request for %s: %w", filename, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", filename, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound && ignoreNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching %s: %s", filename, resp.Status)
	}
	outPath := filepath.Join(cacheDir, filename)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return fmt.Errorf("writing %s: %w", filename, err)
	}
	return nil
}

type flagsManifest struct {
	Flags []flagEntry `yaml:"flags"`
}

type flagEntry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// fetchFlags fetches the sew.flags.yaml manifest from the remote registry,
// downloads each referenced flag patch file to cacheDir, and returns the
// corresponding ContextFlag entries. A 404 on the manifest means no flags.
func (r *HTTPResolver) fetchFlags(ctx context.Context, client *http.Client, baseURL, contextPath, cacheDir string) ([]config.ContextFlag, error) {
	u := baseURL + "/" + contextPath + "/sew.flags.yaml"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for flags manifest: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching flags manifest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching flags manifest: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading flags manifest: %w", err)
	}

	var manifest flagsManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing flags manifest: %w", err)
	}

	var flags []config.ContextFlag
	for _, entry := range manifest.Flags {
		flagFile := flagFilePrefix + entry.Name + ".yaml"
		if err := r.fetchAndCache(ctx, client, baseURL, contextPath, cacheDir, flagFile, true); err != nil {
			return nil, err
		}
		flags = append(flags, config.ContextFlag{
			Name:        entry.Name,
			Description: entry.Description,
			Dir:         cacheDir,
		})
	}
	return flags, nil
}

// fetchContextFile fetches sew.yaml from the registry. It returns the body
// bytes, the HTTP status code, and any transport-level error. A 404 is
// returned as status (not as an error) so the caller can attempt the
// .default fallback.
func (r *HTTPResolver) fetchContextFile(ctx context.Context, client *http.Client, baseURL, contextPath string) ([]byte, int, error) {
	u := baseURL + "/" + contextPath + "/sew.yaml"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("building request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching context: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("reading context: %w", err)
	}
	return data, http.StatusOK, nil
}



