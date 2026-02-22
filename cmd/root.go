package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a-cordier/sew/core"
	"github.com/a-cordier/sew/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	contextPath string
	cfg         *core.Config
)

var rootCmd = &cobra.Command{
	Use:   "sew",
	Short: "Spin up local Kubernetes clusters and deploy ready-to-use applications",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
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

// resolveConfig loads the configuration from the first available path:
// 1. Explicit --config flag
// 2. ./sew.yaml in the current directory
// 3. ~/.sew/sew.yaml in the user's home directory
func resolveConfig(explicit string) (*core.Config, error) {
	if explicit != "" {
		return config.Load(explicit)
	}

	if _, err := os.Stat("sew.yaml"); err == nil {
		return config.Load("sew.yaml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	defaultPath := filepath.Join(home, ".sew", "sew.yaml")
	if _, err := os.Stat(defaultPath); err == nil {
		return config.Load(defaultPath)
	}

	return nil, fmt.Errorf("no config file found (tried ./sew.yaml and %s)", defaultPath)
}
