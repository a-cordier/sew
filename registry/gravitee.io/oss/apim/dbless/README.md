---
title: "APIM - DB-less"
description: "Gravitee APIM gateway in DB-less mode with the Gravitee Kubernetes Operator"
tags: [api-management, gateway, kubernetes]
---

# APIM DB-less

Deploys the Gravitee API Management gateway in DB-less mode alongside the
Gravitee Kubernetes Operator (GKO). No database involved — APIs are defined
entirely through Kubernetes custom resources.

## Usage

```bash
sew create --from gravitee.io/oss/apim/dbless
```

## Endpoints

| Service        | URL                        |
|----------------|----------------------------|
| APIM Gateway   | http://localhost:30082      |

## Context flags

This context inherits flags from `base`. Optional flags you can pass to
`sew create` to customize this deployment:

| Flag                 | Description                                          |
|----------------------|------------------------------------------------------|
| `--enable-hc-vault`  | Deploy HashiCorp Vault and configure it as a secret provider |
| `--enable-redis`     | Deploy Redis and use it for gateway rate limiting             |

```bash
sew create --from gravitee.io/oss/apim/dbless --enable-hc-vault
```

Use `sew info` to see the full list of flags and components for this context.

## Details

- **Kind cluster:** `gravitee-dbless`
- **Gateway port:** `http://localhost:30082`
- **Components:** `apim` (Helm), `gko` (Helm)
- **Database:** none (DB-less mode)
- **Elasticsearch:** disabled

This is the lightest Gravitee APIM setup, ideal for testing gateway
functionality and GKO-managed API definitions without any persistence layer.

## Dependencies

This context composes from:

- `gravitee.io/oss/apim/base` — shared APIM Helm configuration and flags
- `gravitee.io/oss/gko` — Gravitee Kubernetes Operator
