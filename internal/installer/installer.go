package installer

import (
	"context"
	"fmt"

	"github.com/a-cordier/sew/core"
)

// Installer deploys and removes components.
type Installer interface {
	Install(ctx context.Context, comp core.Component, dir string) error
	Uninstall(ctx context.Context, comp core.Component) error
}

var installers = map[string]Installer{
	"helm": &HelmInstaller{},
	"k8s":  &ManifestInstaller{},
}

func ForType(componentType string) (Installer, error) {
	inst, ok := installers[componentType]
	if !ok {
		return nil, fmt.Errorf("unknown component type: %q", componentType)
	}
	return inst, nil
}
