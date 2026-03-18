---
description: "Single-node MongoDB 7 deployment for Kubernetes"
tags: [database]
---

# MongoDB Standalone

Deploys a single-replica MongoDB instance as a Kubernetes Deployment with a
ClusterIP Service on port 27017.

## Usage

```bash
sew create mongodb/standalone
```

## Details

- **Image:** `mongo:7`
- **Port:** 27017
- **Resources:** 250m–1 CPU, 512Mi–1Gi memory

This is a minimal, persistence-free MongoDB suitable for development and
testing. It is used as a dependency by higher-level contexts such as
`gravitee.io/apim/aio/mongodb`.
