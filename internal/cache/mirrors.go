package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/core"
)

type MirrorConfig struct {
	Patch  string
	Mounts []core.Mount
}

// PrepareMirrors writes containerd 2.x file-based registry mirror
// configuration (hosts.toml per upstream) and returns a containerd config patch
// plus the extra mount needed to expose the config inside Kind nodes.
func PrepareMirrors(cfg *core.MirrorsConfig, sewHome string) (*MirrorConfig, error) {
	hostsDir := filepath.Join(sewHome, "mirrors", "containerd-hosts")
	upstreams := AllUpstreams(cfg)

	for _, upstream := range upstreams {
		dir := filepath.Join(hostsDir, upstream)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating hosts dir for %s: %w", upstream, err)
		}

		name := ContainerName(upstream)
		var b strings.Builder
		fmt.Fprintf(&b, "server = %q\n\n", remoteURL(upstream))
		fmt.Fprintf(&b, "[host.\"http://%s:5000\"]\n", name)
		b.WriteString("  capabilities = [\"pull\", \"resolve\"]\n")

		hostsFile := filepath.Join(dir, "hosts.toml")
		if err := os.WriteFile(hostsFile, []byte(b.String()), 0o644); err != nil {
			return nil, fmt.Errorf("writing hosts.toml for %s: %w", upstream, err)
		}
	}

	return &MirrorConfig{
		Patch: "[plugins.\"io.containerd.grpc.v1.cri\".registry]\n  config_path = \"/etc/containerd/certs.d\"",
		Mounts: []core.Mount{{
			HostPath:      hostsDir,
			ContainerPath: "/etc/containerd/certs.d",
		}},
	}, nil
}
