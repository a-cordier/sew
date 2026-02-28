package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-cordier/sew/core"
	"github.com/a-cordier/sew/internal/cache"
	"github.com/a-cordier/sew/internal/installer"
	"github.com/a-cordier/sew/internal/kind"
	"github.com/a-cordier/sew/internal/logger"
	"github.com/a-cordier/sew/internal/registry"
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

func runUp(_ *cobra.Command, _ []string) error {
	if err := os.MkdirAll(sewHome, 0o755); err != nil {
		return fmt.Errorf("failed to create home directory %s: %w", sewHome, err)
	}

	var resolved *core.ResolvedContext
	if cfg.Registry != "" && cfg.Context != "" {
		registryURL := cfg.Registry
		if strings.HasPrefix(registryURL, "file://") {
			path := strings.TrimPrefix(registryURL, "file://")
			if abs, err := filepath.Abs(path); err == nil {
				registryURL = "file://" + abs
			}
		}
		resolver := registry.NewResolver(registryURL, sewHome)
		var resolveErr error
		resolved, resolveErr = resolver.Resolve(context.Background(), cfg.Context)
		if resolveErr != nil {
			return fmt.Errorf("resolving context %q: %w", cfg.Context, resolveErr)
		}
		cfg.Kind.MergeWithContext(resolved.Kind)
	}

	logDir := filepath.Join(sewHome, "logs")
	if cfg.Context != "" {
		logDir = filepath.Join(logDir, cfg.Context)
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("creating log directory %s: %w", logDir, err)
	}
	logFile, err := os.OpenFile(filepath.Join(logDir, "install.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer logFile.Close()
	klog.LogToStderr(false)
	klog.SetOutput(logFile)

	ctx := context.Background()

	if cfg.Images.Mirrors != nil {
		if err := logger.WithSpinner("Starting image mirror proxies", func() error {
			return cache.EnsureProxies(ctx, cfg.Images.Mirrors, sewHome)
		}); err != nil {
			return err
		}
		mirrors, err := cache.PrepareMirrors(cfg.Images.Mirrors, sewHome)
		if err != nil {
			return fmt.Errorf("preparing image mirrors: %w", err)
		}
		cfg.Kind.ContainerdConfigPatches = append(cfg.Kind.ContainerdConfigPatches, mirrors.Patch)
		for i := range cfg.Kind.Nodes {
			cfg.Kind.Nodes[i].ExtraMounts = append(cfg.Kind.Nodes[i].ExtraMounts, mirrors.Mounts...)
		}
	}

	kindConfig, err := cfg.Kind.RawYAML()
	if err != nil {
		return fmt.Errorf("serializing kind config: %w", err)
	}
	if err := logger.WithSpinner(
		fmt.Sprintf("Creating cluster %q", cfg.Kind.Name),
		func() error { return kind.Create(cfg.Kind.Name, kindConfig) },
	); err != nil {
		return err
	}

	if cfg.Images.Mirrors != nil {
		if err := logger.WithSpinner("Connecting image mirrors to Kind network", func() error {
			return cache.ConnectToKindNetwork(ctx, cfg.Kind.Name, cfg.Images.Mirrors)
		}); err != nil {
			return err
		}
	}

	if cfg.Registry == "" || cfg.Context == "" {
		return nil
	}

	registry.MergeComponents(resolved, cfg.Components, cfg.Dir)

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
			if err := logger.WithSpinner("Adding Helm repositories", func() error {
				return hi.AddRepos(resolved.Repos, sewHome)
			}); err != nil {
				return err
			}
		}
	}

	compByName := make(map[string]core.Component)
	for _, c := range resolved.Components {
		compByName[c.Name] = c
	}

	const defaultReadyTimeout = 5 * time.Minute
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
				if err := logger.WithSpinner(fmt.Sprintf("Waiting for %q to be ready", req.Component), func() error {
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
		if err := logger.WithSpinner(fmt.Sprintf("Installing %q", comp.Name), func() error {
			return inst.Install(ctx, comp, resolved.Dir)
		}); err != nil {
			return err
		}
	}

	return nil
}
