---
title: "APIM - PostgreSQL"
description: "Gravitee APIM with PostgreSQL backend and Elasticsearch"
tags: [api-management, gateway]
---

# APIM PostgreSQL

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by PostgreSQL for persistence and Elasticsearch for
analytics. Management uses JDBC. Rate limiting is also stored in PostgreSQL.

## Usage

```bash
sew create --from gravitee.io/oss/apim/postgres
```

## Endpoints

| Service        | URL                        |
|----------------|----------------------------|
| APIM Console   | http://localhost:30080      |
| APIM Portal    | http://localhost:30081      |
| APIM Gateway   | http://localhost:30082      |
| Management API | http://localhost:30083      |

## Context flags

Optional flags you can pass to `sew create` to customize this deployment:

| Flag           | Description                                    |
|----------------|------------------------------------------------|
| `--no-es`      | Disable Elasticsearch and analytics reporters  |
| `--no-ui`      | Disable both Console and Portal UIs            |
| `--no-portal`  | Disable the developer portal UI                |

```bash
sew create --from gravitee.io/oss/apim/postgres --no-es --no-portal
```

Use `sew info` to see the full list of flags and components for this context.

## Dependencies

This context composes from:

- `postgresql/standalone` — PostgreSQL 17 database
- `elastic/elasticsearch/standalone` — Elasticsearch for reporting
- `gravitee.io/oss/apim/base` — shared APIM Helm configuration
