package cmd

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

func requireDocker() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("docker is not running — start Docker and try again")
	}
	defer cli.Close()
	if _, err := cli.Ping(context.Background()); err != nil {
		return fmt.Errorf("docker is not running — start Docker and try again")
	}
	return nil
}
