---
title: "APIM - Alert Engine MongoDB"
description: "Gravitee APIM with Alert Engine and MongoDB backend"
tags: [networking]
---

# APIM Alert Engine MongoDB

Deploys a full Gravitee API Management stack (Console, Portal, Gateway, and
Management API) backed by MongoDB for persistence and Elasticsearch for
analytics, with the Gravitee Alert Engine for real-time alerting.

## Usage

```bash
sew create --from gravitee-io/ee/apim/alert-engine/mongodb
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

## License

This is an Enterprise Edition (EE) context. Place your Gravitee license
key at `$HOME/opt/gravitee/license.key` and sew will automatically mount
it into the cluster. If the file is missing, the license component is
silently skipped (`onMissing: ignore`).

To use a different path, override it in your `sew.yaml`:

```yaml
components:
  - name: license
    k8s:
      secrets:
        - name: gravitee-license
          fromFile: '/custom/path/to/license.key'
```
