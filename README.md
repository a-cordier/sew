<p align="center"><img src="logo.svg" alt="sew" width="140"/><br><em>Your local Kubernetes tailor shop</em></p>

**sew** spins up local Kubernetes clusters and deploys ready-to-use applications from a **registry** and **context**. You point it at a context (e.g. `gravitee.io/apim/db-less`), and it creates a Kind cluster and installs the components defined there (Helm charts, and in the future manifests or Kustomize).

## Concepts

- **Registry** — A tree of context directories, either on the filesystem (`file:///path`) or over HTTP. The binary does not ship a registry; you use your own or a remote one.
- **Context** — A path inside the registry following `org/product/variant` (e.g. `gravitee.io/apim/db-less`). Each context has a `sew.yaml` that lists Helm repos and components (charts + values). If you omit the variant (e.g. `gravitee.io/apim`), sew looks for a `.default` file to pick one automatically (see [Default variant resolution](#default-variant-resolution)).
- **Config** — Configuration is layered. A **user-level** config at `$SEW_HOME/sew.yaml` (default `~/.sew/sew.yaml`) provides base settings; a **project-level** `./sew.yaml` (or `--config`) is merged on top. Each layer sets the registry URL, the context to use, the Kind cluster definition, and optional local components and repos. Set the `SEW_HOME` environment variable to change the user-level config location.

## Commands

| Command | Description |
|--------|-------------|
| `sew create` | Create the Kind cluster (if missing) and install the context: add Helm repos, then install each component (Helm upgrade --install). If no registry/context is configured, only creates the cluster. |
| `sew delete` | Delete the Kind cluster defined in the config. |
| `sew setup dns` | One-time OS-level DNS routing so `*.sew.local` queries reach the local DNS server. Requires sudo. |
| `sew teardown dns` | Remove the OS-level DNS routing created by `sew setup dns`. |
| `sew refresh dns` | Re-collect DNS records from the running cluster (picks up Gateways and Services created after `sew create`). |

### Global flags

- `--config <path>` — Project-level config file to merge on top of the user-level base (`$SEW_HOME/sew.yaml`). Defaults to `./sew.yaml` when present.
- `--registry <url>` — Registry URL (e.g. `file://./registry` or `https://…`). Overrides the value from the config file.
- `--context <path>` — Context path (e.g. `gravitee.io/apim/db-less`). Overrides the value from the config file.

## Quick start

1. **Config** — Create a `sew.yaml` (or use the one in this repo) with at least:
   - `registry`: e.g. `file://./registry` for the local test registry, or an HTTP URL.
   - `context`: e.g. `gravitee.io/apim/db-less`.
   - `kind`: Kind cluster spec (name, nodes, port mappings).

2. **Run** — From the `sew` directory (or with `--config` pointing to this config):
   ```bash
   go run . create
   ```
   Or build and run:
   ```bash
   go build -o sew .
   ./sew create
   ```

3. **Tear down**:
   ```bash
   ./sew delete
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

# Optional: add Helm repos required by local components
repos:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami

# Optional: override context components or add new ones
components:
  - name: apim          # matches a context component → merged
    helm:
      version: "4.10.0"
      valueFiles:
        - ./my-overrides.yaml
  - name: redis          # no match in context → added as a new component
    namespace: gravitee
    helm:
      chart: bitnami/redis
      values:
        architecture: standalone
```

Value file paths under `components.*.helm.valueFiles` are resolved relative to the directory containing the config file.

## Local components

Beyond overriding fields on components defined by the context, you can declare entirely new components and Helm repos in your `sew.yaml`. This is useful when you need supporting services (databases, caches, message brokers, …) that are not part of the upstream context.

### Adding a component

List the component under `components`. If its `name` does not match any component from the context, sew appends it as a new component and installs it alongside the context ones:

```yaml
components:
  - name: redis
    namespace: gravitee
    helm:
      chart: bitnami/redis
      values:
        architecture: standalone
```

### Adding Helm repos

If the new component's chart comes from a repo that the context does not declare, add it under `repos`:

```yaml
repos:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
```

Local repos are merged with context repos. When both lists contain a repo with the same name, the local entry wins.

### Adding dependencies between components

You can make a context component wait for a local component (or vice-versa) by adding `requires` entries. Requirements are deduplicated by component name:

```yaml
components:
  - name: apim
    requires:
      - component: redis
        conditions:
          ready: true
        selector:
          matchLabels:
            app.kubernetes.io/instance: redis
```

### Merge rules

When a local component matches a context component by name, the following merge rules apply:

| Field | Behaviour |
|-------|-----------|
| `requires` | Local requirements are appended (deduplicated by component name) |
| `helm.chart` | Local wins if non-empty |
| `helm.version` | Local wins if non-empty |
| `helm.valueFiles` | Local files are appended (higher precedence in Helm) |
| `helm.values` | Local values are merged on top of context values |

When there is no name match, the component is added to the deployment as-is.

## Context format

A context lives at `{registry}/{context_path}/` and must contain `sew.yaml`:

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

## Context composition

A context can inherit from another context by declaring `context` (and optionally `registry`) in its `sew.yaml`. The parent context is resolved first, then the child's overrides are merged on top using the same merge rules as user-level overrides.

```yaml
# registry/org/product/custom/sew.yaml
context: org/product/base

kind:
  name: custom-cluster

components:
  - name: app
    helm:
      values:
        debug: true
  - name: extra
    helm:
      chart: extra/chart
```

This says: start from the `org/product/base` context, override the cluster name, tweak the `app` component's values, and add an `extra` component.

### Same-registry and cross-registry

When only `context` is set, the parent is resolved from the same registry. To inherit from a context in a different registry, set `registry` as well:

```yaml
registry: https://other-registry.example.com
context: org/product/base

components:
  - name: addon
    helm:
      chart: addon/chart
```

Relative `file://` paths are resolved relative to the child context's directory (e.g. `file://../..` navigates up from the child).

### Multi-level composition

Composition chains work to arbitrary depth (grandparent → parent → child). Cycle detection prevents infinite loops — sew tracks visited `(registry, context)` pairs and errors if a cycle is found.

## Default variant resolution

A context path usually includes the variant (`org/product/variant`), but you can also point to the product level and let sew pick the default variant automatically.

When the resolved path has no `sew.yaml`, sew looks for a **`.default`** file in the same directory. This plain-text dotfile contains a single line — the name of the variant to use. sew then appends that variant to the path and resolves again.

```
registry/gravitee.io/apim/
├── .default          # contains "db-less"
├── db-less/
│   ├── sew.yaml
│   ├── values-apim.yaml
│   └── values-gko.yaml
└── standard/         # another variant
    ├── sew.yaml
    └── ...
```

With the tree above, setting `context: gravitee.io/apim` in your config is equivalent to `context: gravitee.io/apim/db-less` — sew reads `.default`, finds `db-less`, and resolves `gravitee.io/apim/db-less`.

To create a default for your own product, add a `.default` file next to the variant directories:

```bash
echo "db-less" > registry/gravitee.io/apim/.default
```

If neither `sew.yaml` nor `.default` is found at the given path, sew returns an error.

## Image mirrors

sew can run local pull-through mirror proxies for container registries. When enabled, images pulled inside the Kind cluster are cached on the host — subsequent cluster recreations reuse the cached layers instead of downloading from the internet.

### How it works

Each upstream registry gets its own `registry:2` container running as a pull-through cache, bound to a local port (5000, 5001, …). The Kind node's containerd is configured with `hosts.toml` files that redirect pulls through these local mirrors. The mirror containers use a `restart: unless-stopped` policy, so they survive `sew delete` and keep their cache across cluster lifecycles.

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

Combine mirrors with local component overrides to pull images from a private registry through the local cache. Create a values file with the private image coordinates:

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

components:
  - name: apim
    helm:
      valueFiles:
      - values-private.yaml
```

The mirror proxy caches layers from `acme.example.com` locally, and the component override swaps the image without modifying the upstream context.

## DNS

sew can run a local DNS server that lets you reach services inside the Kind cluster by hostname (e.g. `api.sew.local`) instead of looking up IPs manually. When enabled, `sew create` collects DNS records from the cluster, starts a background DNS server, and keeps it running across cluster lifecycles.

### How it works

1. **Record collection** — After all components are installed, sew introspects the cluster for hostnames:
   - **Gateway API**: Polls `Gateway` resources until they receive an IP address from the load balancer controller, then maps each `HTTPRoute` hostname to its parent Gateway IP.
   - **Static records**: Maps user-defined hostnames to LoadBalancer Service IPs (see [Static records](#static-records) below).

   Records are written to `$SEW_HOME/dns/<cluster-name>.json`.

2. **DNS server** — A lightweight DNS server (`sew dns serve`, started automatically in the background) serves A queries for `*.<domain>` from the collected record files. It watches the record directory and hot-reloads when files change. When the last record file is removed (i.e. all clusters are stopped), the server shuts down automatically.

3. **OS routing** — A one-time `sew setup dns` command configures the operating system to forward queries for the sew domain to the local server. This step requires sudo but only needs to be done once.

### Enabling DNS

Add `features.dns` to your `sew.yaml`:

```yaml
features:
  dns:
    enabled: true
```

On the next `sew create`, sew will collect records, start the DNS server, and print a reminder if OS routing is not yet configured.

### One-time OS setup

Run once after enabling DNS:

```bash
sew setup dns
```

This configures the OS to route `*.<domain>` queries to the local DNS server:

- **macOS**: creates `/etc/resolver/<domain>` (persists across reboots).
- **Linux**: configures `systemd-resolved` on the loopback interface (runtime only — lost on reboot).

To undo:

```bash
sew teardown dns
```

After setup, `sew create` and `sew delete` run without sudo.

### Refreshing records

If you deploy additional Gateways or LoadBalancer Services after `sew create`, their hostnames won't be in the DNS automatically. Re-collect them with:

```bash
sew refresh dns
```

The running DNS server picks up the updated records immediately.

### Static records

By default, sew discovers hostnames from Gateway API resources (`Gateway` + `HTTPRoute`). You can also map hostnames to LoadBalancer Services explicitly:

```yaml
features:
  dns:
    enabled: true
    records:
      - hostname: api.sew.local
        service: my-api-gateway
        namespace: gravitee
```

Each record resolves `hostname` to the external IP assigned to the named Service. Static records are collected alongside Gateway-derived ones during `sew create` and `sew refresh dns`.

### Options

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `false` | Enable the DNS feature. |
| `domain` | `sew.local` | Domain suffix served by the local DNS server. |
| `port` | `15353` | UDP port the DNS server listens on. |
| `records` | *(none)* | Static hostname-to-Service mappings (see above). |

### Multiple clusters

Each cluster writes its own record file (`<cluster-name>.json`). The DNS server merges records from all files, so hostnames from multiple clusters are resolvable simultaneously. When a cluster is stopped, only its records are removed.
