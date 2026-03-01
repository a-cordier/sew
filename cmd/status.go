package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/a-cordier/sew/core"
	"github.com/a-cordier/sew/internal/cloudprovider"
	"github.com/a-cordier/sew/internal/dns"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the current sew environment",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(_ *cobra.Command, _ []string) error {
	bold := color.New(color.Bold)

	bold.Println("Cluster")
	fmt.Printf("  Name: %s\n", cfg.Kind.Name)
	fmt.Println()

	printFeatureStatus(bold)
	printLBStatus(bold)
	printDNSStatus(bold)
	return nil
}

func printFeatureStatus(bold *color.Color) {
	bold.Println("Features")

	lbEnabled := cfg.Features.LB != nil && cfg.Features.LB.Enabled
	gwEnabled := cfg.Features.Gateway != nil && cfg.Features.Gateway.Enabled
	dnsEnabled := cfg.Features.DNS != nil && cfg.Features.DNS.Enabled

	fmt.Printf("  lb:      %s\n", enabledStr(lbEnabled))
	gwLine := enabledStr(gwEnabled)
	if gwEnabled && cfg.Features.Gateway.Channel != "" {
		gwLine += fmt.Sprintf(" (channel: %s)", cfg.Features.Gateway.Channel)
	}
	fmt.Printf("  gateway:       %s\n", gwLine)

	dnsLine := enabledStr(dnsEnabled)
	if dnsEnabled {
		domain := cfg.Features.DNS.Domain
		if domain == "" {
			domain = "sew.local"
		}
		port := cfg.Features.DNS.Port
		if port == 0 {
			port = core.DNSDefaultPort
		}
		dnsLine += fmt.Sprintf(" (domain: %s, port: %d)", domain, port)
	}
	fmt.Printf("  dns:           %s\n", dnsLine)
	fmt.Println()
}

func printLBStatus(bold *color.Color) {
	bold.Println("Load Balancers")
	ips, err := cloudprovider.ListLBIPs(cfg.Kind.Name)
	if err != nil {
		color.Yellow("  could not list LB containers: %v", err)
		fmt.Println()
		return
	}
	if len(ips) == 0 {
		fmt.Println("  (none)")
	} else {
		for name, ip := range ips {
			fmt.Printf("  %s → %s\n", name, ip)
		}
	}
	fmt.Println()
}

func printDNSStatus(bold *color.Color) {
	bold.Println("DNS")

	dnsEnabled := cfg.Features.DNS != nil && cfg.Features.DNS.Enabled
	if !dnsEnabled {
		fmt.Println("  (disabled)")
		return
	}

	domain := cfg.Features.DNS.Domain
	if domain == "" {
		domain = "sew.local"
	}
	port := cfg.Features.DNS.Port
	if port == 0 {
		port = core.DNSDefaultPort
	}

	resolverOK := dns.ResolverConfigured(domain, port)
	if resolverOK {
		color.Blue("  resolver: configured for %s", domain)
	} else {
		color.Yellow("  resolver: not configured (run \"sew setup dns\")")
	}

	serverRunning := isDNSServerRunning(port)
	if serverRunning {
		color.Blue("  server:   running on 127.0.0.1:%d", port)
	} else {
		color.Yellow("  server:   not running")
	}

	dnsDir := filepath.Join(sewHome, "dns")
	printDNSRecords(dnsDir)
}

func printDNSRecords(dnsDir string) {
	entries, err := os.ReadDir(dnsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("  records:  (none)")
			return
		}
		color.Yellow("  records:  could not read %s: %v", dnsDir, err)
		return
	}

	var totalRecords int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dnsDir, e.Name()))
		if err != nil {
			continue
		}
		var rf dns.RecordFile
		if err := json.Unmarshal(data, &rf); err != nil {
			continue
		}
		cluster := strings.TrimSuffix(e.Name(), ".json")
		for host, ip := range rf.Records {
			if totalRecords == 0 {
				fmt.Println("  records:")
			}
			fmt.Printf("    %s → %s (%s)\n", host, ip, cluster)
			totalRecords++
		}
	}
	if totalRecords == 0 {
		fmt.Println("  records:  (none)")
	}
}

func isDNSServerRunning(_ int) bool {
	pidPath := filepath.Join(sewHome, "pids", "dns.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func enabledStr(enabled bool) string {
	if enabled {
		return color.BlueString("enabled")
	}
	return "disabled"
}
