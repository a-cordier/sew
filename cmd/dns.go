package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/a-cordier/sew/core"
	"github.com/a-cordier/sew/internal/dns"
	"github.com/a-cordier/sew/internal/registry"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Manage the local DNS server",
}

var (
	dnsDir      string
	dnsDomain   string
	dnsAddr     string
	dnsUpstream string
)

var dnsServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the local DNS server",
	Long: `Start the local DNS server that resolves hostnames from per-cluster
record files. The server watches the record directory for changes, hot-reloads
records, and shuts itself down automatically when all record files are removed.

This command is typically started as a background process by "sew up" and does
not need to be invoked directly.`,
	RunE: runDNSServe,
}

var dnsRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Re-collect DNS records from the cluster",
	Long: `Re-run DNS introspection against the current Kind cluster. This picks up
Gateways and LoadBalancer services that were created after "sew up" finished.
The running DNS server hot-reloads the updated record files automatically.`,
	RunE: runDNSRefresh,
}

func init() {
	dnsServeCmd.Flags().StringVar(&dnsDir, "dir", "", "path to DNS record directory (default: $SEW_HOME/dns)")
	dnsServeCmd.Flags().StringVar(&dnsDomain, "domain", core.DNSDefaultDomain, "DNS domain to serve")
	dnsServeCmd.Flags().StringVar(&dnsAddr, "addr", "", "UDP listen address (default: 127.0.0.1:<port>)")
	dnsServeCmd.Flags().StringVar(&dnsUpstream, "upstream", "8.8.8.8:53", "upstream DNS server for non-matching queries")

	dnsCmd.AddCommand(dnsServeCmd)
	dnsCmd.AddCommand(dnsRefreshCmd)
	rootCmd.AddCommand(dnsCmd)
}

func runDNSServe(cmd *cobra.Command, _ []string) error {
	if dnsDir == "" {
		dnsDir = filepath.Join(sewHome, "dns")
	}
	if err := os.MkdirAll(dnsDir, 0o755); err != nil {
		return fmt.Errorf("creating DNS record directory: %w", err)
	}

	if !cmd.Flags().Changed("domain") && cfg.Features.DNS != nil && cfg.Features.DNS.Domain != "" {
		dnsDomain = cfg.Features.DNS.Domain
	}

	if dnsAddr == "" {
		port := core.DNSDefaultPort
		if cfg.Features.DNS != nil && cfg.Features.DNS.Port != 0 {
			port = cfg.Features.DNS.Port
		}
		dnsAddr = fmt.Sprintf("127.0.0.1:%d", port)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return dns.Run(ctx, dns.Config{
		Dir:      dnsDir,
		Domain:   dnsDomain,
		Addr:     dnsAddr,
		Upstream: dnsUpstream,
	})
}

const refreshPollTimeout = 30 * time.Second

func runDNSRefresh(_ *cobra.Command, _ []string) error {
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
		cfg.Features = core.MergeFeatures(resolved.Features, cfg.Features)
	}

	dnsDir := filepath.Join(sewHome, "dns")
	if err := os.MkdirAll(dnsDir, 0o755); err != nil {
		return fmt.Errorf("creating DNS record directory: %w", err)
	}

	clusterName := cfg.Kind.Name
	var dnsRecords []core.DNSRecord
	if cfg.Features.DNS != nil && cfg.Features.DNS.Records != nil {
		dnsRecords = cfg.Features.DNS.Records
	}

	ctx := context.Background()
	if err := dns.IntrospectCluster(ctx, clusterName, dnsDir, refreshPollTimeout, true, dnsRecords); err != nil {
		return fmt.Errorf("introspecting cluster %q: %w", clusterName, err)
	}

	color.Blue("  ✓ DNS records refreshed for cluster %q", clusterName)

	if err := ensureDNSServerRunning(cfg); err != nil {
		color.Yellow("  ⚠ failed to start DNS server: %v", err)
	}

	return nil
}
