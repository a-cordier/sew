---
description: "Gravitee APIM with Kafka Gateway support"
tags: [api-management, kafka]
---

# APIM Kafka Gateway

Deploys a full Gravitee API Management stack with Kafka Gateway enabled,
allowing the APIM Gateway to act as a Kafka proxy. Clients connect to the
gateway using the Kafka protocol via `*.kafka.sew.local:9092` with TLS.

This is an abstract base context. Concrete variants combine it with a
database-specific AIO context. The default variant is **postgres**.

## Prerequisites

The Kafka Gateway uses host-based routing on `*.kafka.sew.local`. Run the
one-time DNS setup so these domains resolve locally:

```bash
sew setup dns
```

## Usage

```bash
# Uses the default variant (postgres)
sew create gravitee.io/apim/ee/kafka

# Explicitly select a variant
sew create gravitee.io/apim/ee/kafka/postgres
sew create gravitee.io/apim/ee/kafka/mongodb
```

## Variants

| Variant    | Database   | Context path                          |
|------------|------------|---------------------------------------|
| `postgres` | PostgreSQL | `gravitee.io/apim/ee/kafka/postgres`  |
| `mongodb`  | MongoDB    | `gravitee.io/apim/ee/kafka/mongodb`   |

Each variant composes the corresponding `gravitee.io/apim/oss/aio/*` context
(which provides the full APIM stack) with the abstract kafka base (which
adds Kafka Gateway configuration on top).

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

This is an Enterprise Edition (EE) context. It requires a valid Gravitee
license key at `$HOME/opt/gravitee/license.key`. If the license file is
missing, the license component is silently skipped (`onMissing: ignore`).

## Dependencies

The abstract base (`gravitee.io/apim/ee/kafka/base`) composes from:

- `kafka/standalone` — single-node Kafka broker in KRaft mode

Each concrete variant additionally composes from:

- `gravitee.io/apim/oss/aio/postgres` or `gravitee.io/apim/oss/aio/mongodb` — full APIM stack with the chosen database backend
