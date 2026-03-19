---
title: "Patch"
weight: 6
type: docs
---

`sew patch` upgrades components on a running cluster by merging a partial configuration file into the current resolved context and re-deploying only the affected components.

This is useful for testing upgrades (e.g. bumping image tags), toggling feature flags, or applying configuration tweaks without tearing down and recreating the entire cluster.

## How it works

1. **Resolve context** — sew loads the config chain and resolves the registry context, exactly as `sew create` does.
2. **Verify cluster** — sew checks that the target Kind cluster is running.
3. **Load patch file** — the patch file is loaded (same format as `sew.yaml`).
4. **Merge** — patch components are merged on top of the resolved context using the same [merge rules]({{< ref "configuration#merge-rules" >}}) as local component overrides.
5. **Upgrade** — only components named in the patch file are upgraded (Helm upgrade / kubectl apply), in dependency order.
6. **Readiness** — sew waits for patched components that have `conditions.ready: true`.

Components not mentioned in the patch file are left untouched.

## Usage

```bash
sew patch upgrade.yaml
```

Patch a specific cluster by name:

```bash
sew patch upgrade.yaml --name gio-apim
```

## Patch file format

The patch file uses the same format as `sew.yaml`. Only the `components` and `helm.repos` sections are relevant — other fields are ignored.

### Example: upgrade image tags

```yaml
components:
  - name: apim
    helm:
      values:
        gateway:
          image:
            tag: "4.11.0"
        api:
          image:
            tag: "4.11.0"
        ui:
          image:
            tag: "4.11.0"
        portal:
          image:
            tag: "4.11.0"
```

### Example: change a chart version

```yaml
components:
  - name: apim
    helm:
      version: "4.11.0"
```

### Example: add a value file overlay

```yaml
components:
  - name: apim
    helm:
      valueFiles:
        - ./staging-overrides.yaml
```

Value file paths are resolved relative to the directory containing the patch file, just like in a regular `sew.yaml`.

### Example: patch a k8s manifest component

```yaml
components:
  - name: custom-routes
    k8s:
      manifestFiles:
        - ./updated-routes.yaml
```

## Typical workflow

A common use case is running an end-to-end test suite before and after an upgrade:

```bash
# 1. Create the cluster with the current version
sew create

# 2. Run e2e tests against the current version
./run-e2e-tests.sh

# 3. Patch: upgrade to the new version
sew patch upgrade.yaml

# 4. Run e2e tests against the new version
./run-e2e-tests.sh

# 5. Tear down
sew delete
```

## Dry-run mode

Use `--dry-run` to preview what a patch would change without applying anything to the cluster. Both Helm and Kubernetes manifest components support server-side dry-run: the API server validates the request and returns what would be created or modified, but nothing is persisted.

```bash
sew patch upgrade.yaml --dry-run
```

When `--dry-run` is active:

- **Helm components** — `helm install` / `helm upgrade` runs with `--dry-run=server`, so the chart is rendered and validated by the API server without creating or updating the release.
- **Kubernetes manifest components** — `kubectl apply` runs with `--dry-run=server` (`DryRun: ["All"]` in apply options), validating the objects without persisting them.
- **Readiness checks are skipped** — since no resources are actually deployed, sew does not wait for pods or other conditions.
- **Colored diff output** — after each component, sew prints a colored unified diff showing exactly what would change: added lines in green, removed lines in red, and hunk headers in cyan. For fresh installs the entire manifest is shown as additions.

This is especially useful in CI pipelines to catch configuration or chart errors before a real upgrade:

```bash
# Validate the patch against the live cluster
sew patch upgrade.yaml --dry-run

# If successful, apply for real
sew patch upgrade.yaml
```

## Flags

| Flag | Description |
|------|-------------|
| `--name <cluster>` | Name of the cluster to patch. When omitted, sew uses `kind.name` from the resolved config. |
| `--dry-run` | Show what would change without applying. Uses server-side dry-run for both Helm and Kubernetes resources. |
