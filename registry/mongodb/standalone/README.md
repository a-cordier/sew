---
title: "MongoDB"
description: "Single-node MongoDB deployment for Kubernetes"
tags: [database]
---

# MongoDB

Deploys a single-node MongoDB 7 instance into a local Kind cluster with
host access on port 30017. Authentication is disabled.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from mongodb/standalone
```

### Cleanup

```bash
sew delete
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
