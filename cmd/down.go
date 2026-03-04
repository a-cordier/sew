package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"strings"
	"syscall"

	"github.com/a-cordier/sew/internal/cache"
	"github.com/a-cordier/sew/internal/cloudprovider"
	"github.com/a-cordier/sew/internal/dns"
	"github.com/a-cordier/sew/internal/kind"
	"github.com/a-cordier/sew/internal/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"
)

var downCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the cluster defined in the config",
	RunE: func(_ *cobra.Command, _ []string) error {
		start := time.Now()

		resolved, err := resolveContextConfig()
		if err != nil {
			return err
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

		dnsDir := filepath.Join(sewHome, "dns")
		if err := dns.RemoveRecordFile(dnsDir, cfg.Kind.Name); err != nil {
			color.Yellow("  ⚠ failed to remove DNS record file: %v", err)
		} else {
			color.Blue("  ✓ Removed DNS records for cluster %q", cfg.Kind.Name)
		}

		if err := logger.WithSpinner("Cleaning up load balancer containers", func() error {
			return cloudprovider.CleanupLBs(cfg.Kind.Name)
		}); err != nil {
			return err
		}

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

		if cfg.Images.Preload != nil && len(cfg.Images.Preload.Refs) > 0 {
			ctx := context.Background()
			if err := logger.WithSpinner("Stopping preload registry", func() error {
				return cache.StopPreloadRegistry(ctx)
			}); err != nil {
				return err
			}
		}

		stopCPKIfNoKindClusters()
		stopDNSIfNoRecords()

		fmt.Println()
		color.Blue("  Total: %s", time.Since(start).Round(time.Millisecond))

		if resolved != nil {
			printNotes(resolved.Notes.Delete, cfg)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func stopDNSIfNoRecords() {
	dnsDir := filepath.Join(sewHome, "dns")
	entries, err := os.ReadDir(dnsDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			return
		}
	}

	pidPath := filepath.Join(sewHome, "pids", "dns.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	if proc.Signal(syscall.Signal(0)) == nil {
		_ = proc.Signal(syscall.SIGTERM)
		color.Blue("  ✓ Stopped DNS server (pid %d)", pid)
	}
	_ = os.Remove(pidPath)
}

func stopCPKIfNoKindClusters() {
	provider := kindcluster.NewProvider()
	clusters, err := provider.List()
	if err != nil || len(clusters) > 0 {
		return
	}

	pidPath := filepath.Join(sewHome, "pids", "cpk.pid")

	if cloudprovider.NeedsTunnels() {
		// On macOS, CPK runs as root -- use sudo to terminate it.
		cmd := exec.Command("sudo", "-p",
			"\n  sew needs administrator privileges to stop the cloud provider controller.\n  Password: ",
			"pkill", "-f", "sew.*cpk serve")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			color.Blue("  ✓ Stopped cloud provider controller")
		}
	} else {
		data, err := os.ReadFile(pidPath)
		if err != nil {
			return
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			return
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			return
		}
		if proc.Signal(syscall.Signal(0)) == nil {
			_ = proc.Signal(syscall.SIGTERM)
			color.Blue("  ✓ Stopped cloud provider controller (pid %d)", pid)
		}
	}

	_ = os.Remove(pidPath)
}
