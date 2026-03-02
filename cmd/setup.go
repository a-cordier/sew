package cmd

import (
	"fmt"

	"github.com/a-cordier/sew/core"
	"github.com/a-cordier/sew/internal/dns"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-time setup tasks",
}

var setupDNSCmd = &cobra.Command{
	Use:   "dns",
	Short: "Configure OS-level DNS routing for sew domains",
	Long: `Configure the operating system to route DNS queries for the sew domain
(e.g. *.sew.local) to the local DNS server.

On macOS, this creates a per-domain resolver file in /etc/resolver/.
On Linux, this configures systemd-resolved on the loopback interface.

This is a one-time operation that requires administrator privileges.
After setup, "sew start" and "sew stop" run without sudo.`,
	RunE: runSetupDNS,
}

var teardownDNSCmd = &cobra.Command{
	Use:   "dns",
	Short: "Remove OS-level DNS routing for sew domains",
	Long: `Remove the DNS routing configuration previously created by "sew setup dns".

On macOS, this removes the resolver file from /etc/resolver/.
On Linux, this reverts the systemd-resolved configuration on loopback.`,
	RunE: runTeardownDNS,
}

var teardownCmd = &cobra.Command{
	Use:   "teardown",
	Short: "Undo one-time setup tasks",
}

func init() {
	setupCmd.AddCommand(setupDNSCmd)
	teardownCmd.AddCommand(teardownDNSCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(teardownCmd)
}

func runSetupDNS(_ *cobra.Command, _ []string) error {
	if _, err := resolveContextConfig(); err != nil {
		return err
	}

	domain := core.DNSDefaultDomain
	port := core.DNSDefaultPort

	if cfg.Features.DNS != nil {
		if cfg.Features.DNS.Domain != "" {
			domain = cfg.Features.DNS.Domain
		}
		if cfg.Features.DNS.Port != 0 {
			port = cfg.Features.DNS.Port
		}
	}

	if dns.ResolverConfigured(domain, port) {
		color.Blue("  ✓ DNS routing for %q is already configured", domain)
		return nil
	}

	fmt.Printf("  Configuring DNS routing: *.%s → 127.0.0.1:%d\n", domain, port)
	if err := dns.SetupResolver(domain, port); err != nil {
		return fmt.Errorf("setting up DNS resolver: %w", err)
	}

	color.Blue("  ✓ DNS routing configured for %q", domain)
	return nil
}

func runTeardownDNS(_ *cobra.Command, _ []string) error {
	if _, err := resolveContextConfig(); err != nil {
		return err
	}

	domain := core.DNSDefaultDomain
	if cfg.Features.DNS != nil && cfg.Features.DNS.Domain != "" {
		domain = cfg.Features.DNS.Domain
	}

	fmt.Printf("  Removing DNS routing for %q\n", domain)
	if err := dns.TeardownResolver(domain); err != nil {
		return fmt.Errorf("tearing down DNS resolver: %w", err)
	}

	color.Blue("  ✓ DNS routing removed for %q", domain)
	return nil
}
