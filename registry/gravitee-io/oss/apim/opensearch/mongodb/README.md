---
title: "APIM - OpenSearch MongoDB"
description: "Gravitee APIM with MongoDB backend and OpenSearch"
tags: [networking]
---

# APIM OpenSearch MongoDB

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by MongoDB for persistence and OpenSearch for
analytics.

## Usage

```bash
sew create --from gravitee-io/oss/apim/opensearch/mongodb
```

## Quick Start

Sign in to the Console at [http://localhost:30080](http://localhost:30080)
with the default admin account (`admin` / `admin`).

To create your first API, follow the Gravitee
[APIM quick start guide](https://documentation.gravitee.io/apim/getting-started/quickstart-guide).

## Endpoints

| Service        | URL                   |
|----------------|-----------------------|
| APIM Console   | http://localhost:30080 |
| APIM Portal    | http://localhost:30081 |
| APIM Gateway   | http://localhost:30082 |
| Management API | http://localhost:30083 |
