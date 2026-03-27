---
title: "APIM - MongoDB"
description: "Gravitee APIM with MongoDB backend and Elasticsearch"
tags: [api-management, gateway]
---

# APIM MongoDB

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by MongoDB for persistence and Elasticsearch for
analytics.

## Usage

```bash
sew create --from gravitee.io/oss/apim/mongodb
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
sew create --from gravitee.io/oss/apim/mongodb --no-es --no-portal
```

Use `sew info` to see the full list of flags and components for this context.

## Dependencies

This context composes from:

- `mongodb/standalone` — MongoDB 7 database
- `elastic/elasticsearch/standalone` — Elasticsearch for reporting
- `gravitee.io/oss/apim/base` — shared APIM Helm configuration
