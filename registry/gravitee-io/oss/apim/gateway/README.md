---
title: "Gateway API"
description: "Gravitee implementation of the Kubernetes Gateway API"
tags: [networking]
---

# APIM Gateway API

Deploys the Gravitee Kubernetes Operator (GKO) configured as a Kubernetes
Gateway API controller. Sets up a `GatewayClass` and its parameters so that
`Gateway` and `HTTPRoute` resources are reconciled by GKO.

## Usage

```bash
sew create --from gravitee-io/oss/apim/gateway
```

## Prerequisites

This context uses DNS to resolve in-cluster services by hostname. After
creating the cluster, run the one-time OS setup so these hostnames resolve
on your machine (may require `sudo`):

```bash
sew setup dns
```

See the [Networking guide](https://a-cordier.github.io/sew/docs/guides/networking/#local-dns) for details.

## Quick Start

Create a `Gateway` and an `HTTPRoute` resource and let GKO provision
Gravitee gateway instances automatically:

```bash
kubectl apply -f my-gateway.yaml -n gravitee
kubectl apply -f my-route.yaml -n gravitee
```

For details on the Gateway API model, see the
[Kubernetes Gateway API documentation](https://gateway-api.sigs.k8s.io/)
and the [GKO documentation](https://documentation.gravitee.io/gko).
