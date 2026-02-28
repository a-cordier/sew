<p align="center"><img src="logo.svg" alt="sew" width="140"/><br><em>Your local Kubernetes tailor shop</em></p>

**sew** spins up local Kubernetes clusters and deploys ready-to-use applications from a **registry** and **context**. You point it at a context (e.g. `gravitee.io/apim/db-less`), and it creates a Kind cluster and installs the components defined there (Helm charts, and in the future manifests or Kustomize).

## Concepts

- **Registry** — A tree of context directories, either on the filesystem (`file:///path`) or over HTTP. The binary does not ship a registry; you use your own or a remote one.
- **Context** — A path inside the registry following `org/product/variant` (e.g. `gravitee.io/apim/db-less`). Each context has a `context.yaml` that lists Helm repos and components (charts + values). If you omit the variant (e.g. `gravitee.io/apim`), sew looks for a `.default` file to pick one automatically (see [Default variant resolution](#default-variant-resolution)).
- **Config** — Configuration is layered. A **user-level** config at `$SEW_HOME/sew.yaml` (default `~/.sew/sew.yaml`) provides base settings; a **project-level** `./sew.yaml` (or `--config`) is merged on top. Each layer sets the registry URL, the context to use, the Kind cluster definition, and optional overrides per component. Set the `SEW_HOME` environment variable to change the user-level config location.

## Commands

| Command | Description |
|--------|-------------|
| `sew up` | Create the Kind cluster (if missing) and install the context: add Helm repos, then install each component (Helm upgrade --install). If no registry/context is configured, only creates the cluster. |
| `sew down` | Delete the Kind cluster defined in the config. |

### Global flags

- `--config <path>` — Project-level config file to merge on top of the user-level base (`$SEW_HOME/sew.yaml`). Defaults to `./sew.yaml` when present.
- `--context <path>` — Context path (e.g. `gravitee.io/apim/db-less`). Overrides the value from the config file.

## Quick start

1. **Config** — Create a `sew.yaml` (or use the one in this repo) with at least:
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

## Default variant resolution

A context path usually includes the variant (`org/product/variant`), but you can also point to the product level and let sew pick the default variant automatically.

When the resolved path has no `context.yaml`, sew looks for a **`.default`** file in the same directory. This plain-text dotfile contains a single line — the name of the variant to use. sew then appends that variant to the path and resolves again.

```
registry/gravitee.io/apim/
├── .default          # contains "db-less"
├── db-less/
│   ├── context.yaml
│   ├── values-apim.yaml
│   └── values-gko.yaml
└── standard/         # another variant
    ├── context.yaml
    └── ...
```

With the tree above, setting `context: gravitee.io/apim` in your config is equivalent to `context: gravitee.io/apim/db-less` — sew reads `.default`, finds `db-less`, and resolves `gravitee.io/apim/db-less`.

To create a default for your own product, add a `.default` file next to the variant directories:

```bash
echo "db-less" > registry/gravitee.io/apim/.default
```

If neither `context.yaml` nor `.default` is found at the given path, sew returns an error.

## Image mirrors

sew can run local pull-through mirror proxies for container registries. When enabled, images pulled inside the Kind cluster are cached on the host — subsequent cluster recreations reuse the cached layers instead of downloading from the internet.

### How it works

Each upstream registry gets its own `registry:2` container running as a pull-through cache, bound to a local port (5000, 5001, …). The Kind node's containerd is configured with `hosts.toml` files that redirect pulls through these local mirrors. The mirror containers use a `restart: unless-stopped` policy, so they survive `sew down` and keep their cache across cluster lifecycles.

### Enabling mirrors

Add an `images.mirrors` section to your `sew.yaml`:

```yaml
# Mirror docker.io only (always implicit)
images:
  mirrors: {}
```

```yaml
# Mirror docker.io + a private registry
images:
  mirrors:
    upstreams:
    - acme.example.com
```

`docker.io` is always included — you don't need to list it explicitly.

### Options

| Field | Default | Description |
|-------|---------|-------------|
| `upstreams` | *(none)* | Additional registries to mirror (on top of `docker.io`) |
| `data` | `$SEW_HOME/mirrors` | Directory for cached layers and containerd host configs |

### Using images from a private registry

Combine mirrors with component overrides to pull images from a private registry through the local cache. Create a values file with the private image coordinates:

```yaml
# values-private.yaml
gateway:
  image:
    repository: acme.example.com/my-gateway
    tag: 1.0.0
```

Then reference it in your `sew.yaml`:

```yaml
registry: https://my-registry.example.com
context: org/product/variant

images:
  mirrors:
    upstreams:
    - acme.example.com

overrides:
  apim:
    helm:
      values:
      - values-private.yaml
```

The mirror proxy caches layers from `acme.example.com` locally, and the override swaps the image without modifying the upstream context.
