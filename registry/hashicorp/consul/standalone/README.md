---
title: "Consul"
description: "Single-node HashiCorp Consul in dev mode"
tags: [networking]
---

# Consul

Deploys a single-node HashiCorp Consul server into a local Kind cluster
with host access on port 30500 (UI). Useful for service discovery and
key-value storage in development.

## Usage

```bash
sew create --from hashicorp/consul/standalone
```

## Quick Start

Open the Consul UI from your host:

```bash
open http://localhost:30500
```

| Parameter | Value                   |
|-----------|-------------------------|
| UI        | http://localhost:30500   |
| Auth      | none                    |
