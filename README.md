# sew

**sew** spins up local Kubernetes clusters and deploys ready-to-use applications from a **registry** and **context**. You point it at a context (e.g. `gravitee.io/apim/db-less`), and it creates a Kind cluster and installs the components defined there (Helm charts, and in the future manifests or Kustomize).

## Concepts

- **Registry** — A tree of context directories, either on the filesystem (`file:///path`) or over HTTP. The binary does not ship a registry; you use your own or a remote one.
- **Context** — A path inside the registry following `org/product/variant` (e.g. `gravitee.io/apim/db-less`). Each context has a `context.yaml` that lists Helm repos and components (charts + values).
- **Config** — Your `config.yaml` sets the registry URL, the context to use, the Kind cluster definition, and optional overrides per component.

## Commands

| Command | Description |
|--------|-------------|
| `sew up` | Create the Kind cluster (if missing) and install the context: add Helm repos, then install each component (Helm upgrade --install). If no registry/context is configured, only creates the cluster. |
| `sew down` | Delete the Kind cluster defined in the config. |

### Global flags

- `--config <path>` — Config file to use (default: `./config.yaml` or `~/.sew/config.yaml`).
- `--context <path>` — Context path (e.g. `gravitee.io/apim/db-less`). Overrides the value from the config file.

## Quick start

1. **Config** — Create a `config.yaml` (or use the one in this repo) with at least:
   - `registry`: e.g. `file://./registry` for the local test registry, or an HTTP URL.
   - `context`: e.g. `gravitee.io/apim/db-less`.
   - `kind`: Kind cluster spec (name, nodes, port mappings).

2. **Run** — From the `sew` directory (or with `--config` pointing to this config):
   ```bash
   go run . up
   ```
   Or build and run:
   ```bash
   go build -o sew .
   ./sew up
   ```

3. **Tear down**:
   ```bash
   ./sew down
   ```

## Config format

```yaml
home: .sew
registry: file://./registry   # or https://...
context: gravitee.io/apim/db-less

kind:
  apiVersion: kind.x-k8s.io/v1alpha4
  kind: Cluster
  name: sew
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 443
      hostPort: 443

# Optional: override component values (e.g. chart version, extra values files)
# overrides:
#   apim:
#     helm:
#       version: "4.10.0"
#       values:
#         - ./my-overrides.yaml
```

Override paths under `overrides.*.helm.values` are resolved relative to the directory containing the config file.

## Context format

A context lives at `{registry}/{context_path}/` and must contain `context.yaml`:

```yaml
repos:
  - name: graviteeio
    url: https://helm.gravitee.io

components:
  - name: apim
    type: helm
    namespace: gravitee
    helm:
      chart: graviteeio/apim
      version: "4.11.0"
      values:
        - values-apim.yaml
```

File paths in `values` are relative to the context directory. `type` defaults to `helm` if omitted.
