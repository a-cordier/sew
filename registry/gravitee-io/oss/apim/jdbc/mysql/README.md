---
title: "APIM - JDBC MySQL"
description: "Gravitee APIM with MySQL JDBC backend and Elasticsearch"
tags: [networking]
---

# APIM JDBC MySQL

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by MySQL via JDBC for persistence and Elasticsearch
for analytics. Rate limiting is also stored in MySQL.

## Usage

```bash
sew create --from gravitee-io/oss/apim/jdbc/mysql
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

| Flag                 | Description                                          |
|----------------------|------------------------------------------------------|
| `--disable-es`       | Disable Elasticsearch and analytics reporters        |
| `--disable-ui`       | Disable both Console and Portal UIs                  |
| `--disable-portal`   | Disable the developer portal UI                      |
| `--enable-hc-vault`  | Deploy HashiCorp Vault and configure it as a secret provider |
| `--enable-redis`     | Deploy Redis and use it for gateway rate limiting             |

```bash
sew create --from gravitee-io/oss/apim/jdbc/mysql --disable-es --disable-portal
```

Use `sew info` to see the full list of flags and components for this context.

## Dependencies

This context composes from:

- `mysql/standalone` — MySQL 9 database
- `elastic/elasticsearch/standalone` — Elasticsearch for reporting
- `gravitee-io/oss/apim/jdbc/base` — shared APIM JDBC configuration
