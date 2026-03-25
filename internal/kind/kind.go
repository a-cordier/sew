package kind

import (
	"fmt"
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
)

const waitForReadyTimeout = 2 * time.Minute

func Create(name string, rawConfig []byte) error {
	provider := cluster.NewProvider()

	exists, err := Exists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return provider.Create(
		name,
		cluster.CreateWithRawConfig(rawConfig),
		cluster.CreateWithWaitForReady(waitForReadyTimeout),
	)
}

func Delete(name string) error {
	provider := cluster.NewProvider()

	exists, err := Exists(name)
	if err != nil {
		return err
	}
	if !exists {
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
