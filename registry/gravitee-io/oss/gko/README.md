---
title: "GKO"
description: "Gravitee Kubernetes Operator standalone deployment"
tags: [networking]
---

# GKO

Deploys the Gravitee Kubernetes Operator (GKO) as a standalone Helm release.
This context is designed to be used on its own for GKO development, or composed
into other contexts (such as `dbless` or `gateway`) via `from:`.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from gravitee-io/oss/gko
```

### Cleanup

```bash
sew delete
```

## Quick Start

GKO reconciles Gravitee custom resources (`ApiDefinition`, `ApiV4Definition`,
`ManagementContext`, etc.) into gateway configuration. For details, see the
[GKO documentation](https://documentation.gravitee.io/gravitee-kubernetes-operator-gko/getting-started/quickstart-guide).

To iterate on GKO itself with a local build, compose from this context:

```yaml
from:
  - gravitee-io/oss/gko
builds:
  - name: gko
    image: graviteeio/kubernetes-operator:latest
    dir: .
```
