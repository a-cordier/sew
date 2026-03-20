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
helm:
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

List the component under `components`. If its `name` does not match any component from the context, sew appends it as a new component and installs it alongside the context ones.

A component can be either a Helm chart (the default) or plain Kubernetes manifests (`type: k8s`):

```yaml
components:
  # Helm component (default type)
  - name: redis
    namespace: gravitee
    helm:
      chart: bitnami/redis
      values:
        architecture: standalone

  # k8s manifest component
  - name: mongodb
    type: k8s
    namespace: gravitee
    k8s:
      manifestFiles:
        - manifests/mongodb.yaml
      manifests:
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

A `k8s` component supports two ways of providing manifests:

- **`manifestFiles`** — a list of paths to YAML files (resolved relative to the config file directory, may contain multiple documents separated by `---`).
- **`manifests`** — a list of inline Kubernetes resource definitions.

Both can be used together; manifest files are applied first, then inline manifests.

### Local Secrets and ConfigMaps

A `k8s` component can also create Kubernetes Secrets and ConfigMaps from local files or environment variables using the `k8s.secrets` and `k8s.configMaps` fields. This is useful for injecting license keys, credentials, or configuration files that live outside version control.

#### Shorthand syntax

When a Secret or ConfigMap has a single file source, use the `fromFile` shorthand. The data key defaults to the file's basename:

```yaml
components:
  - name: licensing
    type: k8s
    namespace: my-app
    k8s:
      secrets:
        - name: gravitee-license
          fromFile: ./gravitee-license.txt
          onMissing: ignore
```

This creates a Secret named `gravitee-license` with a single key `gravitee-license.txt` whose value is the file's contents.

#### Full entries syntax

For resources with multiple data entries, or entries sourced from environment variables, use the `entries` list:

```yaml
components:
  - name: app-config
    type: k8s
    namespace: my-app
    k8s:
      secrets:
        - name: my-credentials
          onMissing: fail
          entries:
            - key: token
              fromFile: ./token.txt
            - key: API_KEY
              fromEnv: MY_API_KEY
      configMaps:
        - name: custom-logging
          onMissing: ignore
          entries:
            - key: logback.xml
              fromFile: ./logback.xml
            - key: LOG_LEVEL
              fromEnv: LOG_LEVEL
```

Each entry specifies either `fromFile` (read from a local file) or `fromEnv` (read from an environment variable), but not both. The `key` field is optional — it defaults to the basename of the file path for `fromFile` entries, or the variable name for `fromEnv` entries.

For Secrets, values are stored in `stringData` (the API server base64-encodes them on write). For ConfigMaps, values go into `data` as plain strings.

#### Missing sources

The `onMissing` field controls what happens when a file or environment variable is not found:

| Value | Behaviour |
|-------|-----------|
| `fail` (default) | The setup errors out immediately |
| `ignore` | A warning is logged and the entire resource is skipped |

`onMissing` applies at the resource level — if any entry in a resource is missing and `onMissing` is `ignore`, the whole resource is skipped (not just the missing entry).

#### Path resolution

`fromFile` paths are resolved relative to the directory containing the config file, following the same convention as `helm.valueFiles` and `k8s.manifestFiles`.

### Adding Helm repos

If the new component's chart comes from a repo that the context does not declare, add it under `helm.repos`:

```yaml
helm:
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
| `helm.values` | Deep-merged on top of context values (see [Values deep merge](#values-deep-merge) below) |
| `k8s.manifestFiles` | Local files are appended |
| `k8s.manifests` | Union by resource identity (`apiVersion`, `kind`, `name`, `namespace`); later wins on conflict |
| `k8s.secrets` | Local secrets are appended |
| `k8s.configMaps` | Local configMaps are appended |

When there is no name match, the component is added to the deployment as-is.

### Values deep merge

When `helm.values` from two layers overlap on the same key, sew picks a merge strategy based on the value type:

| Value type | Strategy | Example keys |
|-----------|----------|-------------|
| Maps | Recursive deep merge — each nested key is merged individually, child wins per leaf | `gateway.image`, `jdbc` |
| Named lists (every element is an object with a `name` key) | Merge by `name` — same-name entries are overridden in place, new entries are appended | `env`, `ports`, `volumeMounts` |
| Everything else (scalars, plain lists, unnamed object lists) | Replace — child value wins | `replicas`, `es.endpoints`, `servers` |

Named-list merging follows the Kubernetes convention: fields like `env`, `ports`, `volumeMounts`, and `containers` all use `name` as an identity key. This means composed contexts can each contribute entries to the same list without clobbering each other.

**Example**: a postgres context sets JDBC rate-limiting env vars, and a Kafka context adds Kafka env vars. After composition, both sets of env vars are present:

```yaml
# Context A (postgres) defines:
gateway:
  env:
    - name: gravitee_ratelimit_type
      value: jdbc

# Context B (kafka) defines:
gateway:
  env:
    - name: KAFKA_PORT
      value: "9092"

# Merged result — both entries are combined:
gateway:
  env:
    - name: gravitee_ratelimit_type
      value: jdbc
    - name: KAFKA_PORT
      value: "9092"
```

If both sides define an entry with the same `name`, the later (child) value wins. An empty list (`env: []`) replaces the parent's list entirely — use this to clear inherited entries.
