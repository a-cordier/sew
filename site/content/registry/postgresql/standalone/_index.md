---
title: postgresql/standalone
description: Single-node PostgreSQL 17 deployment for Kubernetes
tags:
    - postgresql
    - database
    - sql
    - standalone
components:
    - postgresql
type: registry
---

# PostgreSQL Standalone

Deploys a single-replica PostgreSQL 17 instance as a Kubernetes Deployment with
a ClusterIP Service on port 5432.

## Usage

```bash
sew create postgresql/standalone
```

## Details

- **Image:** `postgres:17`
- **Port:** 5432
- **Database:** `gravitee`
- **Credentials:** `postgres` / `postgres`
- **Resources:** 250m–1 CPU, 256Mi–512Mi memory

This is a minimal, persistence-free PostgreSQL suitable for development and
testing. It is used as a dependency by higher-level contexts such as
`gravitee.io/apim/aio/postgres`.
