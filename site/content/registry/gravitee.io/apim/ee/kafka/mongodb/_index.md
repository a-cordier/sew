---
title: APIM - Kafka MongoDB
layout: detail
path: gravitee.io/apim/ee/kafka/mongodb
context: true
description: Gravitee APIM with Kafka Gateway and MongoDB backend
tags:
    - api-management
    - gateway
    - messaging
from:
    - gravitee.io/apim/oss/aio/mongodb
    - gravitee.io/apim/ee/kafka/base
type: registry
---

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
and TLS configuration.

## License

This is an Enterprise Edition (EE) context. Place your Gravitee license
key at `$HOME/opt/gravitee/license.key` and sew will automatically mount
it into the cluster. If the file is missing, the license component is
silently skipped (`onMissing: ignore`).

To use a different path, override it in your `sew.yaml`:

```yaml
components:
  - name: license
    k8s:
      secrets:
        - name: gravitee-license
          fromFile: '/custom/path/to/license.key'
```
