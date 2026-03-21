---
title: "Networking"
weight: 3
type: docs
---

sew gives you production-like networking on your local machine: load balancers, Gateway API support, and automatic DNS resolution. Say goodbye to `/etc/hosts` edits and IP hunting -- services are reachable by name out of the box.

## Load balancers

By default, `LoadBalancer`-type Services in Kind stay in `Pending` state because there's no cloud provider to assign IPs. sew can emulate this with a local cloud provider controller that assigns real IPs from the Docker network range.

Enable it in your `sew.yaml`:

```yaml
features:
  lb:
    enabled: true
```

Once enabled, any Service of type `LoadBalancer` in your cluster gets an external IP that's reachable from your host machine.

> **macOS note:** On macOS, Docker runs inside a lightweight VM, so container networks are not directly routable from the host. sew sets up a packet tunnel to bridge this gap, which requires `sudo` privileges. You will be prompted for your password when creating a cluster with load balancers enabled.

## Gateway API

sew supports the [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/) for managing ingress traffic. Enabling it automatically enables load balancers too (Gateway controllers need them):

```yaml
features:
  gateway:
    enabled: true
    channel: standard    # or "experimental"
```

This installs the Gateway API CRDs so you can define `Gateway` and `HTTPRoute` resources in your contexts or local manifests.

## Local DNS

sew can run a local DNS server that lets you reach services by hostname (e.g. `api.sew.local`) instead of looking up IPs manually.

### How it works

1. **Record collection** -- After all components are installed, sew introspects the cluster for hostnames. It discovers routes from Gateway API resources (`Gateway` + `HTTPRoute`) and resolves static records you define manually.
2. **DNS server** -- A lightweight DNS server runs in the background, serving A queries for `*.<domain>`. It watches for changes and hot-reloads when records are updated.
3. **OS routing** -- A one-time `sew setup dns` command tells your operating system to forward queries for the sew domain to the local server.

### Enabling DNS

```yaml
features:
  dns:
    enabled: true
```

On the next `sew create`, sew collects records, starts the DNS server, and prints a reminder if OS routing isn't configured yet.

### One-time OS setup

Run this once after enabling DNS:

```bash
sew setup dns
```

This configures your OS to route `*.sew.local` queries to the local DNS server:

- **macOS**: creates `/etc/resolver/sew.local` (persists across reboots)
- **Linux**: configures `systemd-resolved` on the loopback interface (runtime only)

To undo: `sew teardown dns`. After setup, `sew create` and `sew delete` run without sudo.

### Static records

By default, sew discovers hostnames from Gateway API resources. You can also map hostnames to `LoadBalancer` Services explicitly:

```yaml
features:
  dns:
    enabled: true
    records:
      - hostname: api.sew.local
        service: my-api-gateway
        namespace: default
```

### Wildcard records

Wildcard hostnames are supported per [RFC 4592](https://datatracker.ietf.org/doc/html/rfc4592). A `*` first label matches any hostname sharing the remaining suffix:

```yaml
features:
  dns:
    enabled: true
    records:
      - hostname: "*.api.sew.local"
        service: my-gateway
        namespace: default
```

Both `demo.api.sew.local` and `v2.demo.api.sew.local` resolve to the same Service IP. Exact records always take priority over wildcards.

### Refreshing records

If you deploy additional Gateways or Services after `sew create`, re-collect their hostnames with:

```bash
sew refresh dns
```

### DNS options

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `false` | Enable the DNS feature |
| `domain` | `sew.local` | Domain suffix served by the local DNS server |
| `port` | `15353` | UDP port the DNS server listens on |
| `records` | *(none)* | Static hostname-to-Service mappings |

### Multiple clusters

Each cluster writes its own record file. The DNS server merges records from all clusters, so hostnames from multiple environments are resolvable simultaneously. When a cluster is deleted, only its records are removed.

## Kind port mappings

Port mappings in your `sew.yaml` are merged with those from the context using union semantics, keyed by `(containerPort, protocol)`:

- Ports only in the context are preserved
- Ports only in your config are added
- When both define the same key, your entry wins

```yaml
kind:
  nodes:
    - role: control-plane
      extraPortMappings:
        - containerPort: 9090
          hostPort: 9090
```

This adds port 9090 alongside any ports the context already defines.
