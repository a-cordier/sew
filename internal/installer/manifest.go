package installer

import (
	"context"
	"fmt"

	"github.com/a-cordier/sew/internal/registry"
)

// ManifestInstaller installs components from plain Kubernetes manifest files.
// Not yet implemented.
type ManifestInstaller struct{}

// Install is not implemented yet.
func (m *ManifestInstaller) Install(_ context.Context, comp registry.Component, _ string) error {
	return fmt.Errorf("manifest installer is not implemented yet (component %q)", comp.Name)
}

// Uninstall is not implemented yet.
func (m *ManifestInstaller) Uninstall(_ context.Context, comp registry.Component) error {
	return fmt.Errorf("manifest installer is not implemented yet (component %q)", comp.Name)
}
