---
title: "APIM - Kafka MongoDB"
description: "Gravitee APIM with Kafka Gateway and MongoDB backend"
tags: [api-management, gateway, messaging]
---

# APIM Kafka MongoDB

Deploys a full Gravitee API Management stack with Kafka Gateway enabled,
backed by MongoDB for persistence and Elasticsearch for analytics.

## Usage

```bash
sew create --from gravitee.io/ee/apim/kafka/mongodb
```

## Prerequisites

This context uses DNS for host-based routing (`*.kafka.sew.local`). After
creating the cluster, run the one-time OS setup so these hostnames resolve
on your machine (may require `sudo`):

```bash
sew setup dns
```

See the [Networking guide](https://a-cordier.github.io/sew/docs/guides/networking/#local-dns) for details.

## Endpoints

| Service        | URL                            |
|----------------|--------------------------------|
| APIM Console   | http://localhost:30080          |
| APIM Portal    | http://localhost:30081          |
| APIM Gateway   | http://localhost:30082          |
| Management API | http://localhost:30083          |
| Kafka Gateway  | `*.kafka.sew.local:9092` (TLS) |

## Kafka Gateway details

- **Routing mode:** host-based (`*.kafka.sew.local`)
- **TLS:** enabled via a self-signed certificate stored as a Kubernetes Secret
- **Upstream broker:** standalone Kafka at `kafka:9092` (ClusterIP)

## Connecting a Kafka client

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

## Context flags

Optional flags inherited from the APIM base context:

| Flag                 | Description                                          |
|----------------------|------------------------------------------------------|
| `--disable-es`       | Disable Elasticsearch and analytics reporters        |
| `--disable-ui`       | Disable both Console and Portal UIs                  |
| `--disable-portal`   | Disable the developer portal UI                      |
| `--enable-hc-vault`  | Deploy HashiCorp Vault and configure it as a secret provider |
| `--enable-redis`     | Deploy Redis and use it for gateway rate limiting             |

```bash
sew create --from gravitee.io/ee/apim/kafka/mongodb --disable-es
```

Use `sew info` to see the full list of flags and components for this context.

## Dependencies

This context composes from:

- `gravitee.io/oss/apim/mongodb` — full APIM stack with MongoDB backend
- `gravitee.io/ee/apim/kafka/base` — Kafka Gateway configuration

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
