---
title: "Container Images"
weight: 4
type: docs
---

Pulling container images over the network every time you recreate a cluster gets old fast. sew gives you two strategies to speed things up: **mirror proxies** that cache layers locally, and **image preloading** that stages images before the cluster starts.

## Image mirrors

Mirror proxies run as local `registry:2` containers that cache image layers on your machine. When you delete and recreate a cluster, cached layers are reused -- no re-downloading.

### How it works

Each upstream registry gets its own pull-through cache container, bound to a local port. The Kind nodes' containerd is configured to check the mirror first. Mirror containers use a `restart: unless-stopped` policy, so they survive `sew delete` and keep their cache across cluster lifecycles.

### Enabling mirrors

Add `images.mirrors` to your `sew.yaml`:

```yaml
# Mirror docker.io only (always implicit)
images:
  mirrors: {}
```

```yaml
# Mirror docker.io + additional registries
images:
  mirrors:
    upstreams:
      - ghcr.io
      - docker.elastic.co
```

`docker.io` is always included -- you don't need to list it explicitly.

### Mirror options

| Field | Default | Description |
|-------|---------|-------------|
| `upstreams` | *(none)* | Additional registries to mirror (on top of `docker.io`) |
| `data` | `$SEW_HOME/mirrors` | Directory for cached layers and containerd host configs |

### Private registries

Combine mirrors with component value overrides to pull from a private registry through the local cache:

```yaml
images:
  mirrors:
    upstreams:
      - acme.example.com

components:
  - name: my-app
    helm:
      valueFiles:
        - values-private.yaml
```

## Image preloading

As an alternative (or complement) to mirrors, you can preload specific images into the cluster before components are deployed. This is especially useful on CI systems with Docker Layer Caching (DLC).

### How it works

1. Before the Kind cluster is created, sew pulls each image listed in `images.preload.refs` on the host Docker daemon.
2. A local `registry:2` container (`sew-preload`) is started. Pre-pulled images are re-tagged and pushed to this registry.
3. Kind nodes are configured to check the preload registry first for each upstream referenced by the listed images.

### Enabling preloading

```yaml
images:
  preload:
    refs:
      - docker.io/library/mongo:7
      - docker.elastic.co/elasticsearch/elasticsearch:8.17.0
      - bitnami/redis:7.4
```

Context authors can ship the image list directly in the context `sew.yaml`, so you don't have to maintain it yourself. You can add extra images in your own config -- refs from the context and your config are merged (deduplicated union).

### Combining mirrors and preloading

Both strategies work together. Preloading handles your known images (fast, DLC-friendly), while mirrors transparently cache anything else pulled at runtime:

```yaml
images:
  preload:
    refs:
      - docker.io/library/mongo:7
      - bitnami/redis:7.4
  mirrors:
    upstreams:
      - docker.elastic.co
```
