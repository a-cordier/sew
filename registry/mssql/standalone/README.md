---
title: "MSSQL Server"
description: "Single-node Microsoft SQL Server deployment for Kubernetes"
tags: [database]
---

# MSSQL Server

Deploys a single-node Microsoft SQL Server 2022 instance into a local Kind
cluster with host access on port 31433. An init Job creates a `gravitee`
database on first startup.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from mssql/standalone
```

### Cleanup

```bash
sew delete
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
