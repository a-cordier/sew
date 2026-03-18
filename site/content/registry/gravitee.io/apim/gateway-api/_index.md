---
title: gravitee.io/apim/gateway-api
description: Gravitee Gateway using the Kubernetes Gateway API with GKO
tags:
    - gravitee
    - gateway
components:
    - gko
    - gateway-class-parameters
    - gateway-class
type: registry
---

# APIM Gateway API

Deploys the Gravitee Kubernetes Operator (GKO) configured as a Kubernetes
Gateway API controller. Sets up a `GatewayClass` and its parameters so that
`Gateway` and `HTTPRoute` resources are reconciled by GKO.

## Usage

```bash
sew create gravitee.io/apim/gateway-api
```

## Details

- **Kind cluster:** `gio-gateway-api`
- **Ports:** 80 (HTTP), 443 (HTTPS), 9092
- **Features:** load balancer, DNS
- **Components:** `gko` (Helm), `gateway-class-parameters` (CRD), `gateway-class` (GatewayClass)

This context enables the Kubernetes Gateway API flow: create `Gateway` and
`HTTPRoute` resources and let GKO provision Gravitee gateway instances
automatically.
