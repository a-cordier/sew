package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/a-cordier/sew/core"
	"github.com/a-cordier/sew/internal/cache"
	"github.com/a-cordier/sew/internal/cloudprovider"
	"github.com/a-cordier/sew/internal/dns"
	"github.com/a-cordier/sew/internal/installer"
	"github.com/a-cordier/sew/internal/kind"
	"github.com/a-cordier/sew/internal/logger"
	"github.com/a-cordier/sew/internal/registry"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create the cluster and install the context",
	RunE:  runUp,
}

func init() {
	rootCmd.AddCommand(upCmd)
}

func runUp(_ *cobra.Command, _ []string) error {
	if err := os.MkdirAll(sewHome, 0o755); err != nil {
		return fmt.Errorf("failed to create home directory %s: %w", sewHome, err)
	}

	var resolved *core.ResolvedContext
	if cfg.Registry != "" && cfg.Context != "" {
		registryURL := cfg.Registry
		if strings.HasPrefix(registryURL, "file://") {
			path := strings.TrimPrefix(registryURL, "file://")
			if abs, err := filepath.Abs(path); err == nil {
				registryURL = "file://" + abs
			}
		}
		resolver := registry.NewResolver(registryURL, sewHome)
		var resolveErr error
		resolved, resolveErr = resolver.Resolve(context.Background(), cfg.Context)
		if resolveErr != nil {
			return fmt.Errorf("resolving context %q: %w", cfg.Context, resolveErr)
		}
		cfg.Kind.MergeWithContext(resolved.Kind)
		cfg.Features = core.MergeFeatures(resolved.Features, cfg.Features)
	}

	featWarnings, err := core.ResolveFeatureDependencies(&cfg.Features)
	if err != nil {
		return fmt.Errorf("validating feature dependencies: %w", err)
	}
	for _, w := range featWarnings {
		color.Yellow("  ⚠ %s", w)
	}

	logDir := filepath.Join(sewHome, "logs")
	if cfg.Context != "" {
		logDir = filepath.Join(logDir, cfg.Context)
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("creating log directory %s: %w", logDir, err)
	}
	logFile, err := os.OpenFile(filepath.Join(logDir, "install.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer logFile.Close()
	klog.LogToStderr(false)
	klog.SetOutput(logFile)

	ctx := context.Background()

	if cfg.Images.Mirrors != nil {
		if err := logger.WithSpinner("Starting image mirror proxies", func() error {
			return cache.EnsureProxies(ctx, cfg.Images.Mirrors, sewHome)
		}); err != nil {
			return err
		}
		mirrors, err := cache.PrepareMirrors(cfg.Images.Mirrors, sewHome)
		if err != nil {
			return fmt.Errorf("preparing image mirrors: %w", err)
		}
		cfg.Kind.ContainerdConfigPatches = append(cfg.Kind.ContainerdConfigPatches, mirrors.Patch)
		for i := range cfg.Kind.Nodes {
			cfg.Kind.Nodes[i].ExtraMounts = append(cfg.Kind.Nodes[i].ExtraMounts, mirrors.Mounts...)
		}
	}

	kindConfig, err := cfg.Kind.RawYAML()
	if err != nil {
		return fmt.Errorf("serializing kind config: %w", err)
	}
	if err := logger.WithSpinner(
		fmt.Sprintf("Creating cluster %q", cfg.Kind.Name),
		func() error { return kind.Create(cfg.Kind.Name, kindConfig) },
	); err != nil {
		return err
	}
	if err := kind.RemoveControlPlaneTaint(cfg.Kind.Name); err != nil {
		return err
	}

	if cfg.Images.Mirrors != nil {
		if err := logger.WithSpinner("Connecting image mirrors to Kind network", func() error {
			return cache.ConnectToKindNetwork(ctx, cfg.Kind.Name, cfg.Images.Mirrors)
		}); err != nil {
			return err
		}
	}

	if cfg.Registry == "" || cfg.Context == "" {
		return nil
	}

	gatewayEnabled := cfg.Features.Gateway != nil && cfg.Features.Gateway.Enabled
	lbEnabled := cfg.Features.LB != nil && cfg.Features.LB.Enabled

	if lbEnabled {
		if err := ensureCPKController(cfg); err != nil {
			return err
		}
	}

	registry.MergeComponents(resolved, cfg.Components, cfg.Dir)
	resolved.Repos = registry.MergeRepos(resolved.Repos, cfg.Repos)

	if gatewayEnabled {
		injectGatewayComponents(resolved)
	}

	if err := registry.Validate(resolved.Components); err != nil {
		return fmt.Errorf("validating components: %w", err)
	}
	sorted, err := registry.TopoSort(resolved.Components)
	if err != nil {
		return fmt.Errorf("resolving dependencies: %w", err)
	}

	helmInst, _ := installer.ForType("helm")
	if hi, ok := helmInst.(*installer.HelmInstaller); ok {
		if err := logger.WithSpinner("Initializing Helm", func() error {
			return hi.AddRepos(resolved.Repos, sewHome)
		}); err != nil {
			return err
		}
	}

	compByName := make(map[string]core.Component)
	for _, c := range resolved.Components {
		compByName[c.Name] = c
	}

	const defaultReadyTimeout = 5 * time.Minute
	for _, comp := range sorted {
		for _, req := range comp.Requires {
			if req.Conditions.Ready {
				dep := compByName[req.Component]
				depNamespace := dep.Namespace
				if depNamespace == "" {
					depNamespace = "default"
				}
				timeout := defaultReadyTimeout
				if req.Timeout != "" {
					if d, err := time.ParseDuration(req.Timeout); err == nil && d > 0 {
						timeout = d
					}
				}
				var matchLabels map[string]string
				if req.Selector != nil && len(req.Selector.MatchLabels) > 0 {
					matchLabels = req.Selector.MatchLabels
				}
				if err := logger.WithSpinner(fmt.Sprintf("Waiting for %q to be ready", req.Component), func() error {
					return installer.WaitForReady(ctx, req.Component, depNamespace, timeout, matchLabels)
				}); err != nil {
					return fmt.Errorf("requirement %q not ready: %w", req.Component, err)
				}
			}
		}
		inst, err := installer.ForType(comp.EffectiveType())
		if err != nil {
			return fmt.Errorf("component %q: %w", comp.Name, err)
		}
		comp := comp
		if err := logger.WithSpinner(fmt.Sprintf("Installing %q", comp.Name), func() error {
			return inst.Install(ctx, comp, resolved.Dir)
		}); err != nil {
			return err
		}
	}

	if cfg.Features.DNS != nil && cfg.Features.DNS.Enabled {
		if err := setupDNSRecords(ctx, cfg); err != nil {
			return err
		}
	}

	return nil
}

// injectGatewayComponents prepends the shared sew-gateway Gateway resource to
// the resolved component list so the topo-sort places it before any
// user-defined HTTPRoutes. CPK's gateway controller (cloud-provider-kind)
// handles provisioning the Envoy data-plane container automatically.
func injectGatewayComponents(resolved *core.ResolvedContext) {
	sewGW := core.Component{
		Name: "sew-gateway",
		Type: "k8s",
		K8s: &core.K8sSpec{
			Manifests: []map[string]interface{}{
				{
					"apiVersion": "gateway.networking.k8s.io/v1",
					"kind":       "Gateway",
					"metadata":   map[string]interface{}{"name": "sew-gateway"},
					"spec": map[string]interface{}{
						"gatewayClassName": "cloud-provider-kind",
						"listeners": []map[string]interface{}{
							{
								"name": "http", "protocol": "HTTP", "port": 80,
								"allowedRoutes": map[string]interface{}{
									"namespaces": map[string]interface{}{"from": "All"},
								},
							},
						},
					},
				},
			},
		},
	}
	resolved.Components = append([]core.Component{sewGW}, resolved.Components...)
}

// The CPK controller is restarted fresh on each sew up, so it discovers the
// cluster immediately. 90s gives time for the CCM to start, informers to
// sync, and the gateway controller to reconcile and set status.addresses.
const gatewayPollTimeout = 90 * time.Second

func setupDNSRecords(ctx context.Context, cfg *core.Config) error {
	dnsDir := filepath.Join(sewHome, "dns")
	var dnsRecords []core.DNSRecord
	if cfg.Features.DNS.Records != nil {
		dnsRecords = cfg.Features.DNS.Records
	}

	// Always introspect Gateways when DNS is enabled. Gateway API resources
	// may come from sew's own sew-gateway (features.gateway) or from
	// user-installed controllers (e.g. GKO). Both set status.addresses.
	if err := logger.WithSpinner("Collecting DNS records from cluster", func() error {
		return dns.IntrospectCluster(ctx, cfg.Kind.Name, dnsDir, gatewayPollTimeout, true, dnsRecords)
	}); err != nil {
		return err
	}

	if err := ensureDNSServer(cfg); err != nil {
		color.Yellow("  ⚠ failed to start DNS server: %v", err)
	}

	if !dns.ResolverConfigured(cfg.Features.DNS.Domain, cfg.Features.DNS.Port) {
		fmt.Println()
		color.Yellow("  DNS server is running but OS-level routing is not configured.")
		color.Yellow("  Run \"sew setup dns\" once to enable automatic DNS resolution.")
	}

	return nil
}

func ensureCPKController(_ *core.Config) error {
	pidDir := filepath.Join(sewHome, "pids")
	pidPath := filepath.Join(pidDir, "cpk.pid")

	// Kill any stale controller so the new one immediately discovers the
	// current cluster on its first poll iteration (no 30-second wait).
	if cloudprovider.NeedsTunnels() {
		// On macOS, CPK runs as root (via sudo). A non-root process cannot
		// signal a root process, so killProcess (which uses os.Signal) is
		// ineffective. Use sudo pkill to kill ALL root-owned CPK processes
		// accumulated from prior sew up invocations.
		_ = exec.Command("sudo", "-n", "pkill", "-f", "sew.*cpk serve").Run()
	} else {
		killProcess(pidPath)
	}

	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		return fmt.Errorf("creating pid directory: %w", err)
	}

	sewBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding sew executable: %w", err)
	}

	var cmdArgs []string
	if cfgFile != "" {
		cmdArgs = append(cmdArgs, "--config", cfgFile)
	}
	cmdArgs = append(cmdArgs, "cpk", "serve")

	// On macOS, CPK needs root for loopback aliases and tunnels.
	if cloudprovider.NeedsTunnels() {
		fullCmd := sewBin
		for _, a := range cmdArgs {
			fullCmd += " " + a
		}
		cmd := exec.Command("sudo", "-p",
			"\n  sew needs administrator privileges for network routing.\n  Password: ",
			"sh", "-c", fullCmd+" &")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("starting cloud provider controller: %w", err)
		}

		// sudo spawns the background process; find its PID.
		time.Sleep(500 * time.Millisecond)
		pidCmd := exec.Command("pgrep", "-f", "sew.*cpk serve")
		out, err := pidCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			if len(lines) > 0 {
				if err := os.WriteFile(pidPath, []byte(lines[0]+"\n"), 0o644); err != nil {
					return fmt.Errorf("writing CPK PID file: %w", err)
				}
				color.Blue("  ✓ Cloud provider controller started (pid %s)", lines[0])
			}
		}
		return nil
	}

	cmd := exec.Command(sewBin, cmdArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting cloud provider controller: %w", err)
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing CPK PID file: %w", err)
	}

	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("releasing CPK process: %w", err)
	}

	color.Blue("  ✓ Cloud provider controller started (pid %d)", pid)
	return nil
}

