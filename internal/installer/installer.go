package installer

import (
	"fmt"

	"github.com/a-cordier/sew/core"
)

var installers = map[string]core.Installer{
	"helm":     &HelmInstaller{},
	"manifest": &ManifestInstaller{},
}

func ForType(componentType string) (core.Installer, error) {
	inst, ok := installers[componentType]
	if !ok {
		return nil, fmt.Errorf("unknown component type: %q", componentType)
	}
	return inst, nil
}
