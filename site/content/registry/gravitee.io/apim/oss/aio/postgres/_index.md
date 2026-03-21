---
title: APIM - AIO PostgreSQL
path: gravitee.io/apim/oss/aio/postgres
context: true
description: Gravitee APIM all-in-one with PostgreSQL backend and Elasticsearch
tags:
    - gravitee
    - api-management
from:
    - postgresql/standalone
    - elastic/elasticsearch
    - gravitee.io/apim/oss/aio/base
components:
    - postgresql
    - elasticsearch
    - apim
notes_create: |-
    Your cluster "{{ .Kind.Name }}" is ready.

    Everything has been deployed in the `gravitee` namespace.

    APIM Console     http://localhost:30080
    APIM Portal      http://localhost:30081
    APIM Gateway     http://localhost:30082
    Management API   http://localhost:30083
type: registry
---

# APIM AIO PostgreSQL

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by PostgreSQL for persistence and Elasticsearch for
analytics. Management uses JDBC. Rate limiting is also stored in PostgreSQL.

## Usage

```bash
sew create gravitee.io/apim/oss/aio/postgres
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

- `postgresql/standalone` — PostgreSQL 17 database
- `elastic/elasticsearch` — Elasticsearch for reporting
- `gravitee.io/apim/oss/aio/base` — shared APIM Helm configuration
