---
title: "APIM - Kafka JDBC MySQL"
description: "Gravitee APIM with Kafka Gateway and MySQL JDBC backend"
tags: [networking, messaging]
---

# APIM Kafka JDBC MySQL

Deploys a full Gravitee API Management stack with Kafka Gateway enabled,
backed by MySQL via JDBC for persistence and Elasticsearch for analytics.

## Usage

```bash
sew create --from gravitee-io/ee/apim/kafka/jdbc/mysql
```

## Prerequisites

This context uses DNS for host-based routing (`*.kafka.sew.local`). After
creating the cluster, run the one-time OS setup so these hostnames resolve
on your machine (may require `sudo`):

```bash
sew setup dns
```

See the [Networking guide](https://a-cordier.github.io/sew/docs/guides/networking/#local-dns) for details.

## Quick Start

Sign in to the APIM Console at [http://localhost:30080](http://localhost:30080)
with the default admin account (`admin` / `admin`).

To create your first API, follow the Gravitee
[APIM quick start guide](https://documentation.gravitee.io/apim/getting-started/quickstart-guide).

### Connecting a Kafka client

Extract the TLS certificate from the running cluster:

```bash
kubectl get secret kafka-tls -n gravitee -o jsonpath='{.data.tls\.crt}' | base64 -d > kafka-tls.crt
```

Then configure your Kafka client properties:

```properties
security.protocol=SSL
ssl.truststore.type=PEM
ssl.truststore.location=/path/to/kafka-tls.crt
ssl.endpoint.identification.algorithm=
```

The `ssl.endpoint.identification.algorithm` must be set to empty because the
self-signed certificate covers `*.kafka.sew.local` but broker metadata
addresses use two-level subdomains (e.g. `broker-0-acr.kafka.sew.local`)
that don't match the single-level wildcard.

## Endpoints

| Service        | URL                             |
|----------------|---------------------------------|
| APIM Console   | http://localhost:30080           |
| APIM Portal    | http://localhost:30081           |
| APIM Gateway   | http://localhost:30082           |
| Management API | http://localhost:30083           |
| Kafka Gateway  | `*.kafka.sew.local:9092` (TLS)  |

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
