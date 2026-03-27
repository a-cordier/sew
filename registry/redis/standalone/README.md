---
title: "Redis - Standalone"
description: "Single-node Redis 7 deployment for Kubernetes"
tags: [cache, database]
---

# Redis Standalone

Deploys a single-replica Redis 7 instance as a Kubernetes Deployment with
a NodePort Service on port 6379.

## Usage

```bash
sew create --from redis/standalone
```

## Details

- **Image:** `redis:7`
- **Port:** 6379
- **Resources:** 100m–250m CPU, 128Mi–256Mi memory

### Host access

Kind maps `hostPort 30379` → `containerPort 30379` (NodePort) → `targetPort 6379`.
From the host, connect to `localhost:30379`.

This is a minimal, persistence-free Redis suitable for development and
testing. It is used as a dependency by higher-level contexts such as
`gravitee.io/ee/edge-stack`.
