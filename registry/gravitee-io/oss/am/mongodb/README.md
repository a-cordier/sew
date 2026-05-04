---
title: "AM - MongoDB"
description: "Gravitee Access Management with MongoDB backend"
tags: []
---

# AM MongoDB

Deploys a full Gravitee Access Management stack (Console UI, Gateway, and
Management API) backed by MongoDB for persistence.

## Usage

```bash
sew create --from gravitee-io/oss/am/mongodb
```

## Endpoints

| Service        | URL                        |
|----------------|----------------------------|
| AM Console     | http://localhost:30090      |
| AM Gateway     | http://localhost:30092      |
| Management API | http://localhost:30093      |

## Dependencies

This context composes from:

- `mongodb/standalone` — MongoDB 7 database
- `gravitee-io/oss/am/base` — shared AM Helm configuration
