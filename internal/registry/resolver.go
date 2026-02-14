package registry

import (
	"context"
	"strings"
)

// Resolver fetches a context from a registry and returns it
// with all referenced files available locally.
type Resolver interface {
	Resolve(ctx context.Context, contextPath string) (*ResolvedContext, error)
}

// NewResolver builds the appropriate Resolver from the registry URL.
// A "file://" prefix selects the filesystem resolver; anything else
// is treated as an HTTP registry.
func NewResolver(registry string) Resolver {
	if strings.HasPrefix(registry, "file://") {
		return &FSResolver{Root: strings.TrimPrefix(registry, "file://")}
	}
	return &HTTPResolver{BaseURL: registry}
}
