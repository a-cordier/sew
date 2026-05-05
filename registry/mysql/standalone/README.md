---
title: "MySQL - Standalone"
description: "Single-node MySQL 9 deployment for Kubernetes"
tags: [database]
---

# MySQL Standalone

Deploys a single-node MySQL 9 instance into a local Kind cluster with
host access on port 30306.

## Usage

```bash
sew create --from mysql/standalone
```

## Quick Start

Connect from your host:

```bash
mysql -h 127.0.0.1 -P 30306 -u root -pmysql gravitee
```

| Parameter | Value      |
|-----------|------------|
| Host      | 127.0.0.1  |
| Port      | 30306      |
| Database  | gravitee   |
| User      | root       |
| Password  | mysql      |
