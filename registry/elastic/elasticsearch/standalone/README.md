---
title: "Elasticsearch - Standalone"
description: "Single-node Elasticsearch cluster"
tags: [search]
---

# Elasticsearch Standalone

Deploys a single-node Elasticsearch cluster into a local Kind cluster with
host access on port 30920. Security is disabled for lightweight development.

## Usage

```bash
sew create --from elastic/elasticsearch/standalone
```

## Quick Start

Check the cluster from your host:

```bash
curl http://localhost:30920
curl http://localhost:30920/_cluster/health?pretty
```

| Parameter | Value                    |
|-----------|--------------------------|
| URL       | http://localhost:30920    |
| Security  | disabled                 |
