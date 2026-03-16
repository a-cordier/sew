---
title: "DNS"
weight: 6
type: docs
---

sew can run a local DNS server that lets you reach services inside the Kind cluster by hostname (e.g. `api.sew.local`) instead of looking up IPs manually. When enabled, `sew create` collects DNS records from the cluster, starts a background DNS server, and keeps it running across cluster lifecycles.

## How it works

1. **Record collection** — After all components are installed, sew introspects the cluster for hostnames:
   - **Gateway API**: Polls `Gateway` resources until they receive an IP address from the load balancer controller, then maps each `HTTPRoute` hostname to its parent Gateway IP.
   - **Static records**: Maps user-defined hostnames to LoadBalancer Service IPs (see [Static records](#static-records) below).

   Records are written to `$SEW_HOME/dns/<cluster-name>.json`.

2. **DNS server** — A lightweight DNS server (`sew dns serve`, started automatically in the background) serves A queries for `*.<domain>` from the collected record files. It watches the record directory and hot-reloads when files change. When the last record file is removed (i.e. all clusters are stopped), the server shuts down automatically.

3. **OS routing** — A one-time `sew setup dns` command configures the operating system to forward queries for the sew domain to the local server. This step requires sudo but only needs to be done once.

## Enabling DNS

Add `features.dns` to your `sew.yaml`:

```yaml
features:
  dns:
    enabled: true
```

On the next `sew create`, sew will collect records, start the DNS server, and print a reminder if OS routing is not yet configured.

## One-time OS setup

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

## Refreshing records

If you deploy additional Gateways or LoadBalancer Services after `sew create`, their hostnames won't be in the DNS automatically. Re-collect them with:

```bash
sew refresh dns
```

The running DNS server picks up the updated records immediately.

## Static records

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

## Options

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `false` | Enable the DNS feature. |
| `domain` | `sew.local` | Domain suffix served by the local DNS server. |
| `port` | `15353` | UDP port the DNS server listens on. |
| `records` | *(none)* | Static hostname-to-Service mappings (see above). |

## Multiple clusters

Each cluster writes its own record file (`<cluster-name>.json`). The DNS server merges records from all files, so hostnames from multiple clusters are resolvable simultaneously. When a cluster is stopped, only its records are removed.
