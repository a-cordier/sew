---
title: "OpenSearch - Standalone"
description: "Single-node OpenSearch cluster"
tags: [search]
---

# OpenSearch Standalone

Deploys a single-node OpenSearch cluster into a local Kind cluster with
host access on port 30921. Security is disabled for lightweight development.

## Usage

```bash
sew create --from opensearch/standalone
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
