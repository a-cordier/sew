---
title: "Kafka - Standalone"
description: "Single-node Kafka broker in KRaft combined mode"
tags: [messaging, kafka]
---

# Kafka Standalone

Deploys a single-node Apache Kafka broker running in KRaft combined mode
(controller + broker in one process, no ZooKeeper). A NodePort Service exposes
the broker to both in-cluster and host clients.

## Usage

```bash
sew create kafka/standalone
```

## Details

- **Image:** `apache/kafka:latest`
- **Heap:** 256 MB (`-Xmx256m -Xms256m`)
- **Resources:** 250m–1 CPU, 512Mi–1Gi memory
- **Persistence:** disabled

### Listeners

| Listener    | Port | Purpose                                       |
| ----------- | ---- | --------------------------------------------- |
| PLAINTEXT   | 9092 | In-cluster clients connect via `kafka:9092`   |
| CONTROLLER  | 9093 | KRaft quorum (internal only)                  |
| EXTERNAL    | 9094 | Host access via Kind NodePort 30092 → `localhost:9092` |

### Host access

Kind maps `hostPort 9092` → `containerPort 30092` (NodePort) → `targetPort 9094`
(EXTERNAL listener). From the host, connect to `localhost:9092`.

This is a minimal, persistence-free Kafka suitable for development and testing.
