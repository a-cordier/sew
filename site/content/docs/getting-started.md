---
title: "Getting Started"
weight: 1
type: docs
---

**sew** spins up local Kubernetes clusters and deploys ready-to-use applications from a **registry**. You point it at one or more contexts (e.g. `gravitee.io/apim/db-less`), and it creates a Kind cluster and installs the components defined there (Helm charts, and in the future manifests or Kustomize).

## Quick start

1. **Config** — Create a `sew.yaml` (or use the one in this repo) with at least:
   - `registry`: e.g. `file://./registry` for the local test registry, or an HTTP URL.
   - `from`: list of context paths, e.g. `[gravitee.io/apim/db-less]`.
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

## Commands

| Command | Description |
|--------|-------------|
| `sew create` | Create the Kind cluster (if missing) and install the context: add Helm repos, then install each component (Helm upgrade --install). If no registry/context is configured, only creates the cluster. |
| `sew delete` | Delete the Kind cluster defined in the config. |
| `sew setup dns` | One-time OS-level DNS routing so `*.sew.local` queries reach the local DNS server. Requires sudo. |
| `sew teardown dns` | Remove the OS-level DNS routing created by `sew setup dns`. |
| `sew refresh dns` | Re-collect DNS records from the running cluster (picks up Gateways and Services created after `sew create`). |

## Global flags

- `--config <path>` — Project-level config file to merge on top of the user-level base (`$SEW_HOME/sew.yaml`). Defaults to `./sew.yaml` when present.
- `--registry <url>` — Registry URL (e.g. `file://./registry` or `https://…`). Overrides the value from the config file.
- `--from <path>` — Context path to compose (e.g. `gravitee.io/apim/db-less`). Repeatable: use multiple `--from` flags to compose several contexts. Overrides the `from` list from the config file.
