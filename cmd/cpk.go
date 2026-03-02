package cmd

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	"sigs.k8s.io/cloud-provider-kind/pkg/config"
	"sigs.k8s.io/cloud-provider-kind/pkg/controller"
	"sigs.k8s.io/kind/pkg/cluster"
)

var cpkCmd = &cobra.Command{
	Use:    "cpk",
	Short:  "Manage the cloud-provider-kind controller",
	Hidden: true,
}

var cpkServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the cloud-provider-kind controller",
	Long: `Run CPK's controller manager as a background process. It watches for
Kind clusters, reconciles LoadBalancer services and Gateway API resources
(creating Docker proxy/envoy containers as needed).

This command is started automatically by "sew up" when the gateway feature
is enabled. It does not need to be invoked directly.`,
	RunE: runCPKServe,
}

func init() {
	cpkCmd.AddCommand(cpkServeCmd)
	rootCmd.AddCommand(cpkCmd)
}

func runCPKServe(_ *cobra.Command, _ []string) error {
	// Override the Portmap default set by cloudprovider.init(). The CPK
	// controller runs as a privileged background process and can handle
	// loopback aliases + userspace tunnels directly (Tunnel mode).
	if runtime.GOOS != "linux" {
		config.DefaultConfig.LoadBalancerConnectivity = config.Tunnel
		config.DefaultConfig.ControlPlaneConnectivity = config.Portmap
	}

	config.DefaultConfig.GatewayReleaseChannel = config.Standard

	option, err := cluster.DetectNodeProvider()
	if err != nil {
		return err
	}
	kindProvider := cluster.NewProvider(option)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Redirect stdout/stderr to /dev/null for background operation;
	// CPK logs via klog which writes to the log file set up by the caller.
	os.Stdout = nil
	os.Stderr = nil

	controller.New(kindProvider).Run(ctx)
	return nil
}
