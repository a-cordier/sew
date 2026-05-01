package registry

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/config"
)

// Resolver resolves a context path against a registry into a ResolvedContext.
type Resolver interface {
	Resolve(ctx context.Context, contextPath string) (*config.ResolvedContext, error)
}

// NewResolver builds the appropriate Resolver from the registry URL.
// A "file://" prefix selects the filesystem resolver; anything else
// is treated as an HTTP registry with cache rooted under sewHome.
// setOverrides are the --set key=value pairs forwarded to the template
// engine when rendering context sew.yaml files.
func NewResolver(registry string, sewHome string, setOverrides map[string]string) Resolver {
	if strings.HasPrefix(registry, "file://") {
		return &FSResolver{
			Root:         strings.TrimPrefix(registry, "file://"),
			SewHome:      sewHome,
			SetOverrides: setOverrides,
		}
	}
	return &HTTPResolver{
		BaseURL:      registry,
		CacheRoot:    filepath.Join(sewHome, "cache"),
		SewHome:      sewHome,
		HTTPClient:   newAuthenticatedClient(registry),
		SetOverrides: setOverrides,
	}
}
