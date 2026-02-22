package registry

import (
	"strings"

	"github.com/a-cordier/sew/core"
)

// NewResolver builds the appropriate Resolver from the registry URL.
// A "file://" prefix selects the filesystem resolver; anything else
// is treated as an HTTP registry.
func NewResolver(registry string) core.Resolver {
	if strings.HasPrefix(registry, "file://") {
		return &FSResolver{Root: strings.TrimPrefix(registry, "file://")}
	}
	return &HTTPResolver{BaseURL: registry}
}
