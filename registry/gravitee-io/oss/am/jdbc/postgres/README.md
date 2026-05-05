---
title: "AM - JDBC PostgreSQL"
description: "Gravitee Access Management with PostgreSQL JDBC backend"
tags: [networking, security]
---

# AM JDBC PostgreSQL

Deploys a full Gravitee Access Management stack (Console UI, Gateway, and
Management API) backed by PostgreSQL via JDBC for persistence.

## Usage

```bash
sew create --from gravitee-io/oss/am/jdbc/postgres
```

## Quick Start

Sign in to the Console at [http://localhost:30090](http://localhost:30090)
with the default admin account (`admin` / `adminadmin`).

To configure your first identity provider, follow the Gravitee
[AM quick start guide](https://documentation.gravitee.io/am/getting-started/quickstart-guide).

## Endpoints

| Service        | URL                   |
|----------------|-----------------------|
| AM Console     | http://localhost:30090 |
| AM Gateway     | http://localhost:30092 |
| Management API | http://localhost:30093 |
