---
title: APIM - DB-less
path: gravitee.io/apim/oss/k8s/dbless
context: true
description: Gravitee APIM gateway in DB-less mode with the Gravitee Kubernetes Operator
tags:
    - gravitee
    - api-management
    - operator
components:
    - apim
    - gko
notes_create: |-
    Your cluster "{{ .Kind.Name }}" is ready.

    Everything has been deployed in the `gravitee` namespace.

    APIM Gateway is listening on http://localhost:30082
type: registry
---

# APIM DB-less

Deploys the Gravitee API Management gateway in DB-less mode alongside the
Gravitee Kubernetes Operator (GKO). No database or analytics backend is
required — APIs are defined entirely through Kubernetes custom resources.

## Usage

```bash
sew create gravitee.io/apim/oss/k8s/dbless
```

## Details

- **Kind cluster:** `gio-dbless`
- **Gateway port:** `http://localhost:30082`
- **Components:** `apim` (Helm), `gko` (Helm)
- **Database:** none (DB-less mode)
- **Elasticsearch:** disabled

This is the lightest Gravitee APIM setup, ideal for testing gateway
functionality and GKO-managed API definitions without any persistence layer.
