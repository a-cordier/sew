---
title: "Context Format"
weight: 3
type: docs
---

This page is for context **authors** -- the people who create and maintain registry contexts for others to use. If you're just using contexts, see [Composing Contexts]({{< ref "/docs/guides/composing-contexts" >}}).

## Anatomy of a context

A context lives at `{registry}/{context_path}/` and must contain a `sew.yaml`. At minimum, it declares components to deploy:

```yaml
helm:
  repos:
    - name: bitnami
      url: https://charts.bitnami.com/bitnami

components:
  - name: mongodb
    type: helm
    namespace: default
    helm:
      chart: bitnami/mongodb
      version: "16.4.0"
      valueFiles:
        - values.yaml
```

File paths in `valueFiles` and `manifestFiles` are resolved relative to the context directory.

## Component types

### Helm components

The default type. Installs a Helm chart:

```yaml
components:
  - name: my-app
    type: helm
    namespace: my-ns
    helm:
      chart: myrepo/my-app
      version: "2.0.0"
      valueFiles:
        - values.yaml
      values:
        replicas: 1
```

### Kubernetes manifest components

Deploy plain Kubernetes resources by setting `type: k8s`. You can use inline manifests, file references, or both:

```yaml
components:
  - name: routes
    type: k8s
    namespace: my-ns
    k8s:
      manifestFiles:
        - gateway.yaml
        - httproutes.yaml
      manifests:
        - apiVersion: v1
          kind: Service
          metadata:
            name: my-service
          spec:
            type: ClusterIP
            ports:
              - port: 8080
```

File-based resources are applied first, then inline manifests. Both component types participate in the same dependency graph.

## Composition with `from`

Contexts can compose other contexts. List the parent paths in `from`:

```yaml
from:
  - mongodb/standalone
  - elastic/elasticsearch/standalone

components:
  - name: my-app
    requires:
      - component: mongodb
      - component: elasticsearch
    helm:
      chart: myrepo/my-app
```

Parents are resolved left-to-right, then local overrides are applied on top using the standard [merge rules]({{< ref "/docs/guides/composing-contexts#merge-rules" >}}).

### Cross-registry composition

By default, `from` entries are resolved against the same registry. To compose from a different registry:

```yaml
registry: https://other-registry.example.com
from:
  - org/product/base
```

Relative `file://` paths are resolved relative to the child context's directory.

## Abstract contexts

Mark a context as `abstract: true` when it's a shared base that shouldn't be deployed on its own:

```yaml
abstract: true

helm:
  repos:
    - name: myrepo
      url: https://charts.example.com

components:
  - name: app
    helm:
      chart: myrepo/app
      version: "2.0.0"
```

Attempting to deploy an abstract context directly with `sew create` produces an error. Concrete variants must compose from it via `from`.

## Default variant resolution

Add a `.default` file next to variant directories to set the default:

```bash
echo "standalone" > registry/mongodb/.default
```

When a user specifies `from: [mongodb]`, sew reads `.default` and resolves it to `mongodb/standalone`. Defaults chain across multiple levels.

## Merge semantics

When contexts are composed, each top-level field is merged as follows:

- **`kind`** -- Scalar fields: child wins if set. `nodes`: child replaces the list, but `extraPortMappings` are union-merged by `(containerPort, protocol)`. `containerdConfigPatches`: child replaces entirely.
- **`components`** -- Matched by name. Helm chart/version: child wins. Value files: appended. Values: deep-merged. Manifest files: appended. Manifests: union by resource identity. Secrets/configMaps: appended. Requirements: appended and deduplicated. Unmatched components are appended.
- **`helm.repos`** -- Deduplicated by name; child wins on conflict.
- **`features`** -- Each feature block is replaced as a whole if the child defines it; otherwise inherited.
- **`images`** -- `preload.refs`: deduplicated union. `mirrors`: child wins if set.

## Overriding service networking

When your context composes from a child that exposes services via `NodePort`, you might need to switch them to `ClusterIP` because your context handles networking differently.

For **Helm components**, override the values and explicitly clear `nodePort`:

```yaml
components:
  - name: child-service
    helm:
      values:
        service:
          type: ClusterIP
          nodePort: null
```

For **k8s manifest components**, provide a full replacement Service manifest. Manifests are merged by resource identity, so your Service replaces the child's entirely.

## Registry organization tips

- Use the `org/product/variant` convention for discoverability
- Extract shared config into `abstract: true` base contexts
- Set `.default` files so users can reference products without spelling out the full variant path
- Include a `README.md` with front matter (`title`, `description`, `tags`) -- the site generator uses it for the registry browser
- Add `notes.create` with post-deployment instructions that `sew create` prints after a successful deploy
