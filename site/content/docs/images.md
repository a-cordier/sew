---
title: "Images"
weight: 5
type: docs
---

## Image mirrors

sew can run local pull-through mirror proxies for container registries. When enabled, images pulled inside the Kind cluster are cached on the host — subsequent cluster recreations reuse the cached layers instead of downloading from the internet.

### How it works

Each upstream registry gets its own `registry:2` container running as a pull-through cache, bound to a local port (5000, 5001, …). The Kind node's containerd is configured with `hosts.toml` files that redirect pulls through these local mirrors. The mirror containers use a `restart: unless-stopped` policy, so they survive `sew delete` and keep their cache across cluster lifecycles.

### Enabling mirrors

Add an `images.mirrors` section to your `sew.yaml`:

```yaml
# Mirror docker.io only (always implicit)
images:
  mirrors: {}
```

```yaml
# Mirror docker.io + a private registry
images:
  mirrors:
    upstreams:
    - acme.example.com
```

`docker.io` is always included — you don't need to list it explicitly.

### Options

| Field | Default | Description |
|-------|---------|-------------|
| `upstreams` | *(none)* | Additional registries to mirror (on top of `docker.io`) |
| `data` | `$SEW_HOME/mirrors` | Directory for cached layers and containerd host configs |

### Using images from a private registry

Combine mirrors with local component overrides to pull images from a private registry through the local cache. Create a values file with the private image coordinates:

```yaml
# values-private.yaml
gateway:
  image:
    repository: acme.example.com/my-gateway
    tag: 1.0.0
```

Then reference it in your `sew.yaml`:

```yaml
registry: https://my-registry.example.com
from:
  - org/product/variant

images:
  mirrors:
    upstreams:
    - acme.example.com

components:
  - name: apim
    helm:
      valueFiles:
      - values-private.yaml
```

The mirror proxy caches layers from `acme.example.com` locally, and the component override swaps the image without modifying the upstream context.

## Image preloading

As an alternative (or complement) to mirror proxies, sew can preload images via a local registry. This is especially useful on CI systems with Docker Layer Caching (DLC), where pulled image layers persist across runs without any filesystem-based cache.

### How it works

1. **`docker pull`** — Before the Kind cluster is created, sew pulls each image listed in `images.preload.refs` on the host Docker daemon. DLC caches these layers, so subsequent CI runs skip the network pull entirely.
2. **Local preload registry** — A plain `registry:2` container (`sew-preload`) is started. Pre-pulled images are re-tagged and pushed to this local registry.
3. **Containerd `hosts.toml`** — For each upstream registry referenced by the preloaded images, sew generates a containerd host configuration that directs Kind nodes to check the preload registry first. If the image is found there, no network pull occurs. If not, containerd falls back to upstream (or to a mirror proxy if configured).

### Enabling preloading

Add an `images.preload` block with a `refs` list to your `sew.yaml`:

```yaml
images:
  preload:
    refs:
      - graviteeio/apim-gateway:latest
      - graviteeio/apim-management-api:latest
      - mongo:7
```

Context authors can ship the image list directly in the context `sew.yaml`, so users don't need to maintain it:

```yaml
# registry/org/product/variant/sew.yaml
images:
  preload:
    refs:
      - graviteeio/apim-gateway:latest
      - graviteeio/apim-management-api:latest
      - mongo:7

components:
  - name: app
    helm:
      chart: app/chart
```

Users can add extra images in their own `sew.yaml` — refs from the context and the user config are merged (deduplicated union):

```yaml
# sew.yaml (user)
registry: https://my-registry.example.com
from:
  - org/product/variant

images:
  preload:
    refs:
      - my-registry.io/my-sidecar:v1.2
```

### Combining preload and mirrors

Both strategies can be enabled simultaneously. Preloading handles the known images (fast, DLC-friendly), while mirrors transparently cache any additional images pulled at runtime:

```yaml
images:
  preload:
    refs:
      - graviteeio/apim-gateway:latest
      - mongo:7
  mirrors:
    upstreams:
      - docker.elastic.co
```
