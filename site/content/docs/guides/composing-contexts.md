---
title: "Composing Contexts"
weight: 2
type: docs
---

One of sew's key strengths is composition. You can layer contexts together using `from` to build complex stacks from simple, reusable building blocks -- without duplicating configuration. The [Architecture]({{< ref "/docs/reference/architecture#context-composition" >}}) page illustrates how this works visually.

## The basics

The `from` field lists registry paths to compose. Each context is resolved and merged in order, with your local overrides applied last:

```yaml
registry: https://raw.githubusercontent.com/a-cordier/sew/refs/heads/main/registry
from:
  - elastic/elasticsearch/standalone

kind:
  name: my-cluster

components:
  - name: elasticsearch
    namespace: my-app
```

This says: start from the `elastic/elasticsearch/standalone` context, rename the cluster, and move Elasticsearch into a different namespace.

## Multi-context composition

You can compose multiple independent contexts into a single stack. This is how you assemble real-world environments from reusable pieces:

```yaml
from:
  - mongodb/standalone
  - elastic/elasticsearch/standalone

kind:
  name: my-stack

components:
  - name: mongodb
    namespace: my-app
  - name: elasticsearch
    namespace: my-app
  - name: my-service
    type: helm
    namespace: my-app
    requires:
      - component: mongodb
      - component: elasticsearch
    helm:
      chart: myrepo/my-service
```

Contexts in `from` are merged left-to-right: later entries override earlier ones on conflicts. Your local fields override last.

## Abstract contexts

When several variants share a common foundation, extract the shared parts into an **abstract** context. Mark it with `abstract: true` -- it can't be deployed on its own, only composed into concrete contexts:

```yaml
# registry/mycompany/myproduct/base/sew.yaml
abstract: true

helm:
  repos:
    - name: myrepo
      url: https://charts.example.com

components:
  - name: app
    namespace: default
    helm:
      chart: myrepo/app
      version: "2.0.0"
      values:
        replicas: 1
```

Concrete variants compose from the abstract base:

```yaml
# registry/mycompany/myproduct/dev/sew.yaml
from:
  - mycompany/myproduct/base

kind:
  name: dev-cluster

components:
  - name: app
    helm:
      values:
        debug: true
```

## Default variant resolution

When a product has multiple variants, the registry can define a **default** so you don't have to spell out the full path. A `.default` file in a directory contains the name of the variant to use:

```
registry/mycompany/myproduct/
├── .default          # contains "dev"
├── dev/
│   └── sew.yaml
└── staging/
    └── sew.yaml
```

With this setup, `from: [mycompany/myproduct]` resolves to `mycompany/myproduct/dev`.

Defaults chain across multiple levels -- sew reads `.default` at each directory until it finds a `sew.yaml`. For example, `from: [elastic]` resolves first to `elastic/elasticsearch` (via `elastic/.default`), then to `elastic/elasticsearch/standalone` (via `elastic/elasticsearch/.default`), where the actual `sew.yaml` lives.

## Config resolution order

When you run `sew create`, sew builds the final configuration by merging multiple layers. Each layer overrides the one before it:

- **User-level base** (`$SEW_HOME/sew.yaml`, defaults to `~/.sew/sew.yaml`) -- Shared settings across all your projects. Use this for things like a custom registry URL, image mirrors, or a default DNS domain. This file is optional.
- **Project-level** (`./sew.yaml` or the path given with `--config`) -- Your project's specific config. This is where you list `from` entries, add components, and set cluster options.
- **Registry contexts** -- Each entry in `from` is fetched and merged left-to-right. Later contexts override earlier ones on conflicts.
- **Embedded defaults** -- sew fills in any remaining gaps with sensible defaults (cluster name, ports, feature flags).

The `--registry` and `--from` CLI flags override the corresponding values from config files, so you can quickly test a different context without editing your `sew.yaml`.

## Local overrides

Beyond composing registry contexts, you can add your own components and Helm repos directly in your project `sew.yaml`. This is useful for supporting services that aren't part of the upstream context:

```yaml
from:
  - mycompany/myproduct/dev

helm:
  repos:
    - name: bitnami
      url: https://charts.bitnami.com/bitnami

components:
  - name: redis
    namespace: my-app
    helm:
      chart: bitnami/redis
      values:
        architecture: standalone
```

