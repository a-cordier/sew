---
title: "PostgreSQL"
description: "Single-node PostgreSQL deployment for Kubernetes"
tags: [database]
---

# PostgreSQL

Deploys a single-node PostgreSQL 17 instance into a local Kind cluster with
host access on port 30432.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from postgresql/standalone
```

### Cleanup

```bash
sew delete
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
