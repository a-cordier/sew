---
title: "MongoDB - Standalone"
description: "Single-node MongoDB 7 deployment for Kubernetes"
tags: [database]
---

# MongoDB Standalone

Deploys a single-node MongoDB 7 instance into a local Kind cluster with
host access on port 30017. Authentication is disabled.

## Usage

```bash
sew create --from mongodb/standalone
```

## Quick Start

Connect from your host:

```bash
mongosh mongodb://localhost:30017
```

| Parameter | Value     |
|-----------|-----------|
| Host      | localhost |
| Port      | 30017     |
| Auth      | none      |
