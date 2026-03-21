---
description: "Single-node PostgreSQL 17 deployment for Kubernetes"
tags: [database]
---

# PostgreSQL Standalone

Deploys a single-replica PostgreSQL 17 instance as a Kubernetes Deployment with
a NodePort Service on port 5432.

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

### Host access

Kind maps `hostPort 5432` → `containerPort 30432` (NodePort) → `targetPort 5432`.
From the host, connect to `localhost:5432`.

This is a minimal, persistence-free PostgreSQL suitable for development and
testing. It is used as a dependency by higher-level contexts such as
`gravitee.io/apim/oss/aio/postgres`.
