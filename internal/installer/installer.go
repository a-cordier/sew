package installer

import (
	"fmt"

	"github.com/a-cordier/sew/api"
)

var installers = map[string]api.Installer{
	"helm":     &HelmInstaller{},
	"manifest": &ManifestInstaller{},
}

func ForType(componentType string) (api.Installer, error) {
	inst, ok := installers[componentType]
	if !ok {
		return nil, fmt.Errorf("unknown component type: %q", componentType)
	}
	return inst, nil
}
