package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/cache"
	"github.com/a-cordier/sew/internal/kind"
	"github.com/a-cordier/sew/internal/logger"
	"github.com/a-cordier/sew/internal/registry"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
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

		logDir := filepath.Join(sewHome, "logs")
		if cfg.Context != "" {
			logDir = filepath.Join(logDir, cfg.Context)
		}
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return fmt.Errorf("creating log directory %s: %w", logDir, err)
		}
		logPath := filepath.Join(logDir, "delete.log")
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return fmt.Errorf("opening log file: %w", err)
		}
		defer logFile.Close()
		klog.LogToStderr(false)
		klog.SetOutput(logFile)
		logger.SetLogFile(logPath)

		if err := logger.WithSpinner(
			fmt.Sprintf("Deleting cluster %q", cfg.Kind.Name),
			func() error {
				return kind.Delete(cfg.Kind.Name)
			},
		); err != nil {
			return err
		}

		if cfg.Images.Mirrors != nil {
			ctx := context.Background()
			if err := logger.WithSpinner("Stopping image mirror proxies", func() error {
				return cache.StopProxies(ctx, cfg.Images.Mirrors)
			}); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
