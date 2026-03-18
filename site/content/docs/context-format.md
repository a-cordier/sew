---
title: "Context Format"
weight: 4
type: docs
---

A context lives at `{registry}/{context_path}/` and must contain `sew.yaml`:

```yaml
helm:
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
      valueFiles:
        - values-apim.yaml
```

File paths in `valueFiles` are relative to the context directory. `type` defaults to `helm` if omitted.

## Kubernetes manifest components

Components can also deploy plain Kubernetes resources (no Helm chart required) by setting `type: k8s`. The `k8s` block supports two fields:

- **`manifests`** — a list of inline Kubernetes resource objects
- **`manifestFiles`** — a list of paths to YAML files containing one or more resources (multi-document YAML separated by `---` is supported)

### Inline manifests

```yaml
components:
  - name: mongodb
    type: k8s
    namespace: default
    k8s:
      manifests:
        - apiVersion: apps/v1
          kind: Deployment
          metadata:
            name: mongodb
          spec:
            replicas: 1
            selector:
              matchLabels:
                app: mongodb
            template:
              metadata:
                labels:
                  app: mongodb
              spec:
                containers:
                  - name: mongodb
                    image: mongo:7
                    ports:
                      - containerPort: 27017
        - apiVersion: v1
          kind: Service
          metadata:
            name: mongodb
          spec:
            type: ClusterIP
            ports:
              - port: 27017
                targetPort: 27017
            selector:
              app: mongodb
```

### External manifest files

```yaml
components:
  - name: routes
    type: k8s
    namespace: my-app
    k8s:
      manifestFiles:
        - gateway.yaml
        - httproutes.yaml
```

File paths in `manifestFiles` are relative to the context directory. Both fields can be used together in the same component — file-based resources are applied first, then inline manifests.

Kubernetes manifest components participate in the same dependency graph as Helm components: they can declare `requires` and be required by other components.

## Context composition

A context can compose other contexts by declaring `from` in its `sew.yaml`. The listed contexts are resolved left-to-right, merged into an accumulator, then the local overrides are applied on top using the same merge rules as user-level overrides.

Both workspace configs (users) and registry contexts (maintainers) use the same `from` field with identical semantics: *"this stack is built from these contexts, plus my local overrides."*

### Single parent

```yaml
# registry/org/product/custom/sew.yaml
from:
  - org/product/base

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

### Multi-context composition

A context can compose multiple independent contexts. This is useful for assembling a stack from reusable building blocks:

```yaml
# registry/gravitee.io/apim/aio/sew.yaml
from:
  - mongodb/standalone
  - elastic/elasticsearch

kind:
  name: gio-apim

components:
  - name: mongodb
    namespace: gravitee
  - name: elasticsearch
    namespace: gravitee
  - name: apim
    type: helm
    namespace: gravitee
    requires:
      - component: mongodb
      - component: elasticsearch
    helm:
      chart: graviteeio/apim
```

Contexts in `from` are merged left-to-right: later entries override earlier ones on conflicts. Local fields override last.

### Abstract contexts

A context can be marked as `abstract: true` to indicate that it is a shared base configuration not meant to be deployed on its own. Attempting to deploy an abstract context directly (via `sew create`) produces an error — it must be composed into a concrete context through `from`.

This is useful when several variants share a common foundation (repos, components, Kind settings) but differ in values or extra components. Extract the shared parts into an abstract context and let each variant compose from it:

```yaml
# registry/org/product/base/sew.yaml
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

Variants compose from the abstract base and add their own overrides:

```yaml
# registry/org/product/dev/sew.yaml
from:
  - org/product/base

kind:
  name: dev-cluster

components:
  - name: app
    helm:
      values:
        debug: true
```

```yaml
# registry/org/product/prod/sew.yaml
from:
  - org/product/base

kind:
  name: prod-cluster

components:
  - name: app
    helm:
      values:
        replicas: 3
```

Deploying `org/product/base` directly fails, but deploying `org/product/dev` or `org/product/prod` works — each inherits the base repos, components, and settings, then applies its own overrides. The resulting context is not abstract, even though its parent is.

### Registry organization

