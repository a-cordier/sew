---
title: APIM - AIO MongoDB
layout: detail
path: gravitee.io/apim/oss/aio/mongodb
context: true
description: Gravitee APIM all-in-one with MongoDB backend and Elasticsearch
tags:
    - api-management
    - gateway
from:
    - mongodb/standalone
    - elastic/elasticsearch
    - gravitee.io/apim/oss/aio/base
components:
    - mongodb
    - elasticsearch
    - tls-server
    - apim
notes_create: |-
    Your cluster "gravitee.io/apim/oss/aio/mongodb" is ready.

    Everything has been deployed in the `gravitee` namespace.

    APIM Console     http://localhost:30080
    APIM Portal      http://localhost:30081
    APIM Gateway     http://localhost:30082
    Management API   http://localhost:30083
icon: gravitee.io/icon.svg
type: registry
---

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by MongoDB for persistence and Elasticsearch for
analytics.

## Usage

```bash
sew create --from gravitee.io/apim/oss/aio/mongodb
```

## Endpoints

| Service        | URL                        |
|----------------|----------------------------|
| APIM Console   | http://localhost:30080      |
| APIM Portal    | http://localhost:30081      |
| APIM Gateway   | http://localhost:30082      |
| Management API | http://localhost:30083      |

## Dependencies

This context composes from:

- `mongodb/standalone` — MongoDB 7 database
- `elastic/elasticsearch` — Elasticsearch for reporting
- `gravitee.io/apim/oss/aio/base` — shared APIM Helm configuration
