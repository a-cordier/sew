---
title: "Configuration"
weight: 1
type: docs
---

This page is generated from the [sew.yaml JSON Schema](https://github.com/a-cordier/sew/blob/main/schema/sew.schema.yaml). It documents every field you can use in your `sew.yaml` configuration file.

## Overview

A `sew.yaml` file is a YAML document with the following top-level fields:

| Field | Type | Description |
|-------|------|-------------|
| `abstract` | boolean | When true, marks this configuration as a shared base that cannot be deployed on its own. Abstract configs are meant to be referenced via 'from' by concrete contexts. |
| `components` | object[] | Ordered list of components to deploy. Components are applied sequentially; use 'requires' to express inter-component dependencies. |
| `features` | map | Optional networking features. Each sub-key uses pointer semantics: setting a feature explicitly overrides the inherited context default; omitting it preserves the parent value. |
| `from` | string[] | List of registry paths to compose from. Each referenced context is merged in order, allowing reuse of shared building blocks (databases, message brokers, etc.). |
| `helm` | map | Global Helm configuration shared across all components. |
| `images` | map | Container image management: preloading images into Kind nodes and configuring registry mirrors. |
| `kind` | map | Configuration for the Kind (Kubernetes-in-Docker) cluster. |
| `registry` | string | Registry path that identifies this configuration context (org/product/variant convention). |

---

## `abstract`

When true, marks this configuration as a shared base that cannot be deployed on its own. Abstract configs are meant to be referenced via 'from' by concrete contexts.

**Type:** `boolean` | **Default:** `false`

## `components`

Ordered list of components to deploy. Components are applied sequentially; use 'requires' to express inter-component dependencies.

**Type:** `array`

Each entry is an object with the following fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `conditions` | map | No | Readiness conditions for a component or dependency. |
| `helm` | map | No | Helm chart installation specification for a component of type "helm". The chart field may be omitted when the component inherits it from a parent context via 'from' composition. |
| `k8s` | map | No | Raw Kubernetes resource specification for a component of type "k8s". |
| `name` | string | Yes | Unique name identifying this component within the configuration. |
| `namespace` | string | No | Kubernetes namespace to deploy into. When omitted, the component is deployed into the default namespace. |
| `requires` | object[] | No | Dependencies on other components. Each requirement must be satisfied before this component is deployed. |
| `selector` | map | No | Label selector used to identify the pods to watch for readiness. |
| `timeout` | string | No | Maximum duration to wait for this component to become ready (e.g. "5m", "120s"). Used when conditions.ready is true. |
| `type` | string | No | Deployment type. "helm" installs a Helm chart; "k8s" applies raw Kubernetes manifests. Default: `helm`. Values: `helm`, `k8s`. |

### `components[*].conditions`

Readiness conditions for a component or dependency.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ready` | boolean | No | When true, the system waits for the component's pods (matched by selector) to be fully ready before proceeding. |

### `components[*].helm`

Helm chart installation specification for a component of type "helm". The chart field may be omitted when the component inherits it from a parent context via 'from' composition.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `chart` | string | No | Chart reference in "repo/chart" format (e.g. "bitnami/postgresql"). |
| `valueFiles` | string[] | No | Paths to YAML value files, resolved relative to the sew.yaml directory. Use for large overrides. |
| `values` | object | No | Inline Helm values merged on top of valueFiles. Use for small tweaks; prefer valueFiles for large overrides. |
| `version` | string | No | Explicit chart version constraint. When omitted, the latest version is used. |

### `components[*].k8s`

Raw Kubernetes resource specification for a component of type "k8s".

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `configMaps` | object[] | No | Local config maps to create as Kubernetes ConfigMap resources before applying other manifests. |
| `manifestFiles` | string[] | No | Paths to YAML manifest files, resolved relative to the sew.yaml directory. |
| `manifests` | object[] | No | Inline Kubernetes resource manifests. Each entry is a complete resource object (apiVersion, kind, metadata, etc.). |
| `secrets` | object[] | No | Local secrets to create as Kubernetes Secret resources before applying other manifests. |

#### `components[*].k8s.configMaps[*]`

A Kubernetes Secret or ConfigMap created from local files or environment variables.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `entries` | object[] | No | Multiple data entries for the resource, each sourced from a file or environment variable. |
| `fromFile` | string | No | Path to a single file whose contents become the resource data. Supports environment variable expansion (e.g. "$HOME/opt/license.key"). Shorthand for a single-entry resource. |
| `name` | string | Yes | Name of the Kubernetes Secret or ConfigMap to create. |
| `onMissing` | string | No | Behavior when a referenced file or env var is missing. "fail" aborts deployment; "ignore" skips the resource silently. Default: `fail`. Values: `fail`, `ignore`. |

#### `components[*].k8s.configMaps[*].entries[*]`

A single data entry in a LocalResource, sourced from a file or environment variable.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `fromEnv` | string | No | Name of an environment variable whose value becomes this entry's value. |
| `fromFile` | string | No | Path to a file whose contents become this entry's value. Supports environment variable expansion. |
| `key` | string | No | Key name in the resulting Secret or ConfigMap data map. |

#### `components[*].k8s.secrets[*]`

A Kubernetes Secret or ConfigMap created from local files or environment variables.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `entries` | object[] | No | Multiple data entries for the resource, each sourced from a file or environment variable. |
| `fromFile` | string | No | Path to a single file whose contents become the resource data. Supports environment variable expansion (e.g. "$HOME/opt/license.key"). Shorthand for a single-entry resource. |
| `name` | string | Yes | Name of the Kubernetes Secret or ConfigMap to create. |
| `onMissing` | string | No | Behavior when a referenced file or env var is missing. "fail" aborts deployment; "ignore" skips the resource silently. Default: `fail`. Values: `fail`, `ignore`. |

#### `components[*].k8s.secrets[*].entries[*]`

A single data entry in a LocalResource, sourced from a file or environment variable.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `fromEnv` | string | No | Name of an environment variable whose value becomes this entry's value. |
| `fromFile` | string | No | Path to a file whose contents become this entry's value. Supports environment variable expansion. |
| `key` | string | No | Key name in the resulting Secret or ConfigMap data map. |

### `components[*].requires[*]`

Declares a dependency on another component, optionally waiting for it to reach a ready state.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `component` | string | Yes | Name of the component this requirement depends on. |
| `conditions` | map | No | Readiness conditions for a component or dependency. |
| `selector` | map | No | Label selector used to identify the pods to watch for readiness. |
| `timeout` | string | No | Maximum duration to wait for the dependency to become ready (e.g. "5m", "15m"). |

#### `components[*].requires[*].conditions`

Readiness conditions for a component or dependency.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ready` | boolean | No | When true, the system waits for the component's pods (matched by selector) to be fully ready before proceeding. |

#### `components[*].requires[*].selector`

Label selector used to identify the pods to watch for readiness.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `matchLabels` | map | No | Map of label key-value pairs that pods must match. |

### `components[*].selector`

Label selector used to identify the pods to watch for readiness.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `matchLabels` | map | No | Map of label key-value pairs that pods must match. |

## `features`

Optional networking features. Each sub-key uses pointer semantics: setting a feature explicitly overrides the inherited context default; omitting it preserves the parent value.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `dns` | map | No | Local DNS server that resolves custom hostnames to in-cluster services, making them reachable from the host. |
| `gateway` | map | No | Kubernetes Gateway API support. Enabling gateway implicitly enables lb (load-balancer) as a dependency. |
| `lb` | map | No | Cloud-provider load-balancer emulation (e.g. MetalLB) for LoadBalancer-type services inside the Kind cluster. |

### `features.dns`

Local DNS server that resolves custom hostnames to in-cluster services, making them reachable from the host.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `domain` | string | No | DNS domain served by the local resolver. Default: `sew.local`. |
| `enabled` | boolean | No | Enable or disable the local DNS feature. |
| `port` | integer | No | UDP port the local DNS server listens on. Default: `15353`. |
| `records` | object[] | No | Static DNS records mapping hostnames to Kubernetes services. |

#### `features.dns.records[*]`

Maps a hostname (supports wildcards) to a Kubernetes service so the local DNS resolver can answer queries for it.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `hostname` | string | Yes | Hostname pattern to resolve (e.g. "*.kafka.sew.local"). |
| `namespace` | string | Yes | Namespace of the target Kubernetes Service. |
| `service` | string | Yes | Name of the Kubernetes Service to resolve the hostname to. |

### `features.gateway`

Kubernetes Gateway API support. Enabling gateway implicitly enables lb (load-balancer) as a dependency.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `channel` | string | No | Gateway API release channel to install. Values: `standard`, `experimental`. |
| `enabled` | boolean | No | Enable or disable Gateway API support. |

### `features.lb`

Cloud-provider load-balancer emulation (e.g. MetalLB) for LoadBalancer-type services inside the Kind cluster.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | boolean | No | Enable or disable the load-balancer feature. |

## `from`

List of registry paths to compose from. Each referenced context is merged in order, allowing reuse of shared building blocks (databases, message brokers, etc.).

**Type:** `array`

## `helm`

Global Helm configuration shared across all components.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repos` | object[] | No | Helm chart repositories to register before installing components. |

### `helm.repos[*]`

A Helm chart repository.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Local alias for the repository (used as prefix in chart references, e.g. "bitnami" in "bitnami/postgresql"). |
| `url` | string | Yes | URL of the Helm chart repository. |

## `images`

Container image management: preloading images into Kind nodes and configuring registry mirrors.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `mirrors` | map | No | Registry mirror configuration. Mirrors are injected into the Kind nodes' containerd config so pulls go through the mirror first. |
| `preload` | map | No | Images to pull and preload into Kind nodes before deploying components, avoiding in-cluster pulls and speeding up startup. |

### `images.mirrors`

Registry mirror configuration. Mirrors are injected into the Kind nodes' containerd config so pulls go through the mirror first.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `data` | string | No | Raw containerd mirror configuration data to inject. |
| `upstreams` | string[] | No | List of upstream registry mirror URLs. |

### `images.preload`

Images to pull and preload into Kind nodes before deploying components, avoiding in-cluster pulls and speeding up startup.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `refs` | string[] | No | List of fully-qualified image references (e.g. "docker.io/library/nginx:1.27-alpine") to preload. |

## `kind`

Configuration for the Kind (Kubernetes-in-Docker) cluster.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | No | Kind API version for the cluster manifest. Default: `kind.x-k8s.io/v1alpha4`. |
| `containerdConfigPatches` | string[] | No | Raw containerd configuration patches applied to all nodes in the cluster. |
| `kind` | string | No | Kind resource type (always "Cluster"). Default: `Cluster`. |
| `name` | string | No | Name of the Kind cluster. Default: `sew`. |
| `nodes` | object[] | No | List of cluster nodes. Defaults to a single control-plane node when omitted. |

### `kind.nodes[*]`

Configuration for a single Kind cluster node.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `extraMounts` | object[] | No | Additional bind-mounts from the host filesystem into the node container. |
| `extraPortMappings` | object[] | No | Additional port mappings from the host to the node container, enabling access to NodePort services from the host machine. |
| `kubeadmConfigPatches` | string[] | No | Raw kubeadm configuration patches applied to this node during cluster creation. |
| `labels` | map | No | Labels to apply to the node. |
| `role` | string | No | Node role in the cluster (e.g. "control-plane" or "worker"). |

#### `kind.nodes[*].extraMounts[*]`

Bind-mount from the host into a Kind node container.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `containerPath` | string | Yes | Path inside the Kind node container. |
| `hostPath` | string | Yes | Absolute path on the host filesystem. |

#### `kind.nodes[*].extraPortMappings[*]`

Maps a port from the Kind node container to the host.

**Type:** `object`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `containerPort` | integer | Yes | Port inside the Kind node container. |
| `hostPort` | integer | Yes | Port on the host machine. |
| `protocol` | string | No | Transport protocol (e.g. "TCP", "UDP"). Defaults to TCP when omitted. |

## `registry`

Registry path that identifies this configuration context (org/product/variant convention).

**Type:** `string`

