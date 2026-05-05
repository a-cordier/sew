---
title: "Redis - Standalone"
description: "Single-node Redis 7 deployment for Kubernetes"
tags: [database]
---

# Redis Standalone

Deploys a single-node Redis 7 instance into a local Kind cluster with
host access on port 30379. No authentication is required.

## Usage

```bash
sew create --from redis/standalone
```

## Quick Start

Connect from your host:

```bash
redis-cli -h localhost -p 30379
```

| Parameter | Value     |
|-----------|-----------|
| Host      | localhost |
| Port      | 30379     |
| Auth      | none      |
