---
title: "Delete"
weight: 7
type: docs
---

`sew delete` tears down a cluster and cleans up all associated resources: the Kind cluster itself, load balancer containers, DNS records, image mirror proxies, and the preload registry.

## How it works

1. **Resolve target** — sew determines which cluster to delete (see [Target resolution](#target-resolution) below).
2. **DNS records** — Removes the cluster's DNS record file (`$SEW_HOME/dns/<cluster>.json`).
3. **Load balancer containers** — Stops and removes Docker containers created by the cloud provider controller for this cluster.
4. **Kind cluster** — Deletes the Kind cluster, which removes all namespaces, Helm releases, and applied manifests in one step.
5. **Image mirrors** — Stops mirror proxy containers if the cluster state indicates mirrors were configured.
6. **Preload registry** — Stops the preload registry container if preloaded images were configured.
7. **Background processes** — Stops the cloud provider controller and DNS server when no Kind clusters or DNS records remain.
8. **State file** — Removes the cluster's state file from `$SEW_HOME/clusters/`.

## Target resolution

`sew delete` does not need the original `sew.yaml` or registry to be available. It resolves the target cluster in this order:

1. **`--name` flag** — If provided, sew looks up the state file for that cluster directly. Works from any directory.
2. **State files** — If `--name` is not given, sew scans `$SEW_HOME/clusters/` for existing state files:
   - **One cluster**: deletes it automatically.
   - **Multiple clusters**: shows an interactive prompt to pick one.
3. **Config fallback** — If no state files exist, sew reads `kind.name` from the `sew.yaml` config chain (same as `sew create`). This handles clusters created before the state file feature was introduced.

### State files

A state file is created automatically at the end of a successful `sew create`. It captures the cluster name, enabled features, image configuration, and any delete notes from the context — everything `sew delete` needs for a clean teardown.

State files live at `$SEW_HOME/clusters/<cluster-name>.yaml` and are removed after a successful delete.

## Usage

Delete the current (or only) cluster:

```bash
sew delete
```

Delete a specific cluster by name:

```bash
sew delete --name my-cluster
```

### Best-effort cleanup

If you pass `--name` but no state file exists for that cluster (e.g. the state was manually deleted or the cluster was created with an older version of sew), delete performs a best-effort cleanup:

- Deletes the Kind cluster
- Removes load balancer containers
- Removes DNS records

A warning is printed about image mirrors and preload not being stopped, since the original configuration is unknown.

## Flags

| Flag | Description |
|------|-------------|
| `--name <cluster>` | Name of the cluster to delete. When omitted, sew auto-selects from state files or falls back to config. |
