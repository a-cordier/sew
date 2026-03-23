---
title: "Commands"
weight: 2
type: docs
---

All sew commands in one place. Every command respects the global flags listed at the bottom of this page.

## sew create

Create a Kind cluster and deploy the components defined by your context.

```bash
sew create
```

sew resolves the config chain (user-level `$SEW_HOME/sew.yaml` merged with your project `sew.yaml`), fetches contexts from the registry, creates the Kind cluster, and installs components in dependency order. If no registry or context is configured, only the bare cluster is created.

When features like load balancers, Gateway API, or DNS are enabled, sew sets them up automatically after the cluster is ready.

### Flags

| Flag | Description |
|------|-------------|
| `--config <path>` | Project-level config file. Defaults to `./sew.yaml` when present. |
| `--registry <url>` | Registry URL. Overrides the value from config. |
| `--from <path>` | Context path to compose. Repeatable. Overrides the `from` list from config. |

### Context flags

Contexts can define optional flags that customize the deployment. These are extra `--flag-name` options defined by the context maintainer as `sew--{flag-name}.yaml` patch files.

```bash
sew create --from gravitee.io/apim --no-portal --no-ui --no-es
```

Each flag merges a patch on top of the resolved context before deployment. Flags are cumulative -- you can combine as many as needed. Passing an unknown flag produces an error listing the available flags for that context.

See [Context Format -- Context flags]({{< ref "/docs/reference/context-format#context-flags" >}}) for how to author flag files.

## sew patch

Upgrade components on a running cluster by merging a patch file into the resolved context and re-deploying only the affected components. This is useful for testing upgrades (bumping image tags or chart versions), toggling feature flags, or applying configuration tweaks.

```bash
sew patch upgrade.yaml
sew patch upgrade.yaml --name my-cluster
sew patch upgrade.yaml --dry-run
```

### How it works

1. **Resolve context** -- sew loads the config chain and resolves the registry context, exactly as `sew create` does.
2. **Verify cluster** -- sew checks that the target Kind cluster is running.
3. **Load patch file** -- the patch file is loaded (same format as `sew.yaml`).
4. **Merge** -- patch components are merged on top of the resolved context using the standard [merge rules]({{< ref "/docs/guides/composing-contexts#merge-rules" >}}).
5. **Upgrade** -- only components named in the patch file are upgraded (Helm upgrade / kubectl apply), in dependency order.
6. **Readiness** -- sew waits for patched components that have `conditions.ready: true`.

Components not mentioned in the patch file are left untouched.

### Patch file format

The patch file uses the same format as `sew.yaml`. Only `components`, `helm.repos`, and `images.preload` are relevant -- other fields are ignored.

#### Upgrade image tags

```yaml
components:
  - name: my-app
    helm:
      values:
        gateway:
          image:
            tag: "2.1.0"
        api:
          image:
            tag: "2.1.0"
```

#### Change a chart version

```yaml
components:
  - name: my-app
    helm:
      version: "2.1.0"
```

#### Add a value file overlay

```yaml
components:
  - name: my-app
    helm:
      valueFiles:
        - ./staging-overrides.yaml
```

Value file paths are resolved relative to the directory containing the patch file.

#### Patch a k8s manifest component

```yaml
components:
  - name: custom-routes
    k8s:
      manifestFiles:
        - ./updated-routes.yaml
```

### Typical workflow

```bash
# Create the cluster with the current version
sew create

# Run tests against the current version
./run-tests.sh

# Patch: upgrade to the new version
sew patch upgrade.yaml

# Run tests against the new version
./run-tests.sh

# Tear down
sew delete
```

### Image preloading

When the cluster was created with image preloading enabled, the Kind nodes are already configured to pull from the local `sew-preload` registry. Adding `images.preload.refs` to your patch file pre-stages images before upgrading, so pods start faster:

```yaml
images:
  preload:
    refs:
      - myrepo/my-app:2.1.0-rc1
      - myrepo/my-api:2.1.0-rc1

components:
  - name: my-app
    helm:
      values:
        image:
          tag: 2.1.0-rc1
```

If no preload registry is running, sew prints a warning and proceeds normally.

### Dry-run mode

Use `--dry-run` to preview what a patch would change without applying anything:

```bash
sew patch upgrade.yaml --dry-run
```

When `--dry-run` is active:

- **Helm components** run `helm upgrade --dry-run=server` -- the chart is rendered and validated by the API server without creating the release
- **Kubernetes manifest components** run `kubectl apply --dry-run=server` -- objects are validated without persisting
- **Readiness checks are skipped** since no resources are actually deployed
- **Colored diff output** shows exactly what would change: added lines in green, removed lines in red

This is especially useful in CI pipelines:

```bash
sew patch upgrade.yaml --dry-run
sew patch upgrade.yaml
```

### Flags

