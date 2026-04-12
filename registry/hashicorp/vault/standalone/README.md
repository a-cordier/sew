---
title: "Vault - Standalone"
description: "Single-node HashiCorp Vault in dev mode"
tags: [secrets]
---

# Vault Standalone

Deploys a single-node HashiCorp Vault server in dev mode using the official
HashiCorp Helm chart. The server starts pre-unsealed with an in-memory backend
and a root token of `root`. The Agent Injector is disabled to keep the
footprint minimal.

## Usage

```bash
sew create --from hashicorp/vault/standalone
```

## Details

- **Image:** `hashicorp/vault:1.21.2`
- **Helm chart:** `hashicorp/vault`
- **Mode:** dev (in-memory, pre-unsealed)
- **Root token:** `root`
- **Resources:** 250m–500m CPU, 256Mi–512Mi memory

### Host access

Kind maps `hostPort 30820` → `containerPort 30820` (NodePort) → `targetPort 8200`.
From the host, connect to `http://localhost:30820`.

```bash
# Check Vault status
curl http://localhost:30820/v1/sys/health

# Authenticate (dev root token)
export VAULT_ADDR=http://localhost:30820
export VAULT_TOKEN=root
vault status
```
