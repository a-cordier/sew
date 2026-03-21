---
title: "MongoDB - Standalone"
description: "Single-node MongoDB 7 deployment for Kubernetes"
tags: [database]
---

# MongoDB Standalone

Deploys a single-replica MongoDB instance as a Kubernetes Deployment with a
NodePort Service on port 27017.

## Usage

```bash
sew create mongodb/standalone
```

## Details

- **Image:** `mongo:7`
- **Port:** 27017
- **Resources:** 250m–1 CPU, 512Mi–1Gi memory

### Host access

Kind maps `hostPort 27017` → `containerPort 30017` (NodePort) → `targetPort 27017`.
From the host, connect to `localhost:27017`.

This is a minimal, persistence-free MongoDB suitable for development and
testing. It is used as a dependency by higher-level contexts such as
`gravitee.io/apim/oss/aio/mongodb`.
