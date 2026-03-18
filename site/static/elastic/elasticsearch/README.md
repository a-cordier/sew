---
description: "Single-node Elasticsearch cluster"
tags: [elasticsearch]
---

# Elasticsearch

Deploys a single-node Elasticsearch cluster using the official Elastic
Helm chart. Security and persistence are disabled for a lightweight development
setup.

## Usage

```bash
sew create elastic/elasticsearch
```

## Details

- **Image:** `docker.elastic.co/elasticsearch/elasticsearch:9.3.1`
- **Helm chart:** `elastic/elasticsearch`
- **Heap:** 256 MB (`-Xmx256m -Xms256m`)
- **Resources:** 250m–1 CPU, 512Mi–1Gi memory
- **Persistence:** disabled

Used as a dependency by Gravitee APIM contexts for analytics and reporting.
