---
title: "APIM - Kafka Gateway"
description: "Gravitee APIM with Kafka Gateway support"
tags: [api-management, gateway, messaging]
---

# APIM Kafka Gateway

Deploys a full Gravitee API Management stack with Kafka Gateway enabled,
allowing the APIM Gateway to act as a Kafka proxy. Clients connect to the
gateway using the Kafka protocol via `*.kafka.sew.local:9092` with TLS.

This is an abstract base context. Concrete variants combine it with a
database-specific APIM OSS context. The default variant is **postgres**.

## Prerequisites

This context uses DNS for host-based routing (`*.kafka.sew.local`). After
creating the cluster, run the one-time OS setup so these hostnames resolve
on your machine (may require `sudo`):

```bash
sew setup dns
```

See the [Networking guide](https://a-cordier.github.io/sew/docs/guides/networking/#local-dns) for details.

## Usage

```bash
# Uses the default variant (postgres)
sew create --from gravitee.io/ee/apim/kafka

# Explicitly select a variant
sew create --from gravitee.io/ee/apim/kafka/postgres
sew create --from gravitee.io/ee/apim/kafka/mongodb
```

## Variants

| Variant    | Database   | Context path                          |
|------------|------------|---------------------------------------|
| `postgres` | PostgreSQL | `gravitee.io/ee/apim/kafka/postgres`  |
| `mongodb`  | MongoDB    | `gravitee.io/ee/apim/kafka/mongodb`   |

Each variant composes the corresponding `gravitee.io/oss/apim/*` context
(which provides the full APIM stack) with the abstract kafka base (which
adds Kafka Gateway configuration on top).

## Context flags

All concrete Kafka variants inherit optional flags from the APIM base context:

| Flag                 | Description                                          |
|----------------------|------------------------------------------------------|
| `--disable-es`       | Disable Elasticsearch and analytics reporters        |
| `--disable-ui`       | Disable both Console and Portal UIs                  |
| `--disable-portal`   | Disable the developer portal UI                      |
| `--enable-hc-vault`  | Deploy HashiCorp Vault and configure it as a secret provider |

```bash
sew create --from gravitee.io/ee/apim/kafka --disable-es
```

Use `sew info` to see the full list of flags and components for a context.

## Endpoints

| Service        | URL                          |
|----------------|------------------------------|
| APIM Console   | http://localhost:30080        |
| APIM Portal    | http://localhost:30081        |
| APIM Gateway   | http://localhost:30082        |
| Management API | http://localhost:30083        |
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

## Dependencies

The abstract base (`gravitee.io/ee/apim/kafka/base`) composes from:

- `kafka/standalone` — single-node Kafka broker in KRaft mode

Each concrete variant additionally composes from:

- `gravitee.io/oss/apim/postgres` or `gravitee.io/oss/apim/mongodb` — full APIM stack with the chosen database backend
