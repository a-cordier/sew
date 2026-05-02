---
title: "AM - JDBC MySQL"
description: "Gravitee Access Management with MySQL JDBC backend"
tags: [networking, security]
---

# AM JDBC MySQL

Deploys a full Gravitee Access Management stack (Console UI, Gateway, and
Management API) backed by MySQL via JDBC for persistence.

## Usage

```bash
sew create --from gravitee.io/oss/am/jdbc/mysql
```

## Endpoints

| Service        | URL                        |
|----------------|----------------------------|
| AM Console     | http://localhost:30090      |
| AM Gateway     | http://localhost:30092      |
| Management API | http://localhost:30093      |

## Context flags

Optional flags you can pass to `sew create` to customize this deployment:

| Flag             | Description                    |
|------------------|--------------------------------|
| `--disable-ui`   | Disable the AM Console UI      |

```bash
sew create --from gravitee.io/oss/am/jdbc/mysql --disable-ui
```

Use `sew info` to see the full list of flags and components for this context.

## Dependencies

This context composes from:

- `mysql/standalone` — MySQL 9 database
- `gravitee.io/oss/am/jdbc/base` — shared AM JDBC configuration
