package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/installer"
	"github.com/a-cordier/sew/internal/kind"
	"github.com/a-cordier/sew/internal/registry"
	sewlog "github.com/a-cordier/sew/internal/log"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create the cluster and install the context",
	RunE:  runUp,
}

func init() {
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, args []string) error {
	// Resolve home to absolute path
	home := cfg.Home
	if !filepath.IsAbs(home) {
		cwd, _ := os.Getwd()
		home = filepath.Join(cwd, home)
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return fmt.Errorf("failed to create home directory %s: %w", home, err)
	}

	// Create Kind cluster
	if err := sewlog.WithSpinner(
		fmt.Sprintf("Creating cluster %q", cfg.Kind.Name),
		func() error { return kind.Create(cfg.Kind.Name, cfg.Kind.RawYAML()) },
	); err != nil {
		return err
	}

	// If no registry/context configured, stop after cluster creation
	if cfg.Registry == "" || cfg.Context == "" {
		return nil
	}

	// Resolve file:// registry path to absolute so it works from any cwd
	registryURL := cfg.Registry
	if strings.HasPrefix(registryURL, "file://") {
		path := strings.TrimPrefix(registryURL, "file://")
		if abs, err := filepath.Abs(path); err == nil {
			registryURL = "file://" + abs
		}
	}

	resolver := registry.NewResolver(registryURL)
	resolved, err := resolver.Resolve(context.Background(), cfg.Context)
	if err != nil {
		return fmt.Errorf("resolving context %q: %w", cfg.Context, err)
	}

	registry.ApplyOverrides(resolved, cfg.Overrides, cfg.Dir)

	// Add Helm repos once before installing components
	if len(resolved.Repos) > 0 {
		helmInst, _ := installer.ForType("helm")
		if hi, ok := helmInst.(*installer.HelmInstaller); ok {
			if err := sewlog.WithSpinner("Adding Helm repositories", func() error {
				return hi.AddRepos(resolved.Repos, home)
			}); err != nil {
				return err
			}
		}
	}

	for _, comp := range resolved.Components {
		inst, err := installer.ForType(comp.EffectiveType())
		if err != nil {
			return fmt.Errorf("component %q: %w", comp.Name, err)
		}
		comp := comp
		if err := sewlog.WithSpinner(fmt.Sprintf("Installing %q", comp.Name), func() error {
			return inst.Install(context.Background(), comp, resolved.Dir)
		}); err != nil {
			return err
		}
	}

	return nil
}
