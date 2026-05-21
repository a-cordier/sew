---
title: "Kafka"
description: "Single-node Kafka broker in KRaft combined mode"
tags: [messaging]
---

# Kafka

Deploys a single-node Apache Kafka broker running in KRaft combined mode
(no ZooKeeper) with host access on port 30092.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from kafka/standalone
```

### Cleanup

```bash
sew delete
```

## Quick Start

List topics from your host using [kcat](https://github.com/edenhill/kcat):

```bash
kcat -b localhost:30092 -L
```

Produce and consume a test message:

```bash
echo "hello" | kcat -b localhost:30092 -P -t test-topic
kcat -b localhost:30092 -C -t test-topic -e
```

| Parameter | Value     |
|-----------|-----------|
| Bootstrap | localhost:30092 |
| Protocol  | PLAINTEXT |
