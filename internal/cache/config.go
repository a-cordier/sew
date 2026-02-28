package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type dockerConfigFile struct {
	Auths map[string]dockerAuthEntry `json:"auths"`
}

type dockerAuthEntry struct {
	Auth string `json:"auth"`
}

type registryCredentials struct {
	Username string
	Password string
}

var dockerIOKeys = []string{
	"docker.io",
	"index.docker.io",
	"https://index.docker.io/v1/",
	"registry-1.docker.io",
}

func loadDockerConfig() (*dockerConfigFile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return &dockerConfigFile{}, nil
	}
	data, err := os.ReadFile(filepath.Join(home, ".docker", "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &dockerConfigFile{}, nil
		}
		return nil, fmt.Errorf("reading docker config: %w", err)
	}
	var cfg dockerConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing docker config: %w", err)
	}
	return &cfg, nil
}

// lookupCredentials searches the Docker config for credentials matching the
// given upstream. For docker.io several well-known key variants are tried.
func lookupCredentials(cfg *dockerConfigFile, upstream string) *registryCredentials {
	if cfg == nil || len(cfg.Auths) == 0 {
		return nil
	}

	candidates := []string{upstream, "https://" + upstream}
	if upstream == "docker.io" {
		candidates = dockerIOKeys
	}

	for _, key := range candidates {
		entry, ok := cfg.Auths[key]
		if !ok || entry.Auth == "" {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(entry.Auth)
		if err != nil {
			continue
		}
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			continue
		}
		return &registryCredentials{Username: parts[0], Password: parts[1]}
	}
	return nil
}

func remoteURL(upstream string) string {
	if upstream == "docker.io" {
		return "https://registry-1.docker.io"
	}
	return "https://" + upstream
}

// generateProxyConfig builds a registry:2 YAML config for a pull-through proxy
// targeting the given upstream. Credentials from ~/.docker/config.json are
// injected when present.
func generateProxyConfig(upstream string, creds *registryCredentials) []byte {
	var b strings.Builder
	b.WriteString("version: 0.1\n")
	b.WriteString("proxy:\n")
	fmt.Fprintf(&b, "  remoteurl: %s\n", remoteURL(upstream))
	if creds != nil {
		fmt.Fprintf(&b, "  username: %s\n", creds.Username)
		fmt.Fprintf(&b, "  password: %s\n", creds.Password)
	}
	b.WriteString("storage:\n")
	b.WriteString("  filesystem:\n")
	b.WriteString("    rootdirectory: /var/lib/registry\n")
	b.WriteString("http:\n")
	b.WriteString("  addr: :5000\n")
	return []byte(b.String())
}

// writeProxyConfig generates and writes a registry:2 pull-through proxy config
// file at path. Auth credentials for the upstream host are read from
// ~/.docker/config.json when available; otherwise the proxy runs anonymously.
func writeProxyConfig(path, upstream string) error {
	dockerCfg, err := loadDockerConfig()
	if err != nil {
		return err
	}
	data := generateProxyConfig(upstream, lookupCredentials(dockerCfg, upstream))
	return os.WriteFile(path, data, 0o644)
}
