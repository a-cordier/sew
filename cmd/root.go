package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/core"
	"github.com/a-cordier/sew/internal/config"
	"github.com/a-cordier/sew/internal/registry"
	"github.com/spf13/cobra"
)

// DefaultConfigData holds the embedded sew.yaml from the project root,
// set by main before Execute().
var DefaultConfigData []byte

var (
	cfgFile     string
	contextPath string
	cfg         *core.Config
	sewHome     string
)

var rootCmd = &cobra.Command{
	Use:   "sew",
	Short: "Spin up local Kubernetes clusters and deploy ready-to-use applications",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		sewHome = os.Getenv("SEW_HOME")
		if sewHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("determining user home directory: %w", err)
			}
			sewHome = filepath.Join(home, ".sew")
		}
		var err error
		cfg, err = resolveConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if contextPath != "" {
			cfg.Context = contextPath
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file (default: ./sew.yaml or ~/.sew/sew.yaml)")
	rootCmd.PersistentFlags().StringVar(&contextPath, "context", "", "context path to use (overrides config file)")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// resolveConfig loads the configuration using layered merging:
//  1. Load $sewHome/sew.yaml as the base config (if it exists).
//  2. If --config is given, load and merge on top; otherwise if ./sew.yaml
//     exists, load and merge on top.
//  3. Apply embedded defaults to fill any remaining gaps.
func resolveConfig(explicit string) (*core.Config, error) {
	basePath := filepath.Join(sewHome, "sew.yaml")
	var base *core.Config
	if fileExists(basePath) {
		var err error
		base, err = config.Load(basePath)
		if err != nil {
			return nil, fmt.Errorf("loading base config %s: %w", basePath, err)
		}
	} else {
		base = &core.Config{}
		base.Kind.ApplyDefaults()
	}

	var projectCfg *core.Config
	switch {
	case explicit != "":
		var err error
		projectCfg, err = config.Load(explicit)
		if err != nil {
			return nil, err
		}
	case fileExists("sew.yaml"):
		var err error
		projectCfg, err = config.Load("sew.yaml")
		if err != nil {
			return nil, err
		}
	}

	if projectCfg != nil {
		config.Merge(base, projectCfg)
	}

	config.ApplyDefaults(base, DefaultConfigData)
	return base, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// resolveContextConfig resolves the registry context (if configured) and
// merges its Kind and Features settings into the global cfg. The returned
// ResolvedContext is nil when no registry/context is set.
func resolveContextConfig() (*core.ResolvedContext, error) {
	if cfg.Registry == "" || cfg.Context == "" {
		return nil, nil
	}
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
		return nil, fmt.Errorf("resolving context %q: %w", cfg.Context, err)
	}
	cfg.Kind.MergeWithContext(resolved.Kind)
	cfg.Features = core.MergeFeatures(resolved.Features, cfg.Features)
	return resolved, nil
}
