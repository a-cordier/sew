package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/config"
	"github.com/a-cordier/sew/internal/registry"
	"github.com/spf13/cobra"
)

// DefaultConfigData holds the embedded sew.yaml from the project root,
// set by main before Execute().
var DefaultConfigData []byte

var (
	cfgFile     string
	registryURL string
	fromPaths   []string
	cfg         *config.Config
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
		if registryURL != "" {
			cfg.Registry = registryURL
		}
		if len(fromPaths) > 0 {
			cfg.From = fromPaths
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file (default: ./sew.yaml or ~/.sew/sew.yaml)")
	rootCmd.PersistentFlags().StringVar(&registryURL, "registry", "", "registry URL to use (overrides config file)")
	rootCmd.PersistentFlags().StringSliceVar(&fromPaths, "from", nil, "context paths to compose (repeatable, overrides config file)")
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
func resolveConfig(explicit string) (*config.Config, error) {
	basePath := filepath.Join(sewHome, "sew.yaml")
	var base *config.Config
	if fileExists(basePath) {
		var err error
		base, err = config.Load(basePath)
		if err != nil {
			return nil, fmt.Errorf("loading base config %s: %w", basePath, err)
		}
	} else {
		base = &config.Config{}
		base.Kind.ApplyDefaults()
	}

	var projectCfg *config.Config
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
func resolveContextConfig() (*config.ResolvedContext, error) {
	if cfg.Registry == "" || len(cfg.From) == 0 {
		return nil, nil
	}
	regURL := cfg.Registry
	if strings.HasPrefix(regURL, "file://") {
		path := strings.TrimPrefix(regURL, "file://")
		if abs, err := filepath.Abs(path); err == nil {
			regURL = "file://" + abs
		}
	}
	resolver := registry.NewResolver(regURL, sewHome)
	resolved, err := resolver.Resolve(context.Background(), cfg.From[0])
	if err != nil {
		return nil, fmt.Errorf("resolving context %q: %w", cfg.From[0], err)
	}
	if resolved.Abstract {
		return nil, fmt.Errorf("context %q is abstract and cannot be deployed directly; compose it via 'from' in another context", cfg.From[0])
	}
	cfg.Kind.MergeWithContext(&resolved.Kind)
	cfg.Features = config.MergeFeatures(resolved.Features, cfg.Features)
	cfg.Images = config.MergeImages(resolved.Images, cfg.Images)
	return resolved, nil
}