If a component name matches one from the context, your values are merged on top. If there's no match, the component is added as a new deployment.

### Using value files

For large overrides, you can use `valueFiles` instead of (or alongside) inline `values`. Paths are resolved relative to the `sew.yaml` directory:

```yaml
components:
  - name: app
    helm:
      valueFiles:
        - values-dev.yaml
      values:
        debug: true
```

Value files from composed contexts are appended in order, with later files taking higher precedence. Inline `values` are merged on top of everything.

### Kubernetes manifest components

You can also deploy plain Kubernetes resources without a Helm chart by setting `type: k8s`:

```yaml
components:
  - name: routes
    type: k8s
    namespace: my-app
    k8s:
      manifestFiles:
        - gateway.yaml
      manifests:
        - apiVersion: v1
          kind: Service
          metadata:
            name: my-service
          spec:
            type: ClusterIP
            ports:
              - port: 8080
                targetPort: 8080
            selector:
              app: my-service
```

For larger manifests, you can use `manifestFiles` to reference external YAML files instead of inlining them. Paths are resolved relative to the `sew.yaml` directory:

```yaml
components:
  - name: routes
    type: k8s
    namespace: my-app
    k8s:
      manifestFiles:
        - gateway.yaml
        - routes.yaml
```

You can combine `manifestFiles` and inline `manifests` in the same component -- both are applied.

### Local secrets and ConfigMaps

A `k8s` component can create Secrets and ConfigMaps from local files or environment variables:

```yaml
components:
  - name: credentials
    type: k8s
    namespace: my-app
    k8s:
      secrets:
        - name: license-key
          fromFile: ./license.key
          onMissing: ignore
        - name: api-credentials
          entries:
            - key: token
              fromFile: ./token.txt
            - key: API_KEY
              fromEnv: MY_API_KEY
      configMaps:
        - name: logging-config
          entries:
            - key: logback.xml
              fromFile: ./logback.xml
```

The `onMissing` field controls behavior when a source file or env var is missing: `fail` (default) aborts the deployment, `ignore` skips the resource with a warning.

## Dependencies between components

Use `requires` to express inter-component dependencies. sew installs components in dependency order and can wait for readiness:

```yaml
components:
  - name: my-service
    requires:
      - component: mongodb
        conditions:
          ready: true
        selector:
          matchLabels:
            app.kubernetes.io/instance: mongodb
    conditions:
      ready: true
    timeout: 10m
```

## Merge rules

When composing contexts or applying local overrides, sew merges fields following these rules:

| Field | Behavior |
|-------|----------|
| `helm.chart` | Your value wins if non-empty |
| `helm.version` | Your value wins if non-empty |
| `helm.valueFiles` | Your files are appended (higher precedence in Helm) |
| `helm.values` | Deep-merged -- maps recurse, named lists merge by `name`, scalars replace |
| `k8s.manifestFiles` | Your files are appended |
| `k8s.manifests` | Union by resource identity; your version wins on conflict |
| `k8s.secrets` | Your secrets are appended |
| `k8s.configMaps` | Your configMaps are appended |
| `requires` | Your requirements are appended (deduplicated by component name) |
| `conditions` | Your value wins if `ready` is true |
| `selector` | Your value wins if set |
| `timeout` | Your value wins if non-empty |

### Values deep merge

When `helm.values` overlap on the same key, sew picks a strategy based on the value type:

| Value type | Strategy |
|-----------|----------|
| Maps | Recursive deep merge -- each nested key is merged individually |
| Named lists (objects with a `name` key) | Merge by `name` -- same-name entries are overridden, new entries appended |
| Everything else (scalars, plain lists) | Replace -- your value wins |

An empty list (`env: []`) replaces the parent's list entirely -- use this to clear inherited entries.

## Multi-level composition

Composition chains work to arbitrary depth. A grandparent context can be composed by a parent, which is then composed by your project config. sew tracks visited contexts and errors if it detects a cycle.

## Overriding service networking

When composing contexts, you might need to change how a child context exposes services. For Helm components, override the relevant values:

```yaml
components:
  - name: search-engine
    helm:
      values:
        service:
          type: ClusterIP
          nodePort: null    # clear the child's nodePort to avoid Kubernetes rejection
```

For `k8s` manifest components, provide a full replacement Service manifest -- manifests are merged by resource identity, so your Service replaces the child's entirely.
