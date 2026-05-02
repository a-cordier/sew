---
title: "MySQL - Standalone"
description: "Single-node MySQL 9 deployment for Kubernetes"
tags: [database]
---

# MySQL Standalone

Deploys a single-replica MySQL 9 instance as a Kubernetes Deployment with
a NodePort Service on port 3306.

## Usage

```bash
sew create --from mysql/standalone
```

## Details

- **Image:** `mysql:9`
- **Port:** 3306
- **Database:** `gravitee`
- **Credentials:** `root` / `mysql`
- **Resources:** 250m–1 CPU, 256Mi–512Mi memory

### Host access

Kind maps `hostPort 30306` → `containerPort 30306` (NodePort) → `targetPort 3306`.
From the host, connect to `localhost:30306`.

This is a minimal, persistence-free MySQL suitable for development and
testing. It is used as a dependency by higher-level contexts such as
`gravitee.io/oss/apim/jdbc/mysql`.
