package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/kind"
	"github.com/a-cordier/sew/internal/logger"
	"github.com/a-cordier/sew/internal/registry"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Delete the cluster defined in the config",
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfg.Registry != "" && cfg.Context != "" {
			registryURL := cfg.Registry
			if strings.HasPrefix(registryURL, "file://") {
				path := strings.TrimPrefix(registryURL, "file://")
				if abs, err := filepath.Abs(path); err == nil {
					registryURL = "file://" + abs
				}
			}
			resolver := registry.NewResolver(registryURL, sewHome)
			resolved, err := resolver.Resolve(context.Background(), cfg.Context)
			if err != nil {
				return fmt.Errorf("resolving context %q: %w", cfg.Context, err)
			}
			cfg.Kind.MergeWithContext(resolved.Kind)
		}

		return logger.WithSpinner(
			fmt.Sprintf("Deleting cluster %q", cfg.Kind.Name),
			func() error {
				return kind.Delete(cfg.Kind.Name)
			},
		)
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
