---
title: Elasticsearch
layout: detail
path: elastic/elasticsearch
context: true
description: Single-node Elasticsearch cluster
tags:
    - search
components:
    - elasticsearch
notes_create: |-
    Your cluster "es-standalone" is ready.

    Elasticsearch is available at http://localhost:9200 (security disabled).

      curl http://localhost:9200
      curl http://localhost:9200/_cluster/health?pretty
icon: elastic/icon.svg
type: registry
---

Deploys a single-node Elasticsearch cluster using the official Elastic
Helm chart. Security and persistence are disabled for a lightweight development
setup. A NodePort Service exposes the cluster to both in-cluster and host clients.

## Usage

```bash
sew create --from elastic/elasticsearch
```

## Details

- **Image:** `docker.elastic.co/elasticsearch/elasticsearch:9.3.1`
- **Helm chart:** `elastic/elasticsearch`
- **Heap:** 256 MB (`-Xmx256m -Xms256m`)
- **Resources:** 250m–1 CPU, 512Mi–1Gi memory
- **Persistence:** disabled

### Host access

Kind maps `hostPort 9200` → `containerPort 30920` (NodePort) → `targetPort 9200`.
From the host, connect to `localhost:9200`.

Used as a dependency by Gravitee APIM contexts for analytics and reporting.