func killProcess(pidPath string) {
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
	_ = proc.Signal(syscall.SIGKILL)
	_, _ = proc.Wait()
	_ = os.Remove(pidPath)
}

func dnsServerParams(cfg *core.Config) (domain string, port int, dir string) {
	domain = core.DNSDefaultDomain
	port = core.DNSDefaultPort
	dir = filepath.Join(sewHome, "dns")
	if cfg.Features.DNS != nil {
		if cfg.Features.DNS.Domain != "" {
			domain = cfg.Features.DNS.Domain
		}
		if cfg.Features.DNS.Port != 0 {
			port = cfg.Features.DNS.Port
		}
	}
	return
}

// startDNSServer launches a new DNS server process and writes its PID file.
func startDNSServer(domain string, port int, dir string) error {
	pidDir := filepath.Join(sewHome, "pids")
	pidPath := filepath.Join(pidDir, "dns.pid")

	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		return fmt.Errorf("creating pid directory: %w", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	sewBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding sew executable: %w", err)
	}

	cmd := exec.Command(sewBin, "dns", "serve",
		"--dir", dir,
		"--domain", domain,
		"--addr", addr,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting DNS server: %w", err)
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing DNS PID file: %w", err)
	}

	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("releasing DNS server process: %w", err)
	}

	color.Blue("  ✓ DNS server started (pid %d, %s)", pid, addr)
	return nil
}

func ensureDNSServer(cfg *core.Config) error {
	pidPath := filepath.Join(sewHome, "pids", "dns.pid")
	killProcess(pidPath)
	domain, port, dir := dnsServerParams(cfg)
	return startDNSServer(domain, port, dir)
}

// ensureDNSServerRunning starts the DNS server only if it is not already alive.
// Used by "sew dns refresh" to restart a server that exited early (e.g. because
// no record files existed at initial startup).
func ensureDNSServerRunning(cfg *core.Config) error {
	pidPath := filepath.Join(sewHome, "pids", "dns.pid")
	if isProcessAlive(pidPath) {
		return nil
	}
	domain, port, dir := dnsServerParams(cfg)
	return startDNSServer(domain, port, dir)
}

func isProcessAlive(pidPath string) bool {
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
