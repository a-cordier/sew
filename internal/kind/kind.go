package kind

import (
	"fmt"
	"os/exec"
	"strings"

	"sigs.k8s.io/kind/pkg/cluster"
)

func Create(name string, rawConfig []byte) error {
	provider := cluster.NewProvider()

	exists, err := Exists(name)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Cluster %q already exists, skipping creation\n", name)
		return nil
	}

	return provider.Create(
		name,
		cluster.CreateWithRawConfig(rawConfig),
	)
}

// RemoveControlPlaneTaint removes the NoSchedule taint from control-plane
// nodes so that workloads can be scheduled on single-node clusters.
// Silently succeeds when the taint is already absent.
func RemoveControlPlaneTaint(name string) error {
	out, err := exec.Command(
		"docker", "exec", "--privileged", name+"-control-plane",
		"kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"taint", "nodes", "--all",
		"node-role.kubernetes.io/control-plane-",
	).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "not found") {
			return nil
		}
		return fmt.Errorf("removing control-plane taint: %s: %w", string(out), err)
	}
	return nil
}

func Delete(name string) error {
	provider := cluster.NewProvider()

	exists, err := Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Printf("Cluster %q does not exist, skipping deletion\n", name)
		return nil
	}

	return provider.Delete(name, "")
}

func Exists(name string) (bool, error) {
	provider := cluster.NewProvider()

	clusters, err := provider.List()
	if err != nil {
		return false, fmt.Errorf("listing Kind clusters: %w", err)
	}

	for _, c := range clusters {
		if c == name {
			return true, nil
		}
	}
	return false, nil
}
