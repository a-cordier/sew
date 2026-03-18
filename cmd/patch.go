package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-cordier/sew/internal/config"
	"github.com/a-cordier/sew/internal/kind"
	"github.com/a-cordier/sew/internal/registry"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var patchClusterName string

var patchCmd = &cobra.Command{
	Use:   "patch <patch-file>",
	Short: "Patch a running cluster by upgrading components with overrides",
	Long: `Patch merges a partial configuration file into the current resolved context
and upgrades only the affected components on a running Kind cluster.

The patch file uses the same format as sew.yaml. Only the "components" and
"helm.repos" sections are relevant; other fields are ignored.

Example:

  sew patch upgrade.yaml

This resolves the current context (from sew.yaml / --from), merges the patch
on top, and runs helm upgrade / kubectl apply for each component listed in the
patch file.`,
	Args: cobra.ExactArgs(1),
	RunE: runPatch,
}

func init() {
	patchCmd.Flags().StringVar(&patchClusterName, "name", "", "name of the cluster to patch (default: from config)")
	rootCmd.AddCommand(patchCmd)
}

func runPatch(_ *cobra.Command, args []string) error {
	start := time.Now()

	resolved, err := resolveContextConfig()
	if err != nil {
		return err
	}
	if resolved == nil {
		return fmt.Errorf("no registry context configured; patch requires a resolved context (set registry and from in sew.yaml or via flags)")
	}

	clusterName := patchClusterName
	if clusterName == "" {
		clusterName = cfg.Kind.Name
	}
	if clusterName == "" {
		return fmt.Errorf("cannot determine cluster name; use --name or set kind.name in config")
	}

	exists, err := kind.Exists(clusterName)
	if err != nil {
		return fmt.Errorf("checking cluster %q: %w", clusterName, err)
	}
	if !exists {
		return fmt.Errorf("cluster %q not found; create it first with \"sew create\"", clusterName)
	}

	logDir := filepath.Join(sewHome, "logs")
	if len(cfg.From) > 0 {
		logDir = filepath.Join(logDir, strings.Join(cfg.From, "_"))
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("creating log directory %s: %w", logDir, err)
	}
	logFile, err := os.OpenFile(filepath.Join(logDir, "patch.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer logFile.Close()
	klog.LogToStderr(false)
	klog.SetOutput(logFile)

	registry.MergeComponents(resolved, cfg.Components, cfg.Dir)
	resolved.Repos = registry.MergeRepos(resolved.Repos, cfg.Helm.Repos)

	patchFile := args[0]
	patch, err := config.Load(patchFile)
	if err != nil {
		return fmt.Errorf("loading patch file %s: %w", patchFile, err)
	}

	patchedNames := make(map[string]bool, len(patch.Components))
	for _, c := range patch.Components {
		patchedNames[c.Name] = true
	}
	if len(patchedNames) == 0 {
		color.Yellow("  ⚠ patch file contains no components; nothing to do")
		return nil
	}

	registry.MergeComponents(resolved, patch.Components, patch.Dir)
	resolved.Repos = registry.MergeRepos(resolved.Repos, patch.Helm.Repos)

	ctx := context.Background()

	filter := func(c config.Component) bool {
		return patchedNames[c.Name]
	}

	if err := installComponents(ctx, resolved, filter); err != nil {
		return err
	}

	fmt.Println()
	color.Blue("  Patch applied in %s", time.Since(start).Round(time.Millisecond))

	return nil
}
