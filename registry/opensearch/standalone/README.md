---
title: "OpenSearch"
description: "Single-node OpenSearch cluster"
tags: [search, observability]
---

# OpenSearch

Deploys a single-node OpenSearch cluster into a local Kind cluster with
host access on port 30921. Security is disabled for lightweight development.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from opensearch/standalone
```

### Cleanup

```bash
sew delete
```

## Quick Start

Check the cluster from your host:

```bash
curl http://localhost:30921
curl http://localhost:30921/_cluster/health?pretty
```

| Parameter | Value                    |
|-----------|--------------------------|
| URL       | http://localhost:30921    |
| Security  | disabled                 |
