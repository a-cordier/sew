---
title: "Getting Started"
weight: 1
type: docs
---

**sew** spins up local Kubernetes clusters and deploys ready-to-use applications from a **registry**. You point it at one or more contexts (e.g. `gravitee.io/apim/db-less`), and it creates a Kind cluster and installs the components defined there (Helm charts and Kubernetes manifests).

## Quick start

1. **Config** — Create a `sew.yaml` (or use the one in this repo) with at least:
   - `registry`: e.g. `file://./registry` for the local test registry, or an HTTP URL.
   - `from`: list of context paths, e.g. `[gravitee.io/apim/db-less]`.
   - `kind`: Kind cluster spec (name, nodes, port mappings).

2. **Install** — Install the `sew` binary:
   ```bash
   go install github.com/a-cordier/sew@latest
   ```

3. **Run** — Create the cluster (use `--config` to point to a custom config):
   ```bash
   sew create
   ```

4. **Tear down**:
   ```bash
   sew delete
   ```

## Commands

| Command | Description |
|--------|-------------|
| `sew create` | Create the Kind cluster (if missing) and install the context: add Helm repos, then install each component (Helm upgrade --install). If no registry/context is configured, only creates the cluster. |
| `sew patch` | Patch a running cluster by merging overrides from a file and upgrading only the affected components. See {{< ref "patch" >}}. |
| `sew delete` | Delete a cluster and clean up associated resources. Auto-selects the target from state files, or use `--name` to specify. See {{< ref "delete" >}}. |
| `sew setup dns` | One-time OS-level DNS routing so `*.sew.local` queries reach the local DNS server. Requires sudo. |
| `sew teardown dns` | Remove the OS-level DNS routing created by `sew setup dns`. |
| `sew status` | Show the status of the current sew environment: cluster info, enabled features, load balancers, and DNS records. |
| `sew refresh dns` | Re-collect DNS records from the running cluster (picks up Gateways and Services created after `sew create`). |

## Global flags

- `--config <path>` — Project-level config file to merge on top of the user-level base (`$SEW_HOME/sew.yaml`). Defaults to `./sew.yaml` when present.
- `--registry <url>` — Registry URL (e.g. `file://./registry` or `https://…`). Overrides the value from the config file.
- `--from <path>` — Context path to compose (e.g. `gravitee.io/apim/db-less`). Repeatable: use multiple `--from` flags to compose several contexts. Overrides the `from` list from the config file.