The registry follows an **org/product** convention:

```
registry/
  mongodb/
    .default              # -> standalone
    standalone/
      sew.yaml
  elastic/
    elasticsearch/
      sew.yaml
  postgresql/
    .default              # -> standalone
    standalone/
      sew.yaml
  gravitee.io/
    apim/
      .default            # -> aio
      aio/
        .default          # -> mongodb
        base/
          sew.yaml        # abstract: true -- common APIM config
        mongodb/
          sew.yaml        # from: [mongodb/standalone, elastic/elasticsearch, gravitee.io/apim/aio/base]
          notes.create
        postgres/
          sew.yaml        # from: [postgresql/standalone, elastic/elasticsearch, gravitee.io/apim/aio/base]
          notes.create
      dbless/
        sew.yaml
        notes.create
      gateway-api/
        sew.yaml
```

The `.default` chain `gravitee.io/apim` → `aio` → `mongodb` means existing configs with `from: [gravitee.io/apim]` resolve to the mongodb variant without changes. Swapping implementations is consumer choice — replace `mongodb` with `postgres` in your context path.

### Cross-registry composition

By default, `from` entries are resolved against the same registry. To compose from a different registry, set `registry`:

```yaml
registry: https://other-registry.example.com
from:
  - org/product/base

components:
  - name: addon
    helm:
      chart: addon/chart
```

Relative `file://` paths are resolved relative to the child context's directory (e.g. `file://../..` navigates up from the child).

### Multi-level composition

Composition chains work to arbitrary depth (grandparent → parent → child). Cycle detection prevents infinite loops — sew tracks visited `(registry, context)` pairs and errors if a cycle is found.

### Merge semantics

When contexts are composed, each top-level field is merged as follows:

- **`kind`** — Scalar fields (`name`, `apiVersion`, `kind`): child wins if set. `nodes`: child replaces the entire list; `extraPortMappings` are merged as a **union** keyed by `(containerPort, protocol)` — parent-only ports are preserved, child-only ports are added, and when both sides define the same key the child wins. `containerdConfigPatches`: child replaces entirely.
- **`components`** — Matched by name using the same rules as user-level overrides (see [Merge rules]({{< ref "configuration#merge-rules" >}})): `helm.chart` and `helm.version` child wins, `helm.valueFiles` appended, `helm.values` shallow-merged, `k8s.manifestFiles` appended, `k8s.manifests` appended, `k8s.secrets` appended, `k8s.configMaps` appended, `requires` appended and deduplicated. Unmatched components are appended.
- **`helm.repos`** — Deduplicated by name; child entry wins on conflict.
- **`features`** — Each feature block (`lb`, `gateway`, `dns`) is replaced as a whole if the child defines it; otherwise inherited from parent.
- **`images`** — `preload`: when both sides define it, `refs` are deduplicated (union); when only one side defines it, that side's config is used as-is. `mirrors` from the child wins when set; otherwise inherited from parent.

## Default variant resolution

A context path usually includes the variant (`org/product/variant`), but you can also point to the product level and let sew pick the default variant automatically.

When the resolved path has no `sew.yaml`, sew looks for a **`.default`** file in the same directory. This plain-text dotfile contains a single line — the name of the variant to use. sew then appends that variant to the path and resolves again.

```
registry/gravitee.io/apim/
├── .default          # contains "aio"
├── aio/
│   ├── .default      # contains "mongodb"
│   ├── base/
│   │   └── sew.yaml  # abstract: true
│   ├── mongodb/
│   │   └── sew.yaml
│   └── postgres/
│       └── sew.yaml
├── dbless/
│   └── sew.yaml
└── gateway-api/
    └── sew.yaml
```

With the tree above, setting `from: [gravitee.io/apim]` in your config is equivalent to `from: [gravitee.io/apim/aio/mongodb]` — sew reads `.default` at each level (`apim` → `aio` → `mongodb`) and resolves the full path.

To create a default for your own product, add a `.default` file next to the variant directories:

```bash
echo "aio" > registry/gravitee.io/apim/.default
```

If neither `sew.yaml` nor `.default` is found at the given path, sew returns an error.