| Flag | Description |
|------|-------------|
| `--name <cluster>` | Name of the cluster to patch. Defaults to `kind.name` from the resolved config. |
| `--dry-run` | Preview changes without applying. Uses server-side dry-run for both Helm and Kubernetes resources. |

## sew delete

Tear down a cluster and clean up all associated resources: the Kind cluster, load balancer containers, DNS records, mirror proxies, and the preload registry.

```bash
sew delete
sew delete --name my-cluster
```

### How it works

1. **Resolve target** -- sew determines which cluster to delete (see [Target resolution](#target-resolution) below).
2. **DNS records** -- Removes the cluster's DNS record file.
3. **Load balancer containers** -- Stops and removes Docker containers created by the cloud provider controller.
4. **Kind cluster** -- Deletes the Kind cluster, removing all namespaces, Helm releases, and applied manifests.
5. **Image mirrors** -- Stops mirror proxy containers if mirrors were configured.
6. **Preload registry** -- Stops the preload registry container if preloading was configured.
7. **Background processes** -- Stops the cloud provider controller and DNS server when no clusters or DNS records remain.
8. **State file** -- Removes the cluster's state file from `$SEW_HOME/clusters/`.

### Target resolution

`sew delete` doesn't need the original `sew.yaml` or registry to be available. It uses state files to find and clean up the cluster. See [Directory Layout -- clusters/]({{< ref "/docs/reference/directory-layout#clusters" >}}) for how target resolution and best-effort cleanup work.

### Flags

| Flag | Description |
|------|-------------|
| `--name <cluster>` | Name of the cluster to delete. When omitted, sew auto-selects from state files or falls back to config. |

## sew status

Show the status of your current sew environment at a glance: cluster name, enabled features, active load balancers, and DNS state.

```bash
sew status
```

### What it shows

**Cluster** -- The cluster name from the resolved config.

**Features** -- Whether load balancers, Gateway API, and DNS are enabled. For Gateway API, shows the channel (`standard` or `experimental`). For DNS, shows the domain and port.

**Load Balancers** -- Lists active load balancer containers and their IPs. sew queries Docker directly for containers associated with the cluster, so this reflects the actual running state even if the original config is no longer available.

**DNS** -- Three pieces of information:
- **Resolver**: whether OS-level DNS routing is configured (i.e. whether `sew setup dns` has been run)
- **Server**: whether the local DNS server process is running
- **Records**: all registered hostname-to-IP mappings, grouped by cluster

### Example output

```
Cluster
  Name: gio-apim

Features
  lb:      enabled
  gateway: disabled
  dns:     enabled (domain: sew.local, port: 15353)

Load Balancers
  sew-lb-gio-apim-1 → 172.18.0.5

DNS
  resolver: configured for sew.local
  server:   running on 127.0.0.1:15353
  records:
    console.sew.local → 172.18.0.5 (gio-apim)
    gateway.sew.local → 172.18.0.5 (gio-apim)
```

## sew validate

Validate `sew.yaml` and context flag files (`sew--*.yaml`) against the configuration schema. Catches typos, unknown fields, and type mismatches before you deploy.

```bash
sew validate
sew validate registry/kafka/standalone/sew.yaml
sew validate registry/
```

When given a directory, sew walks it recursively and validates every `sew.yaml` and `sew--*.yaml` file it finds. Context flag files are additionally checked for a valid naming convention and a non-empty `description` field. When no argument is given, it validates `./sew.yaml` in the current directory.

Exit code is non-zero when any file fails validation, making it suitable for CI pipelines and pre-commit checks.

### Flags

This command has no additional flags beyond the [global flags](#global-flags).

## sew setup dns

Configure your operating system to forward `*.sew.local` queries to the local DNS server. This is a one-time setup that requires sudo.

```bash
sew setup dns
```

- **macOS**: creates `/etc/resolver/sew.local` (persists across reboots)
- **Linux**: configures `systemd-resolved` on the loopback interface (runtime only)

After this, `sew create` and `sew delete` run without sudo.

## sew teardown dns

Remove the OS-level DNS routing created by `sew setup dns`.

```bash
sew teardown dns
```

## sew refresh dns

Re-collect DNS records from the running cluster. Use this after deploying additional Gateways or LoadBalancer Services that weren't present during `sew create`.

```bash
sew refresh dns
```

The running DNS server picks up the updated records immediately.

## Global flags

These flags are available on all commands:

| Flag | Description |
|------|-------------|
| `--config <path>` | Project-level config file to merge on top of the user-level base (`$SEW_HOME/sew.yaml`). Defaults to `./sew.yaml` when present. |
| `--registry <url>` | Registry URL (e.g. `file://./registry` or `https://…`). Overrides the value from config. |
| `--from <path>` | Context path to compose (e.g. `elastic/elasticsearch/standalone`). Repeatable. Overrides the `from` list from config. |
