---
title: "Directory Layout"
weight: 4
type: docs
---

sew stores all persistent data under a single home directory. By default this is `~/.sew`; override it by setting the `SEW_HOME` environment variable.

```
$SEW_HOME/
├── sew.yaml
├── clusters/
│   └── <cluster-name>.yaml
├── dns/
│   └── <cluster-name>.yaml
├── logs/
│   ├── delete.log
│   └── <context-path>/
│       ├── install.log
│       └── patch.log
├── mirrors/
│   ├── data/
│   └── hosts/
├── preload/
└── pids/
    ├── cpk.pid
    └── dns.pid
```

## `sew.yaml`

The user-level base config. Settings here are merged under every project config -- use it for things like a custom registry URL, image mirrors, or a default DNS domain. This file is optional; see [Config resolution order]({{< ref "/docs/guides/composing-contexts#config-resolution-order" >}}).

## `clusters/`

State files created automatically at the end of a successful `sew create`. Each file captures the cluster name, enabled features, and image configuration -- everything `sew delete` needs for a clean teardown.

```
clusters/
├── my-cluster.yaml
└── dev-stack.yaml
```

`sew delete` uses these files for [target resolution](#target-resolution). State files are removed after a successful delete.

### Target resolution

`sew delete` doesn't need the original `sew.yaml` or registry to be available. It resolves the target cluster in this order:

1. **`--name` flag** -- If provided, sew looks up the state file for that cluster directly. Works from any directory.
2. **State files** -- If `--name` is omitted, sew scans `clusters/` for existing state files. With one cluster, it's deleted automatically. With multiple clusters, you get an interactive prompt.
3. **Config fallback** -- If no state files exist, sew reads `kind.name` from the config chain.

If you pass `--name` but no state file exists, sew performs a best-effort cleanup (Kind cluster, load balancer containers, DNS records) and prints a warning that mirrors and preload cannot be stopped since the original configuration is unknown.

## `dns/`

DNS record files, one per cluster. Each file contains the hostname-to-IP mappings collected by `sew create` or `sew refresh dns`. The local DNS server watches this directory and hot-reloads when records change.

When a cluster is deleted, only its record file is removed -- records from other clusters remain resolvable.

## `logs/`

Log files from cluster operations. Helm and kubectl output that isn't shown in the terminal is captured here for debugging.

- **`delete.log`** -- Output from the most recent `sew delete`.
- **`<context-path>/install.log`** -- Output from `sew create` for a given context. The context path uses `_` as a separator (e.g. `gravitee.io_apim/install.log`).
- **`<context-path>/patch.log`** -- Output from `sew patch` for a given context.

When a component fails to install or times out, the log file usually contains the Helm or kubectl error that explains why. Check it first before diving into `kubectl` commands.

## `mirrors/`

Data and configuration for image mirror proxies.

- **`data/`** -- Cached image layers, persisted across cluster lifecycles. Mirror containers use a `restart: unless-stopped` policy, so cached data survives `sew delete`.
- **`hosts/`** -- Generated `hosts.toml` files that configure containerd on Kind nodes to pull through the local mirrors.

See [Container Images]({{< ref "/docs/guides/container-images" >}}) for how to enable mirrors.

## `preload/`

Cached image layers for the preload registry. Like `mirrors/`, this data persists across cluster lifecycles -- when you delete and recreate a cluster, previously pushed images are still available, and only changed layers need to be re-pushed.

See [Container Images -- Image preloading]({{< ref "/docs/guides/container-images#image-preloading" >}}) for details.

## `pids/`

PID files for background processes managed by sew:

- **`cpk.pid`** -- The cloud provider controller that assigns IPs to `LoadBalancer` Services.
- **`dns.pid`** -- The local DNS server.

These processes are started automatically by `sew create` and stopped by `sew delete` when no clusters or DNS records remain. You shouldn't need to manage them manually.
