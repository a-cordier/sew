package core

import (
	"context"
)

// Installer deploys and removes components.
type Installer interface {
	Install(ctx context.Context, comp Component, dir string) error
	Uninstall(ctx context.Context, comp Component) error
}
