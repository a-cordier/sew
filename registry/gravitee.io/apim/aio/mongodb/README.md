---
description: "Gravitee APIM all-in-one with MongoDB backend and Elasticsearch"
tags: [gravitee, apim, mongodb, all-in-one, api-management]
---

# APIM AIO MongoDB

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by MongoDB for persistence and Elasticsearch for
analytics.

## Usage

```bash
sew create gravitee.io/apim/aio/mongodb
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
- `gravitee.io/apim/aio/base` — shared APIM Helm configuration
