---
title: "APIM - Kafka MongoDB"
description: "Gravitee APIM with Kafka Gateway and MongoDB backend"
tags: [gravitee, api-management, kafka]
---

# APIM Kafka MongoDB

Deploys a full Gravitee API Management stack with Kafka Gateway enabled,
backed by MongoDB for persistence and Elasticsearch for analytics.

## Usage

```bash
sew create gravitee.io/apim/ee/kafka/mongodb
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

- `gravitee.io/apim/oss/aio/mongodb` — full APIM stack with MongoDB backend
- `gravitee.io/apim/ee/kafka/base` — Kafka Gateway configuration

See the [parent README](../README.md) for Kafka Gateway details, DNS setup,
TLS configuration, and license requirements.
