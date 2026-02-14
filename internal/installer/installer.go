package installer

import (
	"context"
	"fmt"

	"github.com/a-cordier/sew/internal/registry"
)

// Installer knows how to install and uninstall a component.
type Installer interface {
	Install(ctx context.Context, comp registry.Component, dir string) error
	Uninstall(ctx context.Context, comp registry.Component) error
}

// registry of installers, keyed by component type
var installers = map[string]Installer{
	"helm":     &HelmInstaller{},
	"manifest": &ManifestInstaller{},
}

// ForType returns the Installer for the given component type.
func ForType(componentType string) (Installer, error) {
	inst, ok := installers[componentType]
	if !ok {
		return nil, fmt.Errorf("unknown component type: %q", componentType)
	}
	return inst, nil
}
