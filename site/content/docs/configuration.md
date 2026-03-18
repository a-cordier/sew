---
title: "Configuration"
weight: 3
type: docs
---

## Config format

```yaml
registry: file://./registry   # or https://...
from:
  - gravitee.io/apim/db-less

kind:
  apiVersion: kind.x-k8s.io/v1alpha4
  kind: Cluster
  name: sew
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 443
      hostPort: 443

# Optional: add Helm repos required by local components
repos:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami

# Optional: override context components or add new ones
components:
  - name: apim          # matches a context component → merged
    helm:
      version: "4.10.0"
      valueFiles:
        - ./my-overrides.yaml
  - name: redis          # no match in context → added as a new component
    namespace: gravitee
    helm:
      chart: bitnami/redis
      values:
        architecture: standalone
```

Value file paths under `components.*.helm.valueFiles` are resolved relative to the directory containing the config file.

## Kind port mappings

Port mappings defined in your `sew.yaml` are **merged** with those coming from the context (or defaults) using union semantics, keyed by `(containerPort, protocol)`:

- Ports that exist only in the context are **preserved**.
- Ports that exist only in your config are **added**.
- When both sides define the same `(containerPort, protocol)`, your entry **wins**.

This means you can extend a context's port requirements without losing its existing mappings. For example, if a context exposes ports 80 and 443, and your config adds port 9090, the cluster gets all three:

```yaml
# context defines ports 80 and 443
# your sew.yaml
kind:
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 9090   # added alongside context ports 80 and 443
      hostPort: 9090
    - containerPort: 443    # overrides the context's mapping for 443
      hostPort: 8443
```

The result contains three port mappings: 80 (from context), 443→8443 (your override), and 9090 (your addition).

The same union logic applies at every merge layer: context composition (`from`), user config over context, and user config over defaults.

## Local components

Beyond overriding fields on components defined by the context, you can declare entirely new components and Helm repos in your `sew.yaml`. This is useful when you need supporting services (databases, caches, message brokers, …) that are not part of the upstream context.

### Adding a component

List the component under `components`. If its `name` does not match any component from the context, sew appends it as a new component and installs it alongside the context ones:

```yaml
components:
  - name: redis
    namespace: gravitee
    helm:
      chart: bitnami/redis
      values:
        architecture: standalone
```

### Adding Helm repos

If the new component's chart comes from a repo that the context does not declare, add it under `repos`:

```yaml
repos:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
```

Local repos are merged with context repos. When both lists contain a repo with the same name, the local entry wins.

### Adding dependencies between components

You can make a context component wait for a local component (or vice-versa) by adding `requires` entries. Requirements are deduplicated by component name:

```yaml
components:
  - name: apim
    requires:
      - component: redis
        conditions:
          ready: true
        selector:
          matchLabels:
            app.kubernetes.io/instance: redis
```

### Component readiness

You can make sew wait for a component to become ready after installation by setting `conditions.ready: true`. This is useful when downstream components or features (e.g. DNS record collection) depend on the component being fully available.

```yaml
components:
  - name: apim
    conditions:
      ready: true
```

By default, sew waits for all pods matching the Helm release to be ready. You can narrow the check to specific pods with a label selector, and control the timeout with a Go duration string:

```yaml
components:
  - name: apim
    conditions:
      ready: true
    selector:
      matchLabels:
        app.kubernetes.io/component: gateway
    timeout: 10m
```

### Merge rules

When a local component matches a context component by name, the following merge rules apply:

| Field | Behaviour |
|-------|-----------|
| `conditions` | Local wins if `conditions.ready` is true |
| `selector` | Local wins if non-nil |
| `timeout` | Local wins if non-empty |
| `requires` | Local requirements are appended (deduplicated by component name) |
| `helm.chart` | Local wins if non-empty |
| `helm.version` | Local wins if non-empty |
| `helm.valueFiles` | Local files are appended (higher precedence in Helm) |
| `helm.values` | Local values are merged on top of context values |

When there is no name match, the component is added to the deployment as-is.
