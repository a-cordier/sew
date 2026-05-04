---
title: "Edge Stack"
description: "Ambassador Edge Stack API gateway"
tags: [networking]
---

# Edge Stack

Deploys Ambassador Edge Stack, a Kubernetes-native API gateway built on
Envoy Proxy, into a local Kind cluster. Edge Stack provides advanced
routing, rate limiting, authentication, and TLS termination.

## Usage

```bash
sew create --from gravitee-io/ee/edge-stack
```

## Endpoints

| Service | URL                    |
|---------|------------------------|
| HTTP    | http://localhost:30080  |
| HTTPS   | https://localhost:30443 |

## Details

- **Kind cluster:** `gravitee-edge-stack`
- **Namespace:** `ambassador`
- **Helm chart:** `datawire/edge-stack`
- **Image:** `docker.io/datawire/aes:3.12.7`
- **Ports:** 30080 (HTTP), 30443 (HTTPS)

The edge-stack component exposes HTTP and HTTPS via NodePort services.
The Helm release uses the `emissary-ingress` sub-chart for the data
plane (Envoy) and adds Edge Stack's authentication and rate limiting
features on top.

## Dependencies

This context composes from:

- `redis/standalone` — Redis instance for rate limiting and authentication

## License

This is an Enterprise Edition (EE) context. Place your Edge Stack license
at `$HOME/opt/gravitee/edge-stack/license.jwt` and sew will automatically
mount it into the cluster as the `ambassador-edge-stack` Secret. If the
file is missing, the license component is silently skipped
(`onMissing: ignore`).

To use a different path, override it in your `sew.yaml`:

```yaml
components:
  - name: license
    k8s:
      secrets:
        - name: ambassador-edge-stack
          entries:
            - key: license-key
              fromFile: '/custom/path/to/license.jwt'
```
