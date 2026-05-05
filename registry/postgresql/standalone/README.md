---
title: "PostgreSQL - Standalone"
description: "Single-node PostgreSQL 17 deployment for Kubernetes"
tags: [database]
---

# PostgreSQL Standalone

Deploys a single-node PostgreSQL 17 instance into a local Kind cluster with
host access on port 30432.

## Usage

```bash
sew create --from postgresql/standalone
```

## Quick Start

Connect from your host:

```bash
PGPASSWORD=postgres psql -h localhost -p 30432 -U postgres -d gravitee
```

| Parameter | Value      |
|-----------|------------|
| Host      | localhost  |
| Port      | 30432      |
| Database  | gravitee   |
| User      | postgres   |
| Password  | postgres   |
