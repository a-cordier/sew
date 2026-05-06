---
title: "MSSQL Server - Standalone"
description: "Single-node Microsoft SQL Server 2022 deployment for Kubernetes"
tags: [database]
---

# MSSQL Server Standalone

Deploys a single-node Microsoft SQL Server 2022 instance into a local Kind
cluster with host access on port 31433. An init Job creates a `gravitee`
database on first startup.

## Usage

```bash
sew create --from mssql/standalone
```

## Quick Start

Connect from your host using `sqlcmd`:

```bash
sqlcmd -S localhost,31433 -U SA -P 'Password1!' -C
```

| Parameter | Value       |
|-----------|-------------|
| Host      | localhost   |
| Port      | 31433       |
| Database  | gravitee    |
| User      | SA          |
| Password  | Password1!  |
