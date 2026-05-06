---
title: "Jaeger - Standalone"
description: "Jaeger all-in-one distributed tracing backend"
tags: [observability]
---

# Jaeger Standalone

Deploys a Jaeger all-in-one instance into a local Kind cluster with
host access on port 30686 (UI), 30317 (OTLP gRPC), and 30318 (OTLP HTTP).
Uses in-memory storage for lightweight development — traces are lost on
restart.

## Usage

```bash
sew create --from jaeger/standalone
```

## Quick Start

Open the Jaeger UI from your host:

```bash
open http://localhost:30686
```

Send a test trace using the OpenTelemetry endpoint:

```bash
curl -X POST http://localhost:30318/v1/traces \
  -H 'Content-Type: application/json' \
  -d '{"resourceSpans":[]}'
```

| Parameter  | Value                  |
|------------|------------------------|
| UI         | http://localhost:30686  |
| OTLP gRPC  | localhost:30317        |
| OTLP HTTP  | http://localhost:30318 |
| Storage    | in-memory              |
