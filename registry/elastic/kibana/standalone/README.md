---
title: "Kibana"
description: "Kibana dashboard for Elasticsearch"
tags: [observability]
---

# Kibana

Deploys Kibana alongside a single-node Elasticsearch cluster into a
local Kind cluster. Kibana UI is accessible on port 30601 and
Elasticsearch on port 30920.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from elastic/kibana/standalone
```

### Cleanup

```bash
sew delete
```

## Quick Start

Open the Kibana UI from your host:

```bash
open http://localhost:30601
```

| Parameter     | Value                   |
|---------------|-------------------------|
| Kibana UI     | http://localhost:30601   |
| Elasticsearch | http://localhost:30920   |
| Security      | disabled                |
