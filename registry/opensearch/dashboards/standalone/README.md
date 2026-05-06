---
title: "OpenSearch Dashboards - Standalone"
description: "OpenSearch Dashboards visualization UI"
tags: [observability]
---

# OpenSearch Dashboards Standalone

Deploys OpenSearch Dashboards alongside a single-node OpenSearch cluster
into a local Kind cluster. Dashboards UI is accessible on port 30601 and
OpenSearch on port 30921. Security is disabled for lightweight development.

## Usage

```bash
sew create --from opensearch/dashboards/standalone
```

## Quick Start

Open the Dashboards UI from your host:

```bash
open http://localhost:30601
```

| Parameter  | Value                   |
|------------|-------------------------|
| Dashboards | http://localhost:30601   |
| OpenSearch | http://localhost:30921   |
| Security   | disabled                |
