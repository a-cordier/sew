---
title: APIM - Gateway API
layout: detail
path: gravitee.io/apim/oss/k8s/gateway
context: true
description: Gravitee Gateway using the Kubernetes Gateway API with GKO
tags:
    - gateway
    - operator
components:
    - gko
    - gateway-class-parameters
    - gateway-class
notes_create: |-
    Your cluster "gravitee-gateway" is ready.

    GKO is deployed in the `gravitee` namespace with Gateway API support enabled.

    Create Gateway and HTTPRoute resources to start routing traffic.
icon: gravitee.io/icon.svg
type: registry
---

Deploys the Gravitee Kubernetes Operator (GKO) configured as a Kubernetes
Gateway API controller. Sets up a `GatewayClass` and its parameters so that
`Gateway` and `HTTPRoute` resources are reconciled by GKO.

## Usage

```bash
sew create --from gravitee.io/apim/oss/k8s/gateway
```

## Details

- **Kind cluster:** `gio-gateway-api`
- **Ports:** 80 (HTTP), 443 (HTTPS), 9092
- **Features:** load balancer, DNS
- **Components:** `gko` (Helm), `gateway-class-parameters` (CRD), `gateway-class` (GatewayClass)

This context enables the Kubernetes Gateway API flow: create `Gateway` and
`HTTPRoute` resources and let GKO provision Gravitee gateway instances
automatically.
