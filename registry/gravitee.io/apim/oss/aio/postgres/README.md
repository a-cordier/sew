---
title: "APIM - AIO PostgreSQL"
description: "Gravitee APIM all-in-one with PostgreSQL backend and Elasticsearch"
tags: [api-management, gateway]
---

# APIM AIO PostgreSQL

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by PostgreSQL for persistence and Elasticsearch for
analytics. Management uses JDBC. Rate limiting is also stored in PostgreSQL.

## Usage

```bash
sew create --from gravitee.io/apim/oss/aio/postgres
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
