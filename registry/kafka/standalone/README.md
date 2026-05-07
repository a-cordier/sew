---
title: "Kafka"
description: "Single-node Kafka broker in KRaft combined mode"
tags: [messaging]
---

# Kafka

Deploys a single-node Apache Kafka broker running in KRaft combined mode
(no ZooKeeper) with host access on port 9092.

## Usage

```bash
sew create --from kafka/standalone
```

## Quick Start

List topics from your host using [kcat](https://github.com/edenhill/kcat):

```bash
kcat -b localhost:9092 -L
```

Produce and consume a test message:

```bash
echo "hello" | kcat -b localhost:9092 -P -t test-topic
kcat -b localhost:9092 -C -t test-topic -e
```

| Parameter | Value     |
|-----------|-----------|
| Bootstrap | localhost:9092 |
| Protocol  | PLAINTEXT |
