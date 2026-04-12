---
title: "GKO"
description: "Gravitee Kubernetes Operator standalone deployment"
tags: [operator, kubernetes]
---

# GKO

Deploys the Gravitee Kubernetes Operator (GKO) as a standalone Helm release.
This context is designed to be used on its own for GKO development, or composed
into other contexts (such as `dbless` or `gateway`) via `from:`.

## Usage

```bash
sew create --from gravitee.io/oss/gko
```

## Details

- **Kind cluster:** `gravitee-gko`
- **Components:** `gko` (Helm)
- **Namespace:** `gravitee`

GKO reconciles Gravitee custom resources (`ApiDefinition`, `ApiV4Definition`,
`ManagementContext`, etc.) into gateway configuration. Use this context when
you need a minimal cluster with just the operator — for example, to iterate on
GKO itself with a local build:

```yaml
from:
  - gravitee.io/oss/gko
builds:
  - name: gko
    image: graviteeio/kubernetes-operator:latest
    dir: .
```
