---
title: "Redis"
description: "Single-node Redis deployment for Kubernetes"
tags: [database]
---

# Redis

Deploys a single-node Redis 7 instance into a local Kind cluster with
host access on port 30379. No authentication is required.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from redis/standalone
```

### Cleanup

```bash
sew delete
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
