package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-cordier/sew/api"
	"github.com/a-cordier/sew/internal/installer"
	"github.com/a-cordier/sew/internal/kind"
	"github.com/a-cordier/sew/internal/registry"
	sewlog "github.com/a-cordier/sew/internal/log"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
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

	logFile, err := os.OpenFile(filepath.Join(home, "sew.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer logFile.Close()
	klog.SetOutput(logFile)

	var resolved *api.ResolvedContext
	if cfg.Registry != "" && cfg.Context != "" {
		registryURL := cfg.Registry
		if strings.HasPrefix(registryURL, "file://") {
			path := strings.TrimPrefix(registryURL, "file://")
			if abs, err := filepath.Abs(path); err == nil {
				registryURL = "file://" + abs
			}
		}
		resolver := registry.NewResolver(registryURL)
		var resolveErr error
		resolved, resolveErr = resolver.Resolve(context.Background(), cfg.Context)
		if resolveErr != nil {
			return fmt.Errorf("resolving context %q: %w", cfg.Context, resolveErr)
		}
		cfg.Kind.MergeWithContext(resolved.Kind)
	}

	if err := sewlog.WithSpinner(
		fmt.Sprintf("Creating cluster %q", cfg.Kind.Name),
		func() error { return kind.Create(cfg.Kind.Name, cfg.Kind.RawYAML()) },
	); err != nil {
		return err
	}

	if cfg.Registry == "" || cfg.Context == "" {
		return nil
	}

	registry.ApplyOverrides(resolved, cfg.Overrides, cfg.Dir)

	if err := registry.Validate(resolved.Components); err != nil {
		return fmt.Errorf("validating components: %w", err)
	}
	sorted, err := registry.TopoSort(resolved.Components)
	if err != nil {
		return fmt.Errorf("resolving dependencies: %w", err)
	}


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

	compByName := make(map[string]api.Component)
	for _, c := range resolved.Components {
		compByName[c.Name] = c
	}

	const defaultReadyTimeout = 5 * time.Minute
	ctx := context.Background()
	for _, comp := range sorted {
		for _, req := range comp.Requires {
			if req.Conditions.Ready {
				dep := compByName[req.Component]
				depNamespace := dep.Namespace
				if depNamespace == "" {
					depNamespace = "default"
				}
				timeout := defaultReadyTimeout
				if req.Timeout != "" {
					if d, err := time.ParseDuration(req.Timeout); err == nil && d > 0 {
						timeout = d
					}
				}
				var matchLabels map[string]string
				if req.Selector != nil && len(req.Selector.MatchLabels) > 0 {
					matchLabels = req.Selector.MatchLabels
				}
				if err := sewlog.WithSpinner(fmt.Sprintf("Waiting for %q to be ready", req.Component), func() error {
					return installer.WaitForReady(ctx, req.Component, depNamespace, timeout, matchLabels)
				}); err != nil {
					return fmt.Errorf("requirement %q not ready: %w", req.Component, err)
				}
			}
		}
		inst, err := installer.ForType(comp.EffectiveType())
		if err != nil {
			return fmt.Errorf("component %q: %w", comp.Name, err)
		}
		comp := comp
		if err := sewlog.WithSpinner(fmt.Sprintf("Installing %q", comp.Name), func() error {
			return inst.Install(ctx, comp, resolved.Dir)
		}); err != nil {
			return err
		}
	}

	return nil
}
