package registry

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type netrcEntry struct {
	machine  string
	login    string
	password string
}

// parseNetrc parses a .netrc file into a list of entries.
// See https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html
func parseNetrc(data string) []netrcEntry {
	var entries []netrcEntry
	var cur netrcEntry
	inMacro := false
	done := false

	for _, line := range strings.Split(data, "\n") {
		if done {
			break
		}
		if inMacro {
			if line == "" {
				inMacro = false
			}
			continue
		}

		fields := strings.Fields(line)
		i := 0
		for ; i < len(fields)-1; i += 2 {
			if fields[i] == "default" {
				done = true
				break
			}
			switch fields[i] {
			case "machine":
				cur = netrcEntry{machine: fields[i+1]}
			case "login":
				cur.login = fields[i+1]
			case "password":
				cur.password = fields[i+1]
			case "macdef":
				inMacro = true
			}
			if cur.machine != "" && cur.login != "" && cur.password != "" {
				entries = append(entries, cur)
				cur = netrcEntry{}
			}
		}
		if !done && i < len(fields) && fields[i] == "default" {
			break
		}
	}
	return entries
}

// netrcPath returns the path to the .netrc file.
// It checks $NETRC first, then falls back to ~/.netrc (or ~/_netrc on Windows).
func netrcPath() (string, error) {
	if env := os.Getenv("NETRC"); env != "" {
		return env, nil
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		legacy := filepath.Join(dir, "_netrc")
		if _, err := os.Stat(legacy); err == nil {
			return legacy, nil
		}
	}
	return filepath.Join(dir, ".netrc"), nil
}

// lookupNetrc reads the .netrc file and returns credentials for the given
// host. Returns empty strings and false when no match is found or the file
// does not exist.
func lookupNetrc(host string) (login, password string, ok bool) {
	path, err := netrcPath()
	if err != nil {
		return "", "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", false
	}
	for _, e := range parseNetrc(string(data)) {
		if e.machine == host {
			return e.login, e.password, true
		}
	}
	return "", "", false
}

// netrcTransport wraps an http.RoundTripper and injects Basic auth credentials.
type netrcTransport struct {
	base     http.RoundTripper
	login    string
	password string
}

func (t *netrcTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.SetBasicAuth(t.login, t.password)
	return t.base.RoundTrip(r)
}

// newAuthenticatedClient returns an *http.Client that injects .netrc
// credentials for the host extracted from registryURL. Returns nil when
// no credentials are found.
func newAuthenticatedClient(registryURL string) *http.Client {
	u, err := url.Parse(registryURL)
	if err != nil || u.Host == "" {
		return nil
	}
	login, password, ok := lookupNetrc(u.Hostname())
	if !ok {
		return nil
	}
	return &http.Client{
		Transport: &netrcTransport{
			base:     http.DefaultTransport,
			login:    login,
			password: password,
		},
	}
}
