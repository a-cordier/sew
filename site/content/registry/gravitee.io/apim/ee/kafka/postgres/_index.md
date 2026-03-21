---
title: APIM - Kafka PostgreSQL
layout: detail
path: gravitee.io/apim/ee/kafka/postgres
context: true
description: Gravitee APIM with Kafka Gateway and PostgreSQL backend
tags:
    - api-management
    - gateway
    - messaging
    - kafka
from:
    - gravitee.io/apim/oss/aio/postgres
    - gravitee.io/apim/ee/kafka/base
type: registry
---

Deploys a full Gravitee API Management stack with Kafka Gateway enabled,
backed by PostgreSQL for persistence and Elasticsearch for analytics.

## Usage

```bash
sew create gravitee.io/apim/ee/kafka/postgres
```

## Endpoints

| Service        | URL                            |
|----------------|--------------------------------|
| APIM Console   | http://localhost:30080          |
| APIM Portal    | http://localhost:30081          |
| APIM Gateway   | http://localhost:30082          |
| Management API | http://localhost:30083          |
| Kafka Gateway  | `*.kafka.sew.local:9092` (TLS) |

## Dependencies

This context composes from:

- `gravitee.io/apim/oss/aio/postgres` — full APIM stack with PostgreSQL backend
- `gravitee.io/apim/ee/kafka/base` — Kafka Gateway configuration

See the [parent README](../README.md) for Kafka Gateway details, DNS setup,
TLS configuration, and license requirements.
