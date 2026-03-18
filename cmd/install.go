package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/a-cordier/sew/internal/config"
	"github.com/a-cordier/sew/internal/installer"
	"github.com/a-cordier/sew/internal/logger"
	"github.com/a-cordier/sew/internal/registry"
)

const defaultReadyTimeout = 5 * time.Minute

// installComponents validates, topo-sorts, initialises Helm repos and installs
// the components from the resolved context. When filter is non-nil only
// components for which filter returns true are installed; dependency readiness
// checks still consider the full component set.
func installComponents(
	ctx context.Context,
	resolved *config.ResolvedContext,
	filter func(config.Component) bool,
	opts installer.InstallOpts,
) error {
	if err := registry.Validate(resolved.Components); err != nil {
		return fmt.Errorf("validating components: %w", err)
	}
	sorted, err := registry.TopoSort(resolved.Components)
	if err != nil {
		return fmt.Errorf("resolving dependencies: %w", err)
	}

	helmInst, _ := installer.ForType("helm")
	if hi, ok := helmInst.(*installer.HelmInstaller); ok {
		if err := logger.WithSpinner("Initializing Helm", func() error {
			return hi.AddRepos(resolved.Repos, sewHome)
		}); err != nil {
			return err
		}
	}

	compByName := make(map[string]config.Component)
	for _, c := range resolved.Components {
		compByName[c.Name] = c
	}

	for _, comp := range sorted {
		for _, req := range comp.Requires {
			if req.Conditions.Ready && !opts.DryRun {
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

		if filter != nil && !filter(comp) {
			continue
		}

		inst, err := installer.ForType(comp.EffectiveType())
		if err != nil {
			return fmt.Errorf("component %q: %w", comp.Name, err)
		}
		comp := comp
		if err := logger.WithSpinner(fmt.Sprintf("Installing %q", comp.Name), func() error {
			return inst.Install(ctx, comp, resolved.Dir, opts)
		}); err != nil {
			return err
		}
		if comp.Conditions.Ready && !opts.DryRun {
			ns := comp.Namespace
			if ns == "" {
				ns = "default"
			}
			timeout := defaultReadyTimeout
			if comp.Timeout != "" {
				if d, err := time.ParseDuration(comp.Timeout); err == nil && d > 0 {
					timeout = d
				}
			}
			var matchLabels map[string]string
			if comp.Selector != nil && len(comp.Selector.MatchLabels) > 0 {
				matchLabels = comp.Selector.MatchLabels
			}
			if err := logger.WithSpinner(
				fmt.Sprintf("Waiting for %q to be ready", comp.Name),
				func() error {
					return installer.WaitForReady(ctx, comp.Name, ns, timeout, matchLabels)
				},
			); err != nil {
				return fmt.Errorf("component %q not ready: %w", comp.Name, err)
			}
		}
	}

	return nil
}
